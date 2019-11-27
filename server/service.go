package server

import (
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/hashicorp/yamux"
	"net"
)

type Service struct {
	ln  net.Listener
	ctl *Control
}

var svnListenAddr = "0.0.0.0:4567"

func NewService() (svr *Service) {
	svr = &Service{}
	ln, err := net.Listen("tcp", svnListenAddr)
	if err != nil {
		return
	}
	log.Info("Service listen on %s succeed.", svnListenAddr)
	svr.ln = ln

	return
}
func (svr *Service) RegisterControl(conn net.Conn, m msg.Message) {
	svr.ctl = NewControl()
	svr.ctl.conn = conn
	go svr.ctl.Run()
}
func (svr *Service) RegisterWorkConn(conn net.Conn, m msg.Message) {
	svr.ctl.RegisterWorkConn(conn)
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
				}
			}
		}(c)
	}
}
