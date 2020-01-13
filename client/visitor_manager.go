package client

import (
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
)

type VisitorManager struct {
	ctl         *Control
	visitors    map[string]Visitor
	visitorCfgs map[string]config.VisitorConf
}

func (vm *VisitorManager) StartAllVisitors() {
	 log.Info("Start all visitors...")
	for name, cfg := range vm.visitorCfgs {
		log.Info("Start visitor:%s",cfg.GetBaseInfo().ServerName)
		v := NewVisitor(vm.ctl, cfg)
		if v != nil {
			v.Run()
			vm.visitors[name] = v
		}
	}
}
func (vm *VisitorManager) StopAllVisitors() {
	for _, v := range vm.visitors {
		v.Close()
	}
}
