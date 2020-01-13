package server

import (
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/aseryang/MyTcpReverseProxy/utils/version"
	"io"
	"net"
	"sync"
	"time"
)

type Control struct {
	svr           *Service
	sendCh        chan msg.Message
	readCh        chan msg.Message
	heartbeatCh   chan msg.Message
	conn          net.Conn
	workConnCh    chan net.Conn
	connPoolCount int
	closeCh       chan struct{}
	RunId         string
	pm            *ProxyManager
	vm            *VisitorManager
	pxys          []string
	listeners     []string
	mtx           sync.Mutex
}

func NewControl(svr *Service, loginMsg *msg.Login, pm *ProxyManager, vm *VisitorManager) *Control {
	var poolCount = 2
	ctl := Control{svr: svr,
		sendCh:        make(chan msg.Message, 10),
		readCh:        make(chan msg.Message, 10),
		heartbeatCh:   make(chan msg.Message, 10),
		connPoolCount: poolCount,
		closeCh:       make(chan struct{}),
		workConnCh:    make(chan net.Conn, poolCount),
		RunId:         loginMsg.RunId,
		pm:            pm,
		vm:            vm}
	return &ctl
}
func (ctl *Control) Run() {
	log.Info("Send LoginResp msg, RunId:%s", ctl.RunId)
	loginRespMsg := &msg.LoginResp{Version: version.Full(), RunId: ctl.RunId}
	msg.WriteMsg(ctl.conn, loginRespMsg)
	go ctl.reader()
	go ctl.writer()
	for i := 0; i < ctl.connPoolCount; i++ {
		ctl.sendCh <- &msg.ReqWorkConn{}
	}
	go ctl.msgHandle()
	go ctl.heartBeat()
	<-ctl.closeCh
}
func (ctl *Control) RegisterWorkConn(workConn net.Conn) {
	ctl.workConnCh <- workConn
}
func (ctl *Control) GetWorkConn() (workConn net.Conn, err error) {
	var ok bool
	select {
	case workConn, ok = <-ctl.workConnCh:
		if !ok {
			log.Info("Get a work connection failed.")
			return
		}
	default:
		log.Info("Not enough,request a new work connection.")
		ctl.sendCh <- &msg.ReqWorkConn{}
		select {
		case workConn, ok = <-ctl.workConnCh:
			if !ok {
				log.Info("Get a work connection failed.")
				return
			}
		case <-time.After(time.Second * 2):
			log.Info("Timeout trying to get work connection.")
			return
		}
	}
	log.Info("Already token a work connection,request a new work connection")
	ctl.sendCh <- &msg.ReqWorkConn{}
	return
}
func (ctl *Control) stopper() {
	close(ctl.closeCh)
}
func (ctl *Control) reader() {
	for {
		if m, err := msg.ReadMsg(ctl.conn); err != nil {
			if err == io.EOF {
				return
			} else {
				ctl.conn.Close()
			}
		} else {
			ctl.readCh <- m
		}
	}
}
func (ctl *Control) writer() {
	for {
		if m, ok := <-ctl.sendCh; !ok {
			return
		} else {
			if err := msg.WriteMsg(ctl.conn, m); err != nil {
				return
			}
		}
	}
}
func (ctl *Control) msgHandle() {
	for {
		rawMsg, ok := <-ctl.readCh
		if !ok {
			return
		}
		switch m := rawMsg.(type) {
		case *msg.NewProxy:
			log.Info("Received NewProxy msg.")
			go ctl.pm.RegisterProxy(*m, ctl)
		case *msg.CloseProxy:
			log.Info("Received CloseProxy msg.")
			go ctl.pm.StopProxy(*m)
		case *msg.Pong:
			log.Info("Received Pong msg.")
			ctl.heartbeatCh <- m
		}
	}
}
func (ctl *Control) close() {
	ctl.conn.Close()
	close(ctl.sendCh)
	close(ctl.readCh)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for ; ; {
			select {
			case conn:=<-ctl.workConnCh:
				log.Info("Close work connection.")
				conn.Close()
			default:
				close(ctl.workConnCh)
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
	log.Info("WorkConn channel is closed.")
	ctl.pm.StopProxies(ctl.pxys)
	log.Info("Proxies stopped.")
	ctl.vm.StopListeners(ctl.listeners)
	log.Info("Visitor listeners stopped.")
	close(ctl.closeCh)
	log.Info("Control:%s is closed.", ctl.RunId)
}
func (ctl *Control) heartBeat() {
	timeOutCount := 0
	maxTimeOutCount := 3
	go func() {
		for ; ; {
			select {
			case <-ctl.heartbeatCh:
				ctl.mtx.Lock()
				timeOutCount = 0
				ctl.mtx.Unlock()
			case <-ctl.closeCh:
				return
			}
		}
	}()
	for ; ; {
		select {
		case <-time.After(time.Second * 1):
			if timeOutCount > maxTimeOutCount {
				log.Info("Heartbeat reached max timeout times, close the control.")
				ctl.close()
				ctl.svr.UnRegisterControl(ctl.RunId)
				return
			}
			ctl.sendCh <- &msg.Ping{}
			ctl.mtx.Lock()
			timeOutCount += 1
			ctl.mtx.Unlock()
		case <-ctl.closeCh:
			return
		}
	}
}
