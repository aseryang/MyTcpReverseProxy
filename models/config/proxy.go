package config

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/consts"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/vaughan0/go-ini"
	"reflect"
	"strconv"
	"strings"
)

var (
	proxyConfTypeMap map[string]reflect.Type
)

func init() {
	proxyConfTypeMap = make(map[string]reflect.Type)
	proxyConfTypeMap[consts.TcpProxy] = reflect.TypeOf(TcpProxyConf{})
	proxyConfTypeMap[consts.StcpProxy] = reflect.TypeOf(StcpProxyConf{})
	proxyConfTypeMap[consts.UdpProxy] = reflect.TypeOf(UdpProxyConf{})
}
func NewConfByType(proxyType string) ProxyConf {
	v, ok := proxyConfTypeMap[proxyType]
	if !ok {
		return nil
	}
	cfg := reflect.New(v).Interface().(ProxyConf)
	return cfg
}

type ProxyConf interface {
	GetBaseInfo() *BaseProxyConf
	UnmarshalFromIni(name string, section ini.Section) error
	MarshalToMsg(pMsg *msg.NewProxy)
}
type BaseProxyConf struct {
	ProxyName string `json:"proxy_name"`
	ProxyType string `json:"proxy_type"`
	LocalSvrConf
}

func (cfg *BaseProxyConf) MarshalToMsg(pMsg *msg.NewProxy) {
	pMsg.ProxyName = cfg.ProxyName
	pMsg.ProxyType = cfg.ProxyType
}
func (cfg *BaseProxyConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	cfg.ProxyName = name
	cfg.ProxyType = section["type"]
	if err = cfg.LocalSvrConf.UnmarshalFromIni(name, section); err != nil {
		return err
	}
	return
}

type LocalSvrConf struct {
	LocalIp   string `json:"local_ip"`
	LocalPort int    `json:"local_port"`
}

func (cfg *LocalSvrConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	if cfg.LocalIp = section["local_ip"]; cfg.LocalIp == "" {
		cfg.LocalIp = "127.0.0.1"
	}
	if tmpStr, ok := section["local_port"]; ok {
		if cfg.LocalPort, err = strconv.Atoi(tmpStr); err != nil {
			return fmt.Errorf("Parse conf error: proxy [%s] local port error", name)
		}
	} else {
		return fmt.Errorf("Parse conf error: proxy [%s] local port not found", name)
	}
	return
}

type TcpProxyConf struct {
	BaseProxyConf
	BindInfoConf
}

func (cfg *TcpProxyConf) GetBaseInfo() *BaseProxyConf {
	return &cfg.BaseProxyConf
}
func (cfg *TcpProxyConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	if err = cfg.BaseProxyConf.UnmarshalFromIni(name, section); err != nil {
		return
	}
	if err = cfg.BindInfoConf.UnmarshalFromIni(name, section); err != nil {
		return
	}
	return
}
func (cfg *TcpProxyConf) MarshalToMsg(pMsg *msg.NewProxy) {
	cfg.BaseProxyConf.MarshalToMsg(pMsg)
	cfg.BindInfoConf.MarshalToMsg(pMsg)
}

type BindInfoConf struct {
	RemotePort int `json:"remote_port"`
}

func (cfg *BindInfoConf) MarshalToMsg(pMsg *msg.NewProxy) {
	pMsg.RemotePort = cfg.RemotePort
}

func (cfg *BindInfoConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	if tmpStr, ok := section["remote_port"]; ok {
		if cfg.RemotePort, err = strconv.Atoi(tmpStr); err != nil {
			log.Error("Parse conf error: proxy [%s] remote port error", name)
			return
		}
	} else {
		log.Error("Parse conf error: proxy [%s] remote port not found", name)
		return
	}
	return
}

type UdpProxyConf struct {
	BaseProxyConf
	BindInfoConf
}

func (cfg *UdpProxyConf) MarshalToMsg(pMsg *msg.NewProxy) {
	cfg.BaseProxyConf.MarshalToMsg(pMsg)
	cfg.BindInfoConf.MarshalToMsg(pMsg)
}
func (cfg *UdpProxyConf) GetBaseInfo() *BaseProxyConf {
	return &cfg.BaseProxyConf
}
func (cfg *UdpProxyConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	if err = cfg.BaseProxyConf.UnmarshalFromIni(name, section); err != nil {
		return
	}
	if err = cfg.BindInfoConf.UnmarshalFromIni(name, section); err != nil {
		return
	}
	return
}

type StcpProxyConf struct {
	BaseProxyConf
	Role string `json:"role"`
}

func (cfg *StcpProxyConf) GetBaseInfo() *BaseProxyConf {
	return &cfg.BaseProxyConf
}
func (cfg *StcpProxyConf) UnmarshalFromIni(name string, section ini.Section) (err error) {
	cfg.Role = section["role"]
	if err = cfg.BaseProxyConf.UnmarshalFromIni(name, section); err != nil {
		return
	}
	return
}

func LoadAllConfFromIni(content string) (proxyConfs map[string]ProxyConf, visitorConfs map[string]VisitorConf, err error) {
	log.Info("Get proxy config and visitor config from ini...")
	conf, errRet := ini.Load(strings.NewReader(content))
	if errRet != nil {
		err = errRet
		return
	}
	proxyConfs = make(map[string]ProxyConf)
	visitorConfs = make(map[string]VisitorConf)
	for name, section := range conf {
		log.Info("section name is :%s", name)
		if name == "common" {
			continue
		}
		log.Info("role is :%s",section["role"])
		if section["role"] == "" {
			section["role"] = "server"
		}
		role := section["role"]
		if role == "server" {
			log.Info("Get proxy config...")
			cfg, errRet := NewProxyConfFromIni(name, section)
			if errRet != nil {
				err = errRet
				return
			}
			proxyConfs[name] = cfg
		} else if role == "visitor" {
			log.Info("Get visitor config...")
			cfg, errRet := NewVisitorConfFromIni(name, section)
			if errRet != nil {
				err = errRet
				return
			}
			visitorConfs[name] = cfg
		}
	}
	return
}
func NewProxyConfFromIni(name string, section ini.Section) (cfg ProxyConf, err error) {
	proxyType := section["type"]
	if proxyType == "" {
		proxyType = consts.TcpProxy
		section["type"] = consts.TcpProxy
	}
	cfg = NewConfByType(proxyType)
	if cfg == nil {
		err = fmt.Errorf("proxy [%s] type [%s] error", name, proxyType)
		return
	}
	if err = cfg.UnmarshalFromIni(name, section); err != nil {
		return
	}
	return
}
