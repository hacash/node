package p2pv2

import (
	"encoding/binary"
	"io"
	"math/rand"
	"net"
)

func (p *P2P) handleNewConn(conn net.Conn, createPeer *Peer) {
	connid := rand.Uint64()

	p.PeerChangeMux.Lock()
	p.TemporaryConns[connid] = conn
	p.PeerChangeMux.Unlock()

	if createPeer == nil {
		createPeer = NewEmptyPeer(p, p.msgHandler)
	}
	createPeer.connid = connid
	createPeer.conn = conn

	// 是否已经握手通过
	//fmt.Println("hacashnodep2phandshake")

	for {
		// read
		lengthBuf := make([]byte, 4)
		_, e0 := io.ReadFull(conn, lengthBuf)
		//fmt.Println(lengthBuf)
		if e0 != nil {
			// fmt.Println("handleNewConn _, e0 := io.ReadFull(conn, lengthBuf)")
			// fmt.Println(e0)
			break // error
		}
		//fmt.Println("next")
		length := binary.BigEndian.Uint32(lengthBuf)
		if length == 0 {
			break // 错误
		}
		if length > P2PMsgDataMaxSize {
			break // 最大消息长度 10 MB
		}
		// 读取消息内容
		bodyBuf := make([]byte, length)
		_, e1 := io.ReadFull(conn, bodyBuf)
		//fmt.Println(bodyBuf)
		if e1 != nil {
			// fmt.Println("handleNewConn _, e1 := io.ReadFull(conn, bodyBuf)")
			// fmt.Println(e1)
			break // error
		}

		// deal msg body
		//fmt.Println("ReadFull deal msg body: ", bodyBuf)
		go p.handleConnMsg(connid, conn, createPeer, bodyBuf)
	}

	// drop
	conn.Close()

	//fmt.Println("p.dropPeerByConnIDUnsafe(connid)", connid)
	p.PeerChangeMux.Lock()
	p.dropPeerByConnIDUnsafe(connid)
	delete(p.TemporaryConns, connid)
	p.PeerChangeMux.Unlock()

}
