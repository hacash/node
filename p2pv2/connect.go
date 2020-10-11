package p2pv2

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"
)

/**
 * 连接指定的节点
 */
func (p *P2P) ConnectNodeInitiative(addr *net.TCPAddr) (net.Conn, error) {

	conn, e0 := net.DialTimeout("tcp", addr.String(), time.Second*5)
	if e0 != nil {
		return nil, e0
	}
	// 自动为公网节点
	peer := NewEmptyPeer(p, p.msgHandler)
	peer.PublicIpPort = addr

	// 处理消息
	go p.handleNewConn(conn, peer)

	// 发送请求加入 peer 消息
	portbts := make([]byte, 4)
	binary.BigEndian.PutUint32(portbts, uint32(p.Config.TCPListenPort))
	idbuf := bytes.NewBuffer(portbts)
	idbuf.Write(p.peerSelf.ID)
	idbuf.Write(p.peerSelf.NameBytes())
	e1 := sendTcpMsg(conn, P2PMsgTypeReportIdKeepConnectAsPeer, idbuf.Bytes())
	if e1 != nil {
		conn.Close()
		return nil, e1
	}

	// 成功返回
	return conn, nil

}
