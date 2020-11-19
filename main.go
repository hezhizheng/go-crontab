package main

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type CrontabCmdList struct {
	Cmd     string
	Crontab string
}

var ccl []CrontabCmdList
var mutex sync.Mutex

func init() {
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
	// cron.New(cron.WithSeconds()) // 秒级
	c := cron.New()

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

		log.Println("定时任务启动", err, id, v.Crontab, v.Cmd)
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
