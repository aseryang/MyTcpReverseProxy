package server

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	network "github.com/aseryang/MyTcpReverseProxy/utils/net"
	"github.com/aseryang/MyTcpReverseProxy/utils/util"
	"github.com/hashicorp/yamux"
	"net"
	"sync"
)

type Service struct {
	ln          net.Listener
	ctlsByRunId map[string]*Control
	pm          *ProxyManager
	vm          *VisitorManager
	mtx         sync.Mutex
}

func NewService(cfg config.ServerCommonConf) (err error, svr *Service) {
	svr = &Service{ctlsByRunId: make(map[string]*Control),
		pm: &ProxyManager{pxys: make(map[string]Proxy)},
		vm: &VisitorManager{listeners: make(map[string]*network.CustomListener)}}
	addr := fmt.Sprintf("%s:%d", cfg.BindAddr, cfg.BindPort)
	svr.ln, err = net.Listen("tcp", addr)
	if err != nil {
		log.Error("tcp bind %s error:%s", addr, err)
		return
	}
	log.Info("Service listen on %s succeed.", addr)
	return
}
func (svr *Service) RegisterControl(conn net.Conn, msg *msg.Login) (err error) {
	if msg.RunId == "" {
		msg.RunId, err = util.RandId()
		if err != nil {
			return
		}
	}
	log.Info("Register new control, RunId is : %s", msg.RunId)
	svr.mtx.Lock()
	svr.ctlsByRunId[msg.RunId] = NewControl(svr, msg, svr.pm, svr.vm)
	svr.ctlsByRunId[msg.RunId].conn = conn
	svr.mtx.Unlock()
	go svr.ctlsByRunId[msg.RunId].Run()
	return
}
func (svr *Service) UnRegisterControl(runId string) {
	svr.mtx.Lock()
	delete(svr.ctlsByRunId, runId)
	svr.mtx.Unlock()
}
func (svr *Service) RegisterWorkConn(conn net.Conn, msg *msg.NewWorkConn) {
	svr.mtx.Lock()
	if ctl, ok := svr.ctlsByRunId[msg.RunId]; ok {
		ctl.RegisterWorkConn(conn)
	} else {
		log.Error("Register work conneciton failed.can't find control runId:%s", msg.RunId)
	}
	svr.mtx.Unlock()
}
func (svr *Service) handleNewVisitor(conn net.Conn, visitorMsg *msg.NewVisitorConn) {
	ln := svr.vm.GetListener(visitorMsg.ProxyName)
	if ln == nil {
		log.Info("can't find visitor proxy name:%s", visitorMsg.ProxyName)
		return
	}
	ln.PutConn(conn)
	newVisitorConnRespMsg := &msg.NewVisitorConnResp{ProxyName: visitorMsg.ProxyName}
	err := msg.WriteMsg(conn, newVisitorConnRespMsg)
	if err != nil {
		log.Info("Get newVisitorConnRespMsg error:%s", err)
		return
	}
}
func (svr *Service) Run() {
	for {
		c, err := svr.ln.Accept()
		if err != nil {
			return
		}
		go func(conn net.Conn) {
			session, err := yamux.Server(conn, nil)
			if err != nil {
				return
			}
			for {
				stream, err := session.AcceptStream()
				if err != nil {
					return
				}
				var rawMsg msg.Message
				if rawMsg, err = msg.ReadMsg(stream); err != nil {
					return
				}
				switch m := rawMsg.(type) {
				case *msg.Login:
					log.Info("Received Login msg.")
					svr.RegisterControl(stream, m)
				case *msg.NewWorkConn:
					log.Info("Received NewWorkConn msg.")
					svr.RegisterWorkConn(stream, m)
				case *msg.NewVisitorConn:
					log.Info("Received NewVisitorConnMsg.")
					go svr.handleNewVisitor(stream, m)
				}
			}
		}(c)
	}
}
