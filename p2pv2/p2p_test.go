package p2pv2

import (
	"net"
	"testing"
	"time"
)

func Test_startNode1(t *testing.T) {

	p2pcnf := NewEmptyP2PConfig()
	p2pcnf.TCPListenPort = 33317

	p2pNode := NewP2P(p2pcnf)
	p2pNode.Start()

	// hold
	time.Sleep(time.Hour * 24)
}

func Test_startNode2(t *testing.T) {

	p2pcnf := NewEmptyP2PConfig()
	p2pcnf.TCPListenPort = 33318

	p2pNode := NewP2P(p2pcnf)
	p2pNode.Start()

	tartcp := net.TCPAddr{
		IP:   net.IPv4zero,
		Port: 33317,
	}

	go func() {
		for {
			//fmt.Println("ConnectNodeInitiative")
			conn, e0 := p2pNode.ConnectNodeInitiative(&tartcp)
			if e0 != nil {
				panic(e0)
			}
			time.Sleep(time.Second * 3)
			conn.Close()
			time.Sleep(time.Second * 3)
		}
	}()

	// hold
	time.Sleep(time.Hour * 24)
}
