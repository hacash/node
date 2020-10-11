package p2pv2

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math/rand"
	"net"
)

func (p *P2P) listen(port int) {

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.IPv4zero, port, ""})
	//laddr := net.TCPAddr{net.IPv4zero, p2p_other.config.TCPListenPort, ""}
	//listener, err := reuseport.Listen("tcp", laddr.String())
	if err != nil {
		panic(err)
	}

	defer listener.Close()

	fmt.Printf("[P2P] Start node %s id:%s listen port %d.\n", p.Config.Name, hex.EncodeToString(p.Config.ID), port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		go p.handleNewConn(conn, nil)
	}

}

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

	//fmt.Println("handleNewConn")
	for {
		// read
		lengthBuf := make([]byte, 4)
		_, e0 := io.ReadFull(conn, lengthBuf)
		//fmt.Println(lengthBuf)
		if e0 != nil {
			//fmt.Println(e0)
			break // error
		}
		//fmt.Println("next")
		length := binary.BigEndian.Uint32(lengthBuf)
		bodyBuf := make([]byte, length)
		_, e1 := io.ReadFull(conn, bodyBuf)
		//fmt.Println(bodyBuf)
		if e1 != nil {
			//fmt.Println(e1)
			break // error
		}

		// deal msg body
		//fmt.Println("ReadFull deal msg body: ", bodyBuf)
		p.handleConnMsg(connid, conn, createPeer, bodyBuf)
	}

	// drop

	//fmt.Println("p.dropPeerByConnIDUnsafe(connid)", connid)
	p.PeerChangeMux.Lock()
	p.dropPeerByConnIDUnsafe(connid)
	delete(p.TemporaryConns, connid)
	p.PeerChangeMux.Unlock()

}
