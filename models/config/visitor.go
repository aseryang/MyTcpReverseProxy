package config

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/consts"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/vaughan0/go-ini"
	"reflect"
	"strconv"
)

var (
	visitorConfTypeMap map[string]reflect.Type
)

func init() {
	visitorConfTypeMap = make(map[string]reflect.Type)
	visitorConfTypeMap[consts.StcpProxy] = reflect.TypeOf(StcpVisitorConf{})
}

type VisitorConf interface {
	GetBaseInfo() *BaseVisitorConf
	UnmarshalFromIni(name string, section ini.Section) error
}

func NewVisitorConfByType(cfgType string) VisitorConf {
	v, ok := visitorConfTypeMap[cfgType]
	if !ok {
		return nil
	}
	cfg := reflect.New(v).Interface().(VisitorConf)
	return cfg
}
func NewVisitorConfFromIni(name string, section ini.Section) (cfg VisitorConf, err error) {
	log.Info("Get vistor config from ini...")
	cfgType := section["type"]
	if cfgType == "" {
		err = fmt.Errorf("visitor [%s] type shouldn't be empty", name)
		return
	}
	cfg = NewVisitorConfByType(cfgType)
	if cfg == nil {
		err = fmt.Errorf("visitor [%s] type [%s] error", name, cfgType)
		return
	}
	if err = cfg.UnmarshalFromIni(name, section); err != nil {
		return
	}
	return
}

type StcpVisitorConf struct {
	BaseVisitorConf
}
type BaseVisitorConf struct {
	ProxyName  string `json:"proxy_name"`
	ProxyType  string `json:"proxy_type"`
	Role       string `json:"role"`
	ServerName string `json:"server_name"`
	BindAddr   string `json:"bind_addr"`
	BindPort   int    `json:"bind_port"`
}

func (cfg *BaseVisitorConf) GetBaseInfo() *BaseVisitorConf {
	return cfg
}
func (cfg *BaseVisitorConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	var (
		tmpStr string
		ok     bool
	)
	cfg.ProxyName = name
	cfg.ProxyType = section["type"]
	cfg.Role = section["role"]
	if cfg.Role != "visitor" {
		return fmt.Errorf("Parse conf error:proxy[%s] incorrect role[%s]", name, cfg.Role)
	}
	cfg.ServerName = section["server_name"]
	if cfg.BindAddr = section["bind_addr"]; cfg.BindAddr == "" {
		cfg.BindAddr = "127.0.0.1"
	}
	if tmpStr, ok = section["bind_port"]; ok {
		if cfg.BindPort, err = strconv.Atoi(tmpStr); err != nil {
			return fmt.Errorf("Parse conf error:proxy[%s] bind_prot incorrect", name)
		}
	} else {
		return fmt.Errorf("Parse conf error:proxy[%s] bind_port not found", name)
	}
	log.Info("BaseVisitor config ServerName：%s",cfg.ServerName)
	return nil
}
