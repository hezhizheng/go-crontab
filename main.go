package main

import (
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type CrontabCmdList struct {
	Cmd     string
	Crontab string
}

var ccl []CrontabCmdList
var mutex sync.Mutex

func init() {

	path := "./logs/go.log"
	/* 日志轮转相关函数
	`WithLinkName` 为最新的日志建立软连接
	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	WithMaxAge 和 WithRotationCount二者只能设置一个
	  `WithMaxAge` 设置文件清理前的最长保存时间
	  `WithRotationCount` 设置文件清理前最多保存的个数
	*/
	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	writer, _ := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(180)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(60)*time.Second),
	)
	log.SetOutput(writer)
	//log.SetFormatter(&log.JSONFormatter{})


	viper.SetConfigType("json") // 设置配置文件的类型
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("no such config file")
		} else {
			// Config file was found but another error was produced
			log.Println("read config error")
		}
		log.Fatal(err) // 读取配置文件失败致命错误
	}
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
		fmt.Println("json Marshal error  ", e)
	}

	json.Unmarshal(jsonStr, &ccl)

	for _, v := range ccl {

		Crontab := v.Crontab
		Cmd := v.Cmd

		id, err := c.AddFunc(Crontab, func() {

			p := Cmd
			pa := Explode(p," ")
			CommandName := pa[0]
			mutex.Lock()
			pa = append(pa[:0], pa[0+1:]...)
			mutex.Unlock()
			CommandArg := pa

			f, err := exec.Command(CommandName, CommandArg...).Output()

			if err != nil {
				log.Println(err.Error())
			}
			log.Println(string(f))

		})

		if err != nil {
			log.Println("定时任务启动错误：", err, id, Crontab, Cmd)
		}else{
			log.Println("已启动的定时任务： ", id, Crontab, Cmd)
		}

	}

	c.Start()

	log.Println("Start ing ")

	select {}
}


func Explode(delimiter, text string) []string {
	if len(delimiter) > len(text) {
		return strings.Split(delimiter, text)
	} else {
		return strings.Split(text, delimiter)
	}
}
