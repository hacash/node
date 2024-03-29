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
	//p.TemporaryConns[connid] = conn
	p.TemporaryConns.Store(connid, conn)
	p.TemporaryConnsLen += 1
	p.PeerChangeMux.Unlock()

	if createPeer == nil {
		createPeer = NewEmptyPeer(p, p.msgHandler)
	}
	createPeer.connid = connid
	createPeer.conn = conn

	// Has the handshake passed
	//fmt.Println("hacashnodep2phandshake")

	for {
		// read
		lengthBuf := make([]byte, 4)
		_, e0 := io.ReadFull(conn, lengthBuf)
		//fmt.Println(lengthBuf)
		if e0 != nil {
			//fmt.Println("handleNewConn _, e0 := io.ReadFull(conn, lengthBuf)")
			//fmt.Println(e0)
			break // error
		}
		//fmt.Println("next")
		length := binary.BigEndian.Uint32(lengthBuf)
		if length == 0 {
			break // error
		}
		if length > P2PMsgDataMaxSize {
			break // Maximum message length 10 MB
		}
		// Read message content
		bodyBuf := make([]byte, length)
		_, e1 := io.ReadFull(conn, bodyBuf)
		//fmt.Println(bodyBuf)
		if e1 != nil {
			//fmt.Println("handleNewConn _, e1 := io.ReadFull(conn, bodyBuf)")
			//fmt.Println(e1)
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
	p.TemporaryConns.Delete(connid)
	p.TemporaryConnsLen -= 1
	p.PeerChangeMux.Unlock()

}
