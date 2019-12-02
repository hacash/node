package p2p

import (
	"fmt"
	"net"
)

func (p2p *P2PManager) TryConnectToNode(local_addr *net.TCPAddr, target_addr *net.TCPAddr) error {

	conn, err := net.DialTCP("tcp", local_addr, target_addr)
	if err != nil {
		fmt.Println("TryConnectToNode error", err)
		//os.Exit(1)
		return err
	}

	// hankshake and handle msg
	p2p.handleNewConn(conn)

	return nil

}
