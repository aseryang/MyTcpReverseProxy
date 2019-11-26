package main

import "MyTcpReverseProxy/server"
import "MyTcpReverseProxy/utils/log"

type LogConf struct {
	LogFile    string
	LogWay     string
	LogLevel   string
	LogMaxDays int64
}

func GetDefaultLogConf() LogConf {
	return LogConf{
		LogFile:    "console",
		LogWay:     "console",
		LogLevel:   "info",
		LogMaxDays: 30,
	}
}
func main() {
	cfg := GetDefaultLogConf()
	log.InitLog(cfg.LogWay, cfg.LogFile, cfg.LogLevel, cfg.LogMaxDays, false)
	svr := server.NewService()
	svr.Run()
}
