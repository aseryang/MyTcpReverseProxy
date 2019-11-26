package main

import (
	"MyTcpReverseProxy/utils/log"
	"MyTcpReverseProxy/utils/net"
	"fmt"
	"testing"
	"time"
)

//type LogConf struct {
//	LogFile    string
//	LogWay     string
//	LogLevel   string
//	LogMaxDays int64
//}
//
//func GetDefaultLogConf() LogConf {
//	return LogConf{
//		LogFile:    "console",
//		LogWay:     "console",
//		LogLevel:   "info",
//		LogMaxDays: 30,
//	}
//}
//func TestLog(t *testing.T) {
//	cfg := GetDefaultLogConf()
//	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel, cfg.LogMaxDays, false)
//	log.Info("haha test")
//}
func TestTcpAPI(t *testing.T) {
	c, err := net.ConnectTcpServer("106.12.17.239:22")
	if err != nil {
		fmt.Println(err)
	}
	buf := make([]byte, 10)
	for {
		c.Read(buf)
		fmt.Println(string(buf))
	}
}

func TestGoRuting(t *testing.T) {
	go func() {
		go func() {
			defer fmt.Println("inner func exit")
			var i = 0
			for {
				fmt.Println("i = ", i)
				i++
				time.Sleep(time.Second * 1)
			}
		}()
		defer fmt.Println("main func exit")
	}()
}
