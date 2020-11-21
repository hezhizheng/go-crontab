## Go-Crontab

> 基于 golang 的 crontab 定时任务管理器

### features
- 支持分钟跟秒级
- 内部调用 bash -c 命令
- 理论上跨平台支持 Windows 、Linux、MacOs
- 自用test（windows下的确没啥好用的定时任务管理器。。。还不如自己搞一个）

### 使用

自定义json配置文件
```
# 参数说明
{
  # 支持定义多个定时任务，直接在 crontab_cmd 追加对象即可
  "app": {
    "model": "s", # 默认 s 秒级 如需使用分钟级改为 m
    "crontab_cmd": [
      {
        "crontab": "0/1 * * * * ?", #crontab 表达式
        "cmd": "go version" # 要执行的命令 
      }
    ]
  }
}

# windows 实现 laravel 的任务调度 schedule:run (只支持分钟级别，不能定义秒级的！！！)
# app.model可以定义为s 但是对应的crontab表达式必须为每分钟运行！可参考 Java(Quartz) 表达式书写
# 或直接定义app.model 为 m ,则crontab表达式为 "* * * * *"
{
  "app": {
    "model": "s",
    "crontab_cmd": [
      {
        "crontab": "0 0/1 * * * ?",
        "cmd": "go version"
      },
      {
        "crontab": "0 0/1 * * * ?",
        "cmd": "cd /e/www/project/dexter/laravel-test-demo && php artisan schedule:run"
      }
    ]
  }
}


```

编译 (windows提供编译好的文件下载 [releases](https://github.com/hezhizheng/go-crontab/releases) )
```
go build
```


运行
- 保证编译的文件与 config.json 在同级目录

- 执行 ./go-crontab.exe (不要关闭终端)
![free-pic](https://i.loli.net/2020/11/21/BSqXohbL4NnpmU1.png)

- 执行过程会自动生成log文件(保存一周，会定期清理)

