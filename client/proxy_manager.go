package client

import (
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/hashicorp/yamux"
)

type ProxyManager struct {
	ctl     *Control
	pxyCfgs map[string]config.ProxyConf
	pxys    map[string]Proxy
}

func (pm *ProxyManager) StartAllProxies() {
	for name, cfg := range pm.pxyCfgs {
		pxy := NewProxy(pm.ctl, cfg)
		pxy.Run()
		pm.pxys[name] = pxy
	}
}
func (pm *ProxyManager) StopAllProxies() {
	for name, pxy := range pm.pxys {
		pxy.Close()
		delete(pm.pxys, name)
	}
}
func (pm *ProxyManager) HandleInWorkConn(conn *yamux.Stream, msg *msg.StartWorkConn) {
	pxy, ok := pm.pxys[msg.ProxyName]
	if ok {
		pxy.InWorkConn(conn, msg)
	}
}
