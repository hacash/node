package p2pv2

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"
)

/**
 * 搜寻离自己最近的节点
 */
func (p *P2P) findNodes() {
	if p.Config.DisableFindNodes {
		// Close discovery node
		return
	}

	swapped := atomic.CompareAndSwapUint32(&p.isInFindingNode, 0, 1)
	if !swapped {
		return // Looking for inter node
	}

	tarnodes := make([]*fdNodes, 0, 9)
	fdndmax := 8

	alradySuckAddrStrs := make(map[string]bool)

	// lookup
	p.doFindNearestPublicNodes(nil, nil, &tarnodes, fdndmax, alradySuckAddrStrs)

	// Connect in sequence
	fdnodenum := len(tarnodes)
	realconnectnum := 0 // 真实发起连接的节点数量
	fmt.Printf("[P2P] ")
	if fdnodenum > 0 {
		fmt.Printf("find %d nearest public nodes to update topology table, ", fdnodenum)
	}

	// Determine the parent source and connection in turn
	for i := fdnodenum - 1; i >= 0; i-- {
		node := tarnodes[i]
		//fmt.Println(node.TcpAddr.String())
		// Judge the update status of topology map
		if isCanUpdateTopologyDistancePeerIDTable(p.peerSelf.ID, node.ID, p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax) {
			// Topology parent-source relationship check passed, initiate connection
			// fmt.Println(len(p.BackboneNodeTable), p.Config.BackboneNodeTableSizeMax, hex.EncodeToString(node.ID))
			if realconnectnum == 0 {
				fmt.Printf("\n") // Beautiful printing
			}
			realconnectnum++
			p.ConnectNodeInitiative(node.TcpAddr)
		}

	}
	if realconnectnum > 0 {
		time.Sleep(time.Second * 2)
		fmt.Printf("[P2P] ") // Beautiful printing
	}
	// Print the latest connection
	p.PeerChangeMux.RLock()
	fmt.Printf("connected peers: %d public, %d private, total: %d nodes, %d conns.\n",
		len(p.BackboneNodeTable), len(p.OrdinaryNodeTable), p.AllNodesLen, p.TemporaryConnsLen)
	p.PeerChangeMux.RUnlock()

	// All nodes are queried successfully
	atomic.CompareAndSwapUint32(&p.isInFindingNode, 1, 0)
}

func (p *P2P) readEffectivePublicNodesFromTcpTimeout(addr *net.TCPAddr, waitsec int64) []*fdNodes {

	//fmt.Println("readEffectivePublicNodesFromTcp", addr.String())
	//defer func() {
	//	fmt.Println("readEffectivePublicNodesFromTcp  return")
	//}()

	gotNodes := make([]*fdNodes, 0)
	var retmarkch chan bool = make(chan bool, 1)
	var timeout = time.NewTimer(time.Duration(waitsec) * time.Second)

	go func() {
		gotNodes = p.readEffectivePublicNodesFromTcpTimeoutClose(addr, waitsec)
		retmarkch <- true
	}()

	select {
	case <-timeout.C:
		return gotNodes
	case <-retmarkch:
		return gotNodes
	}

}

func (p *P2P) readEffectivePublicNodesFromTcpTimeoutClose(addr *net.TCPAddr, closesec int64) []*fdNodes {

	gotNodes := make([]*fdNodes, 0)
	// Try to connect
	ckpubconn, e1 := dialTimeoutWithHandshakeSignal("tcp", addr.String(), time.Second*5)
	if e1 != nil {
		return gotNodes
	}
	go func() {
		// Close after 10 seconds
		time.Sleep(time.Second * time.Duration(closesec))
		ckpubconn.Close()
	}()
	// Request the nearest public network node
	e0 := sendTcpMsg(ckpubconn, P2PMsgTypeRequestNearestPublicNodes, nil)
	if e0 != nil {
		//fmt.Println(e0)
		ckpubconn.Close() // close
		return gotNodes
	}
	ndn := make([]byte, 1)
	_, e2 := io.ReadFull(ckpubconn, ndn)
	if e2 != nil {
		//fmt.Println(e2)
		ckpubconn.Close() // close
		return gotNodes
	}
	ndnum := int(ndn[0])
	if ndnum == 0 || ndnum > 200 {
		ckpubconn.Close() // close
		return gotNodes
	}
	nodebts := make([]byte, ndnum*FindNodeSize) // ip + port + pid
	_, e3 := io.ReadFull(ckpubconn, nodebts)
	if e3 != nil {
		//fmt.Println(e3)
		ckpubconn.Close() // close
		return gotNodes
	}
	ckpubconn.Close() // close
	// Parse nodebts
	//fmt.Println(nodebts)
	nodes := parseFindNodesFromBytes(nodebts)
	//fmt.Println(nodes[0].ID)
	// Exclude the public network nodes that I have connected to, avoid dead circulation, and write the relationship table
	sortids := make([]PeerID, 0)
	for _, nd := range nodes {
		if p.getPeerByID(nd.ID) == nil {
			if istok, _, newids, _ := insertUpdateTopologyDistancePeerIDTable(p.peerSelf.ID, nd.ID, sortids, 200); istok {
				sortids = newids // insert
			}
		}
	}
	// Sort nodes by parent source relationship
	for _, id := range sortids {
		for _, v := range nodes {
			if bytes.Compare(v.ID, id) == 0 {
				gotNodes = append(gotNodes, v)
				break
			}
		}
	}
	return gotNodes

}

func (p *P2P) doFindNearestPublicNodes(addr *net.TCPAddr, tarpid PeerID, tarnodes *[]*fdNodes, fdndmax int, alradySuckAddrStrs map[string]bool) {

	if addr != nil {
		addrstr := addr.String()
		if alradySuckAddrStrs[addrstr] {
			return
		}
		alradySuckAddrStrs[addrstr] = true
	}

	//fmt.Println("doFindNearestPublicNodes", addrstr, tarpid)

	// Judge add
	if addr != nil && tarpid != nil {
		// Avoid yourself
		if bytes.Compare(p.Config.ID, tarpid) == 0 {
			return // Find yourself and return
		}
		hsp := p.getPeerByID(tarpid)
		// Add unconnected nodes
		if hsp == nil {
			// And weight removal
			ishave := false
			for _, v := range *tarnodes {
				if bytes.Compare(v.ID, tarpid) == 0 {
					ishave = true
					break
				}
			}
			if !ishave {
				//fmt.Println("*tarnodes = append(*tarnodes, &fdNodes{", addr.String())
				*tarnodes = append(*tarnodes, &fdNodes{
					TcpAddr: addr,
					ID:      tarpid,
				})
			}
		}
	}

	// quantity
	if len(*tarnodes) >= fdndmax {
		return
	}
	// query
	gotNodes := make([]*fdNodes, 0)
	if addr == nil {
		// root node
		for _, id := range p.BackboneNodeTable {
			peer := p.getPeerByID(id)
			if peer != nil && peer.PublicIpPort != nil {
				gotNodes = append(gotNodes, &fdNodes{
					TcpAddr: peer.PublicIpPort,
					ID:      peer.ID,
				})
			}
		}
	} else {
		// TCP connection reads public network nodes, 10 seconds timeout
		// Send the message to get the public network node list
		gotNodes = p.readEffectivePublicNodesFromTcpTimeout(addr, 10)
	}

	// recursive lookup 
	for _, v := range gotNodes {
		//fmt.Println("p.findNodes()", hex.EncodeToString(v.ID), v.TcpAddr.String())
		p.doFindNearestPublicNodes(v.TcpAddr, v.ID, tarnodes, fdndmax, alradySuckAddrStrs)
		// Inspection quantity
		if len(*tarnodes) >= fdndmax {
			break
		}
	}

}
