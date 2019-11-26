package client

import (
	"MyTcpReverseProxy/models/msg"
	"MyTcpReverseProxy/utils/log"
	"fmt"
	"github.com/fatedier/frp/utils/version"
	"github.com/hashicorp/yamux"
	"net"
	"runtime"
	"time"
)

var addr = "106.12.17.239:4567"

type Service struct {
	ctl *Control
}

func NewService() (svr *Service) {
	return &Service{}
}

func (svr *Service) Run() {
	for {
		if c, session, err := svr.login(); err != nil {
			time.Sleep(time.Second * 2)
		} else {
			svr.ctl = NewControl(c, session)
			svr.ctl.Run()
			break
		}
	}
}

func (svr *Service) login() (c net.Conn, session *yamux.Session, err error) {
	c, err = net.Dial("tcp", addr)
	session, err = yamux.Client(c, nil)
	stream, _ := session.OpenStream()
	c = stream
	now := time.Now().Unix()
	logingMsg := &msg.Login{Arch: runtime.GOARCH,
		Os:        runtime.GOOS,
		Version:   version.Full(),
		Timestamp: now,
		RunId:     "1"}
	if err = msg.WriteMsg(stream, logingMsg); err != nil {
		return
	}
	log.Info("Send Login msg succeed.")
	var loginRespMsg msg.LoginResp
	if err = msg.ReadMsgInto(stream, loginRespMsg); err != nil {
		fmt.Errorf("%s", loginRespMsg.Error)
		return
	}
	log.Info("Received LoginResp msg.")
	return
}
