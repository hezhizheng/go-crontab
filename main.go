package main

import (
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os/exec"
	"time"
)

type CrontabCmdList struct {
	Cmd     string
	Crontab string
}

var ccl []CrontabCmdList

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

		Crontab := v.Crontab
		Cmd := v.Cmd

		// 添加所有配置的 Crontab
		id, err := c.AddFunc(Crontab, func() {

			f, err := exec.Command("bash", "-c", Cmd).Output()

			if err != nil {
				log.Error(err.Error())
			}
			log.Println("执行命令：", Cmd, "输出：", string(f))

		})

		if err != nil {
			fmt.Println("定时任务启动错误：", err, id, Crontab, Cmd)
		} else {
			fmt.Println("已启动监听的定时任务： ", id, "表达式：", Crontab, "命令：", Cmd)
		}

	}

	c.Start()

	fmt.Println("Start ing ")

	select {}
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
			log.Error("no such config file")
		} else {
			// Config file was found but another error was produced
			log.Error("read config error")
		}
		log.Fatal(err) // 读取配置文件失败致命错误
	}
}
