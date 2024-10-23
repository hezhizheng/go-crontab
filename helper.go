package main

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/simplifiedchinese"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

// 处理终端中文显示乱码
func convertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}
	return str
}

// 按需切割、拼接字符串
func interceptStrFunc(str string, num int) string {
	if num <= 0 {
		return str
	}
	strLen := utf8.RuneCountInString(str)
	if strLen <= num {
		return str
	}
	// 换行符
	symbol := "\n"
	// 向上取整
	float64Num := float64(num)
	CeilNum := math.Ceil(float64(strLen) / float64Num)
	intC := int(CeilNum)

	// 初始值
	s := 0
	_num := num

	var builder strings.Builder
	for j := 1; j <= intC; j++ {
		if j == intC {
			num = strLen
		}
		builder.WriteString(string([]rune(str)[s:num]))
		builder.WriteString(symbol)
		s = s + _num
		num = num + _num
	}
	return builder.String()
}

func execBash(Cmd string) {
	// 执行时间标记
	startTime := time.Now()
	outputByte, outputErr := exec.Command("bash", "-c", Cmd).CombinedOutput()
	checkExec(outputErr, Cmd, outputByte, startTime)
}

func execCmd(Cmd string) {
	startTime := time.Now()
	outputByte, outputErr := exec.Command("cmd", "/c", Cmd).CombinedOutput()
	checkExec(outputErr, Cmd, outputByte, startTime)
}

// 检测bash 、cmd 的运行环境
func checkExec(outputErr error, Cmd string, outputByte []byte, startTime time.Time) {
	if outputErr != nil {
		// executable file not found
		if strings.Contains(outputErr.Error(), "executable file not found") {
			panic("请确认当前系统支持 bash 或 cmd 命令的执行环境，并且已添加至环境变量。错误：" + outputErr.Error())
		}
		log.Error(outputErr.Error())
	}
	// 结束时间标记
	endTime := time.Since(startTime)
	ExecSecondsS := strconv.FormatFloat(endTime.Seconds(), 'f', 2, 64)

	log.Println("执行命令：", Cmd, "输出：", convertByte2String(outputByte, "GB18030"), "执行耗时：", ExecSecondsS+" s")
}
