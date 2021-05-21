package main

import (
	"encoding/json"
	"fmt"
	"github.com/alexeyco/simpletable"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type CrontabCmdList struct {
	Cmd     string
	Crontab string
}

type CrontabTaskList struct {
	Id      cron.EntryID
	Cmd     string
	Crontab string
	ErrMsg  interface{}
}

var (
	chanPool = make(chan int, 3)
	wg       sync.WaitGroup
	mutex    sync.Mutex
	ccl []CrontabCmdList
	tasks []CrontabTaskList
)

func init() {
	initLog()
	initConfig()
}

func main() {

	crontabModel := viper.Get(`app.model`)

	c := cron.New()

	if crontabModel == "s" {
		c = cron.New(cron.WithSeconds()) // 秒级
	}

	// 从配置文件中读取
	crontabCmdMap := viper.Get(`app.crontab_cmd`)

	jsonStr, e := json.Marshal(crontabCmdMap)

	if e != nil {
		log.Error("json Marshal error  ", e)
	}

	json.Unmarshal(jsonStr, &ccl)

	// 遍历任务
	for _, v := range ccl {
		wg.Add(1)
		Crontab := v.Crontab
		Cmd := v.Cmd
		// 添加所有配置的 Crontab
		go addCrontabTask(c, Crontab, Cmd)
	}

	// 等待所有任务添加完毕
	wg.Wait()

	close(chanPool)
	defer c.Stop()
	c.Start()

	fmt.Println("程序已启动，请不要关闭终端")

	// 表格展示
	table := simpletable.New()

	table.Header = &simpletable.Header{
		Cells: []*simpletable.Cell{
			{Align: simpletable.AlignCenter, Text: "任务ID"},
			{Align: simpletable.AlignCenter, Text: "表达式"},
			{Align: simpletable.AlignCenter, Text: "执行命令"},
			{Align: simpletable.AlignCenter, Text: "错误信息"},
		},
	}

	for _, row := range tasks {
		errMsg := fmt.Sprintf("%s", row.ErrMsg)
		if errMsg == "%!s(<nil>)" {
			errMsg = "nil"
		}
		r := []*simpletable.Cell{
			{Align: simpletable.AlignRight, Text: fmt.Sprintf("%d", row.Id)},
			{Align:simpletable.AlignCenter,Text: row.Crontab},
			{Align:simpletable.AlignCenter,Text: row.Cmd},
			{Align:simpletable.AlignCenter,Text: errMsg},
		}
		mutex.Lock()
		table.Body.Cells = append(table.Body.Cells, r)
		mutex.Unlock()
	}

	table.SetStyle(simpletable.StyleRounded)
	fmt.Println(table.String())

	//select {}
	var exit string
	fmt.Printf("输入任意键退出\n")
	fmt.Scanln(&exit)
	return
}

func addCrontabTask(c *cron.Cron, Crontab , Cmd string) {
	chanPool <- 1
	id, err := c.AddFunc(Crontab, func() {

		execMode := viper.Get(`app.exec_mode`)

		execCommandFirst := ""

		if execMode != "" {
			execCommandFirst = execMode.(string)
			arg := ""
			if execCommandFirst == "cmd" {
				arg = "/c"
			}

			if execCommandFirst == "bash" {
				arg = "-c"
			}

			if arg == "" {
				panic("只支持定义 bash 或 cmd 命令执行！")
			}

			outputByte, outputErr := exec.Command(execCommandFirst,arg,Cmd).CombinedOutput()
			checkExec(outputErr, Cmd, outputByte)
		}else{
			if runtime.GOOS == "windows" {
				execCmd(Cmd)
			} else {
				execBash(Cmd)
			}
		}
	})


	mutex.Lock()
	tasks = append(tasks,CrontabTaskList{
		Id: id,
		Crontab: Crontab,
		Cmd: Cmd,
		ErrMsg: err,
	})
	mutex.Unlock()

	//time.Sleep(time.Second)
	<-chanPool
	wg.Done()
}

func execBash(Cmd string,)  {
	outputByte, outputErr := exec.Command("bash", "-c", Cmd).CombinedOutput()
	checkExec(outputErr, Cmd, outputByte)
}

func execCmd(Cmd string,)  {
	outputByte, outputErr := exec.Command("cmd","/c",  Cmd).CombinedOutput()
	checkExec(outputErr, Cmd, outputByte)
}

// 检测bash 、cmd 的运行环境
func checkExec(outputErr error, Cmd string, outputByte []byte) {
	if outputErr != nil {
		// executable file not found
		if strings.Contains(outputErr.Error(), "executable file not found") {
			panic("请确认当前系统支持 bash 或 cmd 命令的执行环境，并且已添加至环境变量。错误：" + outputErr.Error())
		}
		log.Error(outputErr.Error())
	}
	log.Println("执行命令：", Cmd, "输出：", string(outputByte))
}

func initLog() {
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})

	path := "./logs/"
	/* 日志轮转相关函数
	`WithLinkName` 为最新的日志建立软连接
	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	WithMaxAge 和 WithRotationCount二者只能设置一个
	  `WithMaxAge 设置文件清理前的最长保存时间`
	  `WithRotationCount` 设置文件清理前最多保存的个数
	*/
	// 下面配置日志每隔 1天 转一个新文件，保留最近 1周 的日志文件，多余的自动清理掉。
	LinkName := path + "go-crontab.log"

	writer, _ := rotatelogs.New(
		//path+".%Y%m%d%H%M",
		path+"go-crontab-%Y-%m-%d.log",
		rotatelogs.WithLinkName(LinkName),
		rotatelogs.WithMaxAge(time.Duration(604800)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(86400)*time.Second),
	)
	log.SetOutput(writer)
}

func initConfig() {
	viper.SetConfigType("json") // 设置配置文件的类型
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			panic("no such config file 当前目录没有config.json文件")
		} else {
			// Config file was found but another error was produced
			panic("read config error 读取配置文件错误")
		}
	}
}