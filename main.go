package main

import (
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/olekukonko/tablewriter"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type CrontabCmdList struct {
	Cmd     string
	Crontab string
	Model   string
}

type CrontabTaskList struct {
	Id      cron.EntryID
	Cmd     string
	Crontab string
	Model   string
	ErrMsg  interface{}
}

var (
	chanPool = make(chan int, 3)
	wg       sync.WaitGroup
	mutex    sync.Mutex
	ccl      []CrontabCmdList
	tasks    []CrontabTaskList
)

var modelMap = map[string]*cron.Cron{
	"s": cron.New(cron.WithSeconds()),
	"m": cron.New(),
}

const GoCrontabVersion = "v0.1.0"

func init() {
	initLog()
	initConfig()
}

func main() {

	crontabModel := viper.Get(`app.model`)

	// 默认驱动模式
	defaultC := modelMap[crontabModel.(string)]
	c := modelMap[crontabModel.(string)]

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
		if v.Model != "" {
			c = modelMap[v.Model]
		} else {
			v.Model = crontabModel.(string)
		}
		// 添加所有配置的 Crontab
		go addCrontabTask(c, Crontab, Cmd, v.Model)
	}

	// 等待所有任务添加完毕
	wg.Wait()

	close(chanPool)
	defer c.Stop()
	c.Start()
	if defaultC != c {
		defer defaultC.Stop()
		defaultC.Start()
	}

	fmt.Println("go-crontab 程序已启动，请不要关闭终端", "version："+GoCrontabVersion, "power by https://hzz.cool")

	// 表格展示
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"任务ID", "表达式", "执行命令", "错误信息"})

	for _, v := range tasks {
		errMsg := fmt.Sprintf("%s", v.ErrMsg)
		if errMsg == "%!s(<nil>)" {
			errMsg = "nil"
		}
		// 切割一下 字符 表达式 ，避免字符过长终端表格显示变形
		mutex.Lock()
		table.Append([]string{
			v.Model + "-" + fmt.Sprintf("%d", v.Id),
			v.Crontab,
			interceptStrFunc(v.Cmd, 40),
			interceptStrFunc(errMsg, 40),
		})
		mutex.Unlock()
	}
	table.Render() // Send output

	//select {}
	var exit string
	fmt.Printf("按回车键退出\n")
	fmt.Scanln(&exit)
	return
}

func addCrontabTask(c *cron.Cron, Crontab, Cmd string, Model string) {
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

			// 执行时间标记
			startTime := time.Now()
			outputByte, outputErr := exec.Command(execCommandFirst, arg, Cmd).CombinedOutput()
			checkExec(outputErr, Cmd, outputByte, startTime)
		} else {
			if runtime.GOOS == "windows" {
				execCmd(Cmd)
			} else {
				execBash(Cmd)
			}
		}
	})

	mutex.Lock()
	tasks = append(tasks, CrontabTaskList{
		Id:      id,
		Crontab: Crontab,
		Cmd:     Cmd,
		ErrMsg:  err,
		Model:   Model,
	})
	mutex.Unlock()

	//time.Sleep(time.Second)
	<-chanPool
	wg.Done()
}

func initLog() {
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat:   "2006-01-02 15:04:05",
		DisableHTMLEscape: true,
	})

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
