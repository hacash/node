package p2p

import (
	"fmt"
	"net"
)

func (p2p *P2PManager) TryConnectToNode(local_addr *net.TCPAddr, target_addr *net.TCPAddr) {

	if local_addr != nil {
		local_addr.IP = net.IPv4zero
	}
	conn, err := net.DialTCP("tcp", local_addr, target_addr)
	if err != nil {
		fmt.Println("TryConnectToNode error", err)
		//os.Exit(1)
		return
	}

	// hankshake and handle msg
	p2p.handleNewConn(conn)

	return

}
