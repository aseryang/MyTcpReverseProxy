package server

import (
	network "github.com/aseryang/MyTcpReverseProxy/utils/net"
	"sync"
)

type VisitorManager struct {
	listeners map[string]*network.CustomListener
	mtx sync.Mutex
}

func (vm *VisitorManager) NewVisitorListener(pxyName string) *network.CustomListener {
	vm.mtx.Lock()
	ln, ok := vm.listeners[pxyName]
	if ok {
		ln.Close()
		delete(vm.listeners, pxyName)
	}
	ln = network.NewCustomListener()
	vm.listeners[pxyName] = ln
	vm.mtx.Unlock()
	return ln
}
func (vm *VisitorManager) GetListener(pxyName string) *network.CustomListener {
	vm.mtx.Lock()
	ln, ok := vm.listeners[pxyName]
	vm.mtx.Unlock()
	if ok {
		return ln
	}
	return nil
}
func (vm *VisitorManager) StopAllListener() {
	vm.mtx.Lock()
	for k, v := range vm.listeners {
		v.Close()
		delete(vm.listeners, k)
	}
	vm.mtx.Unlock()
}
func (vm*VisitorManager)StopListeners(lns []string){
	vm.mtx.Lock()
	for _, name := range lns {
		ln, ok := vm.listeners[name]
		if !ok {
			continue
		} else {
			ln.Close()
			delete(vm.listeners, name)
		}
	}
	vm.mtx.Unlock()
}
