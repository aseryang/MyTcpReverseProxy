package client

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	fio "github.com/fatedier/golib/io"
	"net"
	"time"
)

type Visitor interface {
	Run() error
	Close()
}

type BaseVisitor struct {
	listener net.Listener
	ctl      *Control
}
type StcpVisitor struct {
	*BaseVisitor
	cfg *config.StcpVisitorConf
}

func NewVisitor(ctl *Control, cfg config.VisitorConf) (visitor Visitor) {
	baseVisitor := BaseVisitor{ctl: ctl}
	switch cfg := cfg.(type) {
	case *config.StcpVisitorConf:
		visitor = &StcpVisitor{
			BaseVisitor: &baseVisitor,
			cfg:         cfg,
		}
	}
	return
}
func (sv *StcpVisitor) Run() (err error) {
	log.Info("Stcp visitor run, tcp listen on %s", fmt.Sprintf("%s:%d", sv.cfg.BindAddr, sv.cfg.BindPort))
	sv.listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", sv.cfg.BindAddr, sv.cfg.BindPort))
	if err != nil {
		return
	}
	go sv.worker()
	return
}
func (sv *StcpVisitor) worker() {
	for ; ; {
		conn, err := sv.listener.Accept()
		if err != nil {
			fmt.Println("Stcp local listener closed")
			return
		}
		go sv.handleConn(conn)
	}
}
func (sv *StcpVisitor) handleConn(userConn net.Conn) {
	visitorConn, err := sv.ctl.session.OpenStream()
	if err != nil {
		return
	}
	now := time.Now().Unix()
	newVisitorConnMsg := &msg.NewVisitorConn{ProxyName: sv.cfg.ServerName, Timestamp: now,}
	err = msg.WriteMsg(visitorConn, newVisitorConnMsg)
	if err != nil {
		fmt.Println("Send NewVisitorConnMsg to server error:", err)
		return
	}
	var newVisitorConnRespMsg msg.NewVisitorConnResp
	err = msg.ReadMsgInto(visitorConn, newVisitorConnRespMsg)
	if err != nil {
		fmt.Println("Get newVisitorConnRespMsg error:", err)
		return
	}
	fio.Join(userConn, visitorConn)
}
func (sv *StcpVisitor) Close() {

}
