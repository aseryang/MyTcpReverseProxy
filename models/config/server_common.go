package config

import (
	"fmt"
	"github.com/vaughan0/go-ini"
	"strconv"
	"strings"
)

type ServerCommonConf struct {
	BindAddr        string `json:"bind_addr"`
	BindPort        int    `json:"bind_port"`
	LogFile         string `json:"log_file"`
	LogWay          string `json:"log_way"`
	LogLevel        string `json:"log_level"`
	LogMaxDays      int64  `json:"log_max_days"`
	DisableLogColor bool   `json:"disable_log_color"`
	MaxPoolCount    int    `json:"max_pool_count"`
}

func UnmarshalServerConfFromIni(content string) (cfg ServerCommonConf, err error) {
	cfg = GetDefaultServerConf()
	conf, err := ini.Load(strings.NewReader(content))
	if err != nil {
		return ServerCommonConf{}, fmt.Errorf("Parse ini conf file error:%v", err)
	}
	var (
		tmpStr string
		ok     bool
		v      int64
	)
	if tmpStr, ok = conf.Get("common", "bind_addr"); ok {
		cfg.BindAddr = tmpStr
	}
	if tmpStr, ok = conf.Get("common", "bind_port"); ok {
		v, err = strconv.ParseInt(tmpStr, 10, 64)
		if err != nil {
			err = fmt.Errorf("parse conf error:invalid server port")
		}
		cfg.BindPort = int(v)
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
	if tmpStr, ok = conf.Get("common", "max_pool_count"); ok {
		if v, err = strconv.ParseInt(tmpStr, 10, 64); err == nil {
			cfg.MaxPoolCount = int(v)
		}
	}
	return cfg, err
}

func GetDefaultServerConf() ServerCommonConf {
	return ServerCommonConf{
		BindAddr:        "0.0.0.0",
		BindPort:        7000,
		LogFile:         "console",
		LogWay:          "console",
		LogLevel:        "info",
		LogMaxDays:      3,
		DisableLogColor: false,
		MaxPoolCount:    2,
	}
}
