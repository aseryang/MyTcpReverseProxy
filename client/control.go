package client

import (
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	fio "github.com/fatedier/golib/io"
	"github.com/hashicorp/yamux"
	"io"
	"net"
)
import "github.com/aseryang/MyTcpReverseProxy/models/msg"

var sshLocalAddr = "127.0.0.1:22"

type Control struct {
	c       net.Conn
	session *yamux.Session
	sendCh  chan (msg.Message)
	readCh  chan (msg.Message)
	closeCh chan struct{}
}

func NewControl(conn net.Conn, session *yamux.Session) *Control {
	return &Control{c: conn,
		session: session,
		sendCh:  make(chan msg.Message, 100),
		readCh:  make(chan msg.Message, 100)}
}

func (ctl *Control) Run() {
	go ctl.reader()
	go ctl.writer()
	go ctl.msgHandle()
	ctl.startTcpProxy()
	<-ctl.closeCh
}

func (ctl *Control) startTcpProxy() {
	newProxyMsg := &msg.NewProxy{ProxyName: "ssh",
		ProxyType:  "tcp",
		RemotePort: 3456}
	if err := msg.WriteMsg(ctl.c, newProxyMsg); err != nil {
		log.Info("Send msg failed! MsgType=NewProxy.")
		return
	}
	log.Info("Send NewProxyMsg succeed, ProxyName=%s,ProxyType=%s,remotePort=%d .",
		newProxyMsg.ProxyName,
		newProxyMsg.ProxyType,
		newProxyMsg.RemotePort)
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
	m = &msg.NewWorkConn{RunId: "1"}
	if err := msg.WriteMsg(workConn, m); err != nil {
		workConn.Close()
	}
	log.Info("Send NewWorkConn succeed.")
	var startMsg msg.StartWorkConn
	if err := msg.ReadMsgInto(workConn, &startMsg); err != nil {
		return
	}
	log.Info("Received StartWorkConn msg.")
	localConn, err := net.Dial("tcp", sshLocalAddr)
	if err != nil {
		return
	}
	fio.Join(localConn, workConn)
}

func (ctl *Control) handleNewProxyResp(m msg.Message) {

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
		}
	}
}
