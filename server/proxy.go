package server

import (
	"encoding/base64"
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/consts"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	network "github.com/aseryang/MyTcpReverseProxy/utils/net"
	fio "github.com/fatedier/golib/io"
	"github.com/fatedier/golib/pool"
	"net"
)

type Proxy interface {
	Run() error
	Stop()
	GetControl() *Control
	GetConfig() config.ProxyConf
}

func NewProxy(ctl *Control, msg msg.NewProxy) (pxy Proxy) {
	switch msg.ProxyType {
	case consts.TcpProxy:
		cfg := config.TcpProxyConf{}
		cfg.ProxyName = msg.ProxyName
		cfg.RemotePort = msg.RemotePort
		pxy = &TcpProxy{ctl: ctl, cfg: cfg}
	case consts.StcpProxy:
		cfg := config.StcpProxyConf{}
		cfg.ProxyName = msg.ProxyName
		pxy = &StcpProxy{ctl: ctl, cfg: cfg}
	case consts.UdpProxy:
		cfg := config.UdpProxyConf{}
		cfg.ProxyName = msg.ProxyName
		cfg.RemotePort = msg.RemotePort
		pxy = &UdpProxy{ctl: ctl, cfg: cfg}
	}
	return
}

type TcpProxy struct {
	ctl *Control
	cfg config.TcpProxyConf
	ln  net.Listener
}

func (pxy *TcpProxy) GetControl() *Control {
	return pxy.ctl
}
func (pxy *TcpProxy) GetConfig() config.ProxyConf {
	return &pxy.cfg
}

func (pxy *TcpProxy) Run() error {
	log.Info("New TcpProxy listen on port %d.", pxy.cfg.RemotePort)
	go func() {
		var err error
		pxy.ln, err = net.Listen("tcp", fmt.Sprintf("%s:%d", "0.0.0.0", pxy.cfg.RemotePort))
		if err != nil {
			log.Info("Proxy Listen ret failed!")
			return
		}
		for {
			userConn, err := pxy.ln.Accept()
			if err != nil {
				return
			}
			handleUserTcpConnection(pxy, userConn)
		}
	}()
	resp := &msg.NewProxyResp{
		ProxyName: pxy.cfg.ProxyName,
	}
	pxy.ctl.sendCh <- resp
	log.Info("Send NewProxyResp msg succeed.")
	return nil
}
func handleUserTcpConnection(pxy Proxy, userConn net.Conn) {
	log.Info("New user connection...")
	workConn, err := pxy.GetControl().GetWorkConn()
	if err != nil {
		return
	}
	err = msg.WriteMsg(workConn, &msg.StartWorkConn{ProxyName: pxy.GetConfig().GetBaseInfo().ProxyName})
	if err != nil {
		return
	}
	log.Info("Send StartWorkConn msg succeed.")
	fio.Join(userConn, workConn)
}
func (pxy *TcpProxy) Stop() {
	if pxy.ln != nil {
		pxy.ln.Close()
		log.Info("Tcp proxy:%s closed.", pxy.cfg.GetBaseInfo().ProxyName)
	} else {
		log.Error("TcpProxy tcp listener point is nil!")
	}
}

type StcpProxy struct {
	ctl *Control
	cfg config.StcpProxyConf
	ln  *network.CustomListener
}

func (pxy *StcpProxy) GetControl() *Control {
	return pxy.ctl
}
func (pxy *StcpProxy) GetConfig() config.ProxyConf {
	return &pxy.cfg
}
func (pxy *StcpProxy) Run() (err error) {
	ln := pxy.ctl.vm.NewVisitorListener(pxy.GetConfig().GetBaseInfo().ProxyName)
	for ; ; {
		var conn net.Conn
		conn, err = ln.Accept()
		if err != nil {
			return
		}
		go handleUserTcpConnection(pxy, conn)
	}
	return
}
func (pxy *StcpProxy) Stop() {
	pxy.ln.Close()
}

type UdpProxy struct {
	ctl  *Control
	cfg  config.UdpProxyConf
	conn *net.UDPConn
}

func (pxy *UdpProxy) GetConfig() config.ProxyConf {
	return &pxy.cfg
}
func (pxy *UdpProxy) GetControl() *Control {
	return pxy.ctl
}
func (pxy *UdpProxy) Run() (err error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", "0.0.0.0", pxy.cfg.RemotePort))
	if err != nil {
		log.Error("Resolve udpAddr err:%s", err)
		return
	}
	pxy.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Error("Listen udp on addr:%s error,error:%s", addr, err)
	}
	go func() {
		isStart := false
		var workConn net.Conn
		for ; ; {
			data := pool.GetBuf(1500)
			n, userAddr, err := pxy.conn.ReadFromUDP(data)
			if err != nil {
				log.Error("failed to read udp msg, because of:%s", err)
			}
			log.Info("User data, byte: %v", data[:n])
			udpMsg := msg.UdpPacket{}
			udpMsg.Content = base64.StdEncoding.EncodeToString(data[:n])
			udpMsg.RemoteAddr = userAddr
			log.Info("userData[%s]ReadRet[%d]L[%d]A[%s]",
				udpMsg.Content, n, len(udpMsg.Content), userAddr.String())
			if !isStart {
				isStart = true
				workConn, err = pxy.ctl.GetWorkConn()
				err = msg.WriteMsg(workConn, &msg.StartWorkConn{ProxyName: pxy.GetConfig().GetBaseInfo().ProxyName})
				if err != nil {
					log.Error("Send startWorkConn msg failed:%s", err)
					return
				}
				go func() {
					for ; ; {
						var udpMsgResp msg.UdpPacket
						if err := msg.ReadMsgInto(workConn, &udpMsgResp); err != nil {
							log.Error("Read UdpPacket msg error:%s", err)
							return
						} else {
							data, err := base64.StdEncoding.DecodeString(udpMsgResp.Content)
							if err != nil {
								log.Error("hex decode base64 string failed:%s", err)
								return
							}
							if _, err := pxy.conn.WriteToUDP(data, udpMsgResp.RemoteAddr); err != nil {
								log.Error("Send msg to user failed:%s, a[%s]c[%s]l[%d]",
									err, udpMsgResp.RemoteAddr.String(), udpMsgResp.Content, len(udpMsgResp.Content))
							} else {
								log.Info("Send udp msg to user succeed.c[%s],a[%s]",
									udpMsgResp.Content, udpMsgResp.RemoteAddr.String())
							}
						}
					}
				}()
			}
			err = msg.WriteMsg(workConn, &udpMsg)
			if err != nil {
				log.Error("Send UdpPacket msg failed:%s", err)
				return
			} else {
				log.Info("Send UdpPacket ok.[%s][%d][%s] ",
					udpMsg.Content, len(udpMsg.Content), udpMsg.RemoteAddr.String())
			}
		}
	}()
	return
}
func (pxy *UdpProxy) Stop() {
	pxy.conn.Close()
}
