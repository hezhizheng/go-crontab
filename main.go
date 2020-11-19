package main

import (
	"encoding/json"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"log"
)

type CrontabCmdList struct {
	Cmd string
	Crontab string
}

var ccl []CrontabCmdList

func init()  {
	viper.SetConfigType("yaml") // 设置配置文件的类型
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
		fmt.Println("333  ",e)
	}

	fmt.Println("v ",string(jsonStr),crontabCmdMap.([]interface{})[0])

	json.Unmarshal(jsonStr, &ccl)

	fmt.Println("v2 ",crontabCmdMap,ccl,string(jsonStr))


	c.AddFunc("30 * * * *", func() {
		fmt.Println("Every hour on the half hour")
	})

	c.Start()

	log.Println("sdfsdfdfdf")
}
