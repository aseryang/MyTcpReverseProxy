package server

import (
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"fmt"
	"github.com/fatedier/frp/utils/version"
	fio "github.com/fatedier/golib/io"
	"io"
	"net"
	"time"
)

type Control struct {
	sendCh        chan msg.Message
	readCh        chan msg.Message
	conn          net.Conn
	workConnCh    chan net.Conn
	connPoolCount int
	closeCh       chan struct{}
}

func NewControl() *Control {
	var poolCount = 2
	return &Control{sendCh: make(chan msg.Message, 10),
		readCh:        make(chan msg.Message, 10),
		connPoolCount: poolCount,
		workConnCh:    make(chan net.Conn, poolCount)}
}
func (ctl *Control) Run() {
	loginRespMsg := &msg.LoginResp{Version: version.Full()}
	msg.WriteMsg(ctl.conn, loginRespMsg)
	go ctl.reader()
	go ctl.writer()
	for i := 0; i < ctl.connPoolCount; i++ {
		ctl.sendCh <- &msg.ReqWorkConn{}
	}

	go ctl.msgHandle()
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

func (ctl *Control) handleUserTcpConnection(userConn net.Conn) {
	log.Info("New user connection...")
	workConn, err := ctl.GetWorkConn()
	if err != nil {
		return
	}
	err = msg.WriteMsg(workConn, &msg.StartWorkConn{ProxyName: "ssh"})
	if err != nil {
		return
	}
	log.Info("Send StartWorkConn msg succeed.")
	fio.Join(userConn, workConn)
}

func (ctl *Control) handleNewProxy(pxymsg msg.NewProxy) {
	log.Info("NewProxy listen on port %d.", pxymsg.RemotePort)
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", pxymsg.RemotePort))
		if err != nil {
			log.Info("Proxy Listen ret failed!")
			return
		}
		for {
			userConn, err := listener.Accept()
			if err != nil {
				return
			}
			go ctl.handleUserTcpConnection(userConn)
		}
	}()
	resp := &msg.NewProxyResp{
		ProxyName: "ssh",
	}
	ctl.sendCh <- resp
	log.Info("Send NewProxyResp msg succeed.")
}
func (ctl *Control) handleCloseProxy(m msg.Message) {

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
			go ctl.handleNewProxy(*m)
		case *msg.CloseProxy:
			log.Info("Received CloseProxy msg.")
			go ctl.handleCloseProxy(m)
		}
	}
}
