package client

import (
	"encoding/base64"
	"fmt"
	"github.com/aseryang/MyTcpReverseProxy/models/config"
	"github.com/aseryang/MyTcpReverseProxy/models/msg"
	"github.com/aseryang/MyTcpReverseProxy/utils/log"
	fio "github.com/fatedier/golib/io"
	"github.com/fatedier/golib/pool"
	"github.com/hashicorp/yamux"
	"net"
)

type Proxy interface {
	Run() error
	InWorkConn(*yamux.Stream, *msg.StartWorkConn)
	Close()
}

func NewProxy(ctl *Control, pxyConf config.ProxyConf) (pxy Proxy) {
	switch cfg := pxyConf.(type) {
	case *config.TcpProxyConf:
		pxy = &TcpProxy{ctl: ctl, cfg: cfg}
	case *config.StcpProxyConf:
		pxy = &StcpProxy{ctl: ctl, cfg: cfg}
	case *config.UdpProxyConf:
		pxy = &UdpProxy{ctl: ctl, cfg: cfg}
	}
	return
}

type TcpProxy struct {
	ctl *Control
	cfg *config.TcpProxyConf
}

func (pxy *TcpProxy) Run() (err error) {
	newProxyMsg := msg.NewProxy{}
	pxy.cfg.MarshalToMsg(&newProxyMsg)
	if err = msg.WriteMsg(pxy.ctl.c, &newProxyMsg); err != nil {
		log.Info("Send msg failed! MsgType=NewProxy.")
		return
	}
	log.Info("Send NewProxyMsg succeed, ProxyName=%s,ProxyType=%s,remotePort=%d .",
		newProxyMsg.ProxyName,
		newProxyMsg.ProxyType,
		newProxyMsg.RemotePort)
	return
}
func (pxy *TcpProxy) InWorkConn(conn *yamux.Stream, msg *msg.StartWorkConn) {
	localConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d",
		pxy.cfg.LocalSvrConf.LocalIp, pxy.cfg.LocalSvrConf.LocalPort))
	if err != nil {
		log.Error(fmt.Sprintf("TcpProxy connect %s:%d failed.", msg.DstAddr, msg.DstPort))
		return
	}
	fio.Join(localConn, conn)
}
func (pxy *TcpProxy) Close() {

}

type UdpProxy struct {
	ctl *Control
	cfg *config.UdpProxyConf
}

func (pxy *UdpProxy) Run() (err error) {
	newProxyMsg := msg.NewProxy{}
	pxy.cfg.MarshalToMsg(&newProxyMsg)
	if err = msg.WriteMsg(pxy.ctl.c, &newProxyMsg); err != nil {
		log.Info("Send msg failed! MsgType=NewProxy.")
		return
	}
	log.Info("Send NewProxyMsg succeed, ProxyName=%s,ProxyType=%s,remotePort=%d .",
		newProxyMsg.ProxyName,
		newProxyMsg.ProxyType,
		newProxyMsg.RemotePort)
	return
}
func (pxy *UdpProxy) InWorkConn(conn *yamux.Stream, startWorkConnMsg *msg.StartWorkConn) {
	serverAddr := fmt.Sprintf("%s:%d", pxy.cfg.LocalSvrConf.LocalIp, pxy.cfg.LocalSvrConf.LocalPort)
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Error("Resolve UDPAddr failed:%s", err)
		return
	}
	localConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Error(fmt.Sprintf("UdpProxy connect %s:%d failed.", startWorkConnMsg.DstAddr, startWorkConnMsg.DstPort))
		return
	}
	log.Info("UdpProxy start work, connect server addr:%s", serverAddr)
	udpMsgChanOut := make(chan msg.UdpPacket, 64)
	udpMsgChanIn := make(chan msg.UdpPacket, 64)
	var userAddr *net.UDPAddr
	go func() {
		for {
			//var udpMsg msg.UdpPacket
			udpMsg := msg.UdpPacket{}
			if err := msg.ReadMsgInto(conn, &udpMsg); err != nil {
				log.Error("Read UdpPacket msg error:%s", err)
				return
			}
			log.Info("UdpProxy get 1 UdpPacket from user. addr[%s],content[%s],len[%d]",
				udpMsg.RemoteAddr.String(), udpMsg.Content, len(udpMsg.Content))
			userAddr = udpMsg.RemoteAddr
			udpMsgChanOut <- udpMsg
		}
	}()
	go func() {
		for {
			if m, ok := <-udpMsgChanOut; !ok {
				return
			} else {
				data, err := base64.StdEncoding.DecodeString(m.Content)
				if err != nil {
					log.Error("hex decode string failed:%s", err)
					return
				}
				if _, err := localConn.Write(data); err != nil {
					log.Error("Send udp msg to real server failed. error info:%s", err)
					return
				}
				log.Info("Send udp msg to real server succeed. byte[%v],len[%d]", data, len(data))
			}
		}
	}()
	go func() {
		for ; ; {
			buf := pool.GetBuf(1500)
			if n, err := localConn.Read(buf); err != nil {
				log.Error("Read from real dns server failed. error info:%s", err)
			} else {
				log.Info("Get udp server response len: %d,byte: %v", n, buf[:n])

				udpMsg := msg.UdpPacket{Content: base64.StdEncoding.EncodeToString(buf[:n]), RemoteAddr: userAddr}
				udpMsgChanIn <- udpMsg
			}
		}
	}()
	go func() {
		for {
			if m, ok := <-udpMsgChanIn; !ok {
				return
			} else {
				if err := msg.WriteMsg(conn, &m); err != nil {
					log.Error("Send udp msg to mrps failed. error info:%s", err)
					return
				}
				log.Info("Send udp msg ok C[%s] L[%d] A[%s]", m.Content, len(m.Content), m.RemoteAddr.String())
			}
		}
	}()
}
func (pxy *UdpProxy) Close() {}

type StcpProxy struct {
	ctl *Control
	cfg *config.StcpProxyConf
}

func (pxy *StcpProxy) Run() (err error) {
	newProxyMsg := msg.NewProxy{}
	pxy.cfg.MarshalToMsg(&newProxyMsg)
	if err = msg.WriteMsg(pxy.ctl.c, &newProxyMsg); err != nil {
		log.Info("Send msg failed! MsgType=NewProxy.")
		return
	}
	log.Info("Send NewProxyMsg succeed, ProxyName=%s,ProxyType=%s,remotePort=%d .",
		newProxyMsg.ProxyName,
		newProxyMsg.ProxyType,
		newProxyMsg.RemotePort)
	return
}
func (pxy *StcpProxy) InWorkConn(conn *yamux.Stream, msg *msg.StartWorkConn) {
	localConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", pxy.cfg.LocalSvrConf.LocalIp, pxy.cfg.LocalSvrConf.LocalPort))
	if err != nil {
		log.Error(fmt.Sprintf("StcpProxy connect %s:%d failed.", msg.DstAddr, msg.DstPort))
		return
	}
	fio.Join(localConn, conn)
}
func (pxy *StcpProxy) Close() {

}
