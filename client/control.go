package client

import (
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/hashicorp/yamux"
	"io"
	"net"
)

type Control struct {
	runId   string
	c       net.Conn
	session *yamux.Session
	sendCh  chan (msg.Message)
	readCh  chan (msg.Message)
	closeCh chan struct{}
	vm      *VisitorManager
	pm      *ProxyManager
}

func NewControl(conn net.Conn, session *yamux.Session, runId string, pxyCfgs map[string]config.ProxyConf, visitorCfgs map[string]config.VisitorConf) *Control {
	log.Info("New control runId:%s",runId)
	ctl := Control{runId: runId,
		c:       conn,
		session: session,
		sendCh:  make(chan msg.Message, 100),
		readCh:  make(chan msg.Message, 100),
	}
	ctl.vm = &VisitorManager{ctl: &ctl, visitors: make(map[string]Visitor), visitorCfgs: visitorCfgs}
	ctl.pm = &ProxyManager{ctl: &ctl, pxyCfgs: pxyCfgs, pxys: make(map[string]Proxy)}

	return &ctl
}
func (ctl *Control) Run() {
	go ctl.reader()
	go ctl.writer()
	go ctl.msgHandle()
	go ctl.vm.StartAllVisitors()
	go ctl.pm.StartAllProxies()
	<-ctl.closeCh
}
func (ctl *Control) Close() {
	close(ctl.closeCh)
}

func (ctl *Control) reader() {
	for {
		if m, err := msg.ReadMsg(ctl.c); err != nil {
			if err == io.EOF {
				return
			} else {
				ctl.c.Close()
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
			if err := msg.WriteMsg(ctl.c, m); err != nil {
				return
			}
		}
	}
}

func (ctl *Control) handleReqWorkConn(m msg.Message) {
	workConn, ret := ctl.session.OpenStream()
	if ret != nil {
		return
	}
	m = &msg.NewWorkConn{RunId: ctl.runId}
	if err := msg.WriteMsg(workConn, m); err != nil {
		workConn.Close()
	}
	log.Info("Send NewWorkConn succeed.RunId:%s", ctl.runId)
	var startMsg msg.StartWorkConn
	if err := msg.ReadMsgInto(workConn, &startMsg); err != nil {
		return
	}
	log.Info("Received StartWorkConn msg.")
	go ctl.pm.HandleInWorkConn(workConn, &startMsg)
}

func (ctl *Control) handleNewProxyResp(m msg.Message) {

}
func (ctl *Control)handlePing(){
	ctl.sendCh<-&msg.Pong{}
	log.Info("Send Pong msg.")
}

func (ctl *Control) msgHandle() {
	for {
		rawMsg, ok := <-ctl.readCh
		if !ok {
			return
		}
		switch m := rawMsg.(type) {
		case *msg.ReqWorkConn:
			log.Info("Received msg ReqWorkConn.")
			go ctl.handleReqWorkConn(m)
		case *msg.NewProxyResp:
			log.Info("Received msg NewProxyResp.")
			go ctl.handleNewProxyResp(m)
		case *msg.Ping:
			log.Info("Received msg Ping.")
			go ctl.handlePing()
		}
	}
}
