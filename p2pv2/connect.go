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

	//fmt.Println("ConnectNodeInitiative", addr.String())
	//defer func() {
	//	fmt.Println("ConnectNodeInitiative return")
	//}()

	conn, e0 := dialTimeoutWithHandshakeSignal("tcp", addr.String(), time.Second*10)
	if e0 != nil {
		return nil, e0
	}

	// Automatic public network node
	peer := NewEmptyPeer(p, p.msgHandler)
	peer.PublicIpPort = addr

	// Processing messages
	go p.handleNewConn(conn, peer)

	// Send a request to join peer message
	portbts := make([]byte, 4)
	binary.BigEndian.PutUint32(portbts, uint32(p.Config.TCPListenPort))
	idbuf := bytes.NewBuffer(portbts)
	idbuf.Write(p.peerSelf.ID)
	idbuf.Write(p.peerSelf.NameBytes())
	e2 := sendTcpMsg(conn, P2PMsgTypeReportIdKeepConnectAsPeer, idbuf.Bytes())
	if e2 != nil {
		conn.Close()
		return nil, e2
	}

	// Successful return
	return conn, nil

}
