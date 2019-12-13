package p2p

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func (p2p *P2PManager) TryConnectToPeer(local_addr *net.TCPAddr, target_addr *net.TCPAddr) error {

	taripportbts := ParseTCPAddrToIPPortBytes(target_addr)
	if p2p.selfRemotePublicIP != nil {
		myipport := make([]byte, 6)
		copy(myipport[0:4], p2p.selfRemotePublicIP[0:4])
		binary.BigEndian.PutUint16(myipport[4:6], uint16(p2p.config.TCPListenPort))
		if bytes.Compare(taripportbts, myipport) == 0 {
			err := fmt.Errorf("cannot connect to my self remote addr: " + target_addr.String())
			return err
		}
	}

	if p2p.peerManager.CheckHasConnectedToRemotePublicAddrByByte(taripportbts) {
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
	if target_addr.IP.IsLoopback() {
		// isConnPublic = false
	}
	// hankshake and handle msg
	p2p.handleNewConn(conn, isConnPublic)
	return nil
}

// Static Boot Nodes
func (p2p *P2PManager) tryConnectToStaticBootNodes() {
	if len(p2p.config.StaticHnodeAddrs) == 0 {
		return
	}
	// try
	for _, tcp := range p2p.config.StaticHnodeAddrs {
		err := p2p.TryConnectToPeer(nil, tcp)
		if err != nil {
			fmt.Println("[P2P Error] Try Connect To Static Boot Node:", err)
		}
	}
}
