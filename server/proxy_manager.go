package server

import (
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"sync"
)

type ProxyManager struct {
	mtx  sync.Mutex
	pxys map[string]Proxy
}

func (pm *ProxyManager) RegisterProxy(pxymsg msg.NewProxy, ctl *Control) {
	if _,ok:=pm.pxys[pxymsg.ProxyName];ok{
		log.Info("The proxyName:%s is exist.so do nothing.",pxymsg.ProxyName)
		return
	}
	pxy := NewProxy(ctl, pxymsg)
	err := pxy.Run()
	if err != nil {
		return
	}
	pm.mtx.Lock()
	pm.pxys[pxymsg.ProxyName] = pxy
	ctl.pxys = append(ctl.pxys, pxymsg.ProxyName)
	pm.mtx.Unlock()
}
func (pm *ProxyManager) StopProxy(pxymsg msg.CloseProxy) {
	pxy, ok := pm.pxys[pxymsg.ProxyName]
	if !ok {
		return
	}
	pxy.Stop()
	pm.mtx.Lock()
	delete(pm.pxys, pxymsg.ProxyName)
	pm.mtx.Unlock()
}
func (pm *ProxyManager) StopAllProxy() {
	pm.mtx.Lock()
	for k, v := range pm.pxys {
		v.Stop()
		delete(pm.pxys, k)
	}
	pm.mtx.Unlock()
}
func (pm *ProxyManager) StopProxies(pxys []string) {
	pm.mtx.Lock()
	for _, pxyName := range pxys {
		pxy, ok := pm.pxys[pxyName]
		if !ok {
			continue
		} else {
			log.Info("Stop proxy:%s",pxyName)
			pxy.Stop()
			delete(pm.pxys, pxyName)
		}
	}
	pm.mtx.Unlock()
}
