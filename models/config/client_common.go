package config

import (
	"fmt"
	"github.com/vaughan0/go-ini"
	"strconv"
	"strings"
)

type ClientCommonConf struct {
	ServerAddr      string `json:"server_addr"`
	ServerPort      int    `json:"server_port"`
	LogFile         string `json:"log_file"`
	LogWay          string `json:"log_way"`
	LogLevel        string `json:"log_level"`
	LogMaxDays      int64  `json:"log_max_days"`
	DisableLogColor bool   `json:"disable_log_color"`
	PoolCount       int    `json:"pool_count"`
}

func UnmarshalClientConfFromIni(content string) (cfg ClientCommonConf, err error) {
	cfg = GetDefaultClientConf()
	conf, err := ini.Load(strings.NewReader(content))
	if err != nil {
		return ClientCommonConf{}, fmt.Errorf("Parse ini conf file error:%v", err)
	}
	var (
		tmpStr string
		ok     bool
		v      int64
	)
	if tmpStr, ok = conf.Get("common", "server_addr"); ok {
		cfg.ServerAddr = tmpStr
	}
	if tmpStr, ok = conf.Get("common", "server_port"); ok {
		v, err = strconv.ParseInt(tmpStr, 10, 64)
		if err != nil {
			err = fmt.Errorf("parse conf error:invalid server port")
		}
		cfg.ServerPort = int(v)
	}
	if tmpStr, ok = conf.Get("common", "log_file"); ok {
		cfg.LogFile = tmpStr
	}
	if tmpStr, ok = conf.Get("common", "log_way"); ok {
		cfg.LogWay = tmpStr
	}
	if tmpStr, ok = conf.Get("common", "log_level"); ok {
		cfg.LogLevel = tmpStr
	}
	if tmpStr, ok = conf.Get("common", "log_max_days"); ok {
		if v, err = strconv.ParseInt(tmpStr, 10, 64); err == nil {
			cfg.LogMaxDays = v
		}
	}
	if tmpStr, ok = conf.Get("common", "disable_log_color"); ok && tmpStr == "true" {
		cfg.DisableLogColor = true
	}
	if tmpStr, ok = conf.Get("common", "pool_count"); ok {
		if v, err = strconv.ParseInt(tmpStr, 10, 64); err == nil {
			cfg.PoolCount = int(v)
		}
	}
	return cfg, err
}
func GetDefaultClientConf() ClientCommonConf {
	return ClientCommonConf{
		ServerAddr:      "0.0.0.0",
		ServerPort:      7000,
		LogFile:         "console",
		LogWay:          "console",
		LogLevel:        "info",
		LogMaxDays:      3,
		DisableLogColor: false,
		PoolCount:       0,
	}
}
