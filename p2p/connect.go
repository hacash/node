package p2p

import (
	"fmt"
	"net"
)

func (p2p *P2PManager) TryConnectToPeer(local_addr *net.TCPAddr, target_addr *net.TCPAddr) error {

	if p2p.peerManager.CheckHasConnectedWithRemotePublicAddr(target_addr) {
		err := fmt.Errorf("have connected remote addr: " + target_addr.String())
		return err
	}

	fmt.Println("[Peer] Try connect to peer addr:", target_addr.String())

	if local_addr != nil {
		//local_addr.IP = net.IPv4zero
	}
	conn, err := net.DialTCP("tcp", local_addr, target_addr)
	if err != nil {
		fmt.Println("TryConnectToNode error", err)
		//os.Exit(1)
		return err
	}
	isConnPublic := true
	if target_addr.IP.IsLoopback() || target_addr.IP.IsUnspecified() {
		isConnPublic = false
	}
	// hankshake and handle msg
	p2p.handleNewConn(conn, isConnPublic)
	return nil
}
