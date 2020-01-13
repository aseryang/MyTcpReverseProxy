package client

import (
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	"github.com/aseryang/MyTcpReverseProxy/utils/version"
	"github.com/hashicorp/yamux"
	"net"
	"runtime"
	"time"
)

type Service struct {
	ctl         *Control
	cfg         config.ClientCommonConf
	pxyCfgs     map[string]config.ProxyConf
	visitorCfgs map[string]config.VisitorConf
}

func NewService(cfg config.ClientCommonConf, pxyCfgs map[string]config.ProxyConf, visitorCfgs map[string]config.VisitorConf) (svr *Service) {
	return &Service{cfg: cfg,
		pxyCfgs:     pxyCfgs,
		visitorCfgs: visitorCfgs}
}

func (svr *Service) Run() {
	for {
		if c, session, runId, err := svr.login(); err != nil {
			time.Sleep(time.Second * 2)
		} else {
			svr.ctl = NewControl(c, session, runId, svr.pxyCfgs, svr.visitorCfgs)
			svr.ctl.Run()
			break
		}
	}
}

func (svr *Service) login() (c net.Conn, session *yamux.Session, runId string, err error) {
	c, err = net.Dial("tcp", fmt.Sprintf("%s:%d", svr.cfg.ServerAddr, svr.cfg.ServerPort))
	if err != nil {
		return
	}
	session, err = yamux.Client(c, nil)
	if err != nil {
		return
	}
	stream, _ := session.OpenStream()
	c = stream
	now := time.Now().Unix()
	logingMsg := &msg.Login{Arch: runtime.GOARCH,
		Os:        runtime.GOOS,
		Version:   version.Full(),
		Timestamp: now}
	if err = msg.WriteMsg(stream, logingMsg); err != nil {
		return
	}
	log.Info("Send Login msg succeed.")
	var loginRespMsg msg.LoginResp
	if err = msg.ReadMsgInto(stream, &loginRespMsg); err != nil {
		fmt.Errorf("%s", loginRespMsg.Error)
		return
	}
	runId = loginRespMsg.RunId
	log.Info("Received LoginResp msg.RunId:%s", loginRespMsg.RunId)
	return
}
