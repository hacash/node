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
		// 关闭搜寻节点
		return
	}

	swapped := atomic.CompareAndSwapUint32(&p.isInFindingNode, 0, 1)
	if !swapped {
		return // 正在寻找节点间
	}

	tarnodes := make([]*fdNodes, 0, 9)
	fdndmax := 8

	alradySuckAddrStrs := make(map[string]bool)

	// 查找
	p.doFindNearestPublicNodes(nil, nil, &tarnodes, fdndmax, alradySuckAddrStrs)

	// 按顺序连接
	fdnodenum := len(tarnodes)
	realconnectnum := 0 // 真实发起连接的节点数量
	fmt.Printf("[P2P] ")
	if fdnodenum > 0 {
		fmt.Printf("find %d nearest public nodes to update topology table, ", fdnodenum)
	}

	// 依次判断亲源和连接
	for i := fdnodenum - 1; i >= 0; i-- {
		node := tarnodes[i]
		//fmt.Println(node.TcpAddr.String())
		// 判断拓扑图更新状态
		if isCanUpdateTopologyDistancePeerIDTable(p.peerSelf.ID, node.ID, p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax) {
			// 拓扑亲源关系检查通过，发起连接
			// fmt.Println(len(p.BackboneNodeTable), p.Config.BackboneNodeTableSizeMax, hex.EncodeToString(node.ID))
			if realconnectnum == 0 {
				fmt.Printf("\n") // 打印美观
			}
			realconnectnum++
			p.ConnectNodeInitiative(node.TcpAddr)
		}

	}
	if realconnectnum > 0 {
		time.Sleep(time.Second * 2)
		fmt.Printf("[P2P] ") // 打印美观
	}
	// 打印最新的连接情况
	p.PeerChangeMux.RLock()
	fmt.Printf("connected peers: %d public, %d private, total: %d nodes, %d conns.\n",
		len(p.BackboneNodeTable), len(p.OrdinaryNodeTable), p.AllNodesLen, p.TemporaryConnsLen)
	p.PeerChangeMux.RUnlock()

	// 节点全部查询成功
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
	// 尝试连接
	ckpubconn, e1 := dialTimeoutWithHandshakeSignal("tcp", addr.String(), time.Second*5)
	if e1 != nil {
		return gotNodes
	}
	go func() {
		// 10 秒后关闭
		time.Sleep(time.Second * time.Duration(closesec))
		ckpubconn.Close()
	}()
	// 请求最近的公网节点
	e0 := sendTcpMsg(ckpubconn, P2PMsgTypeRequestNearestPublicNodes, nil)
	if e0 != nil {
		//fmt.Println(e0)
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	ndn := make([]byte, 1)
	_, e2 := io.ReadFull(ckpubconn, ndn)
	if e2 != nil {
		//fmt.Println(e2)
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	ndnum := int(ndn[0])
	if ndnum == 0 || ndnum > 200 {
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	nodebts := make([]byte, ndnum*FindNodeSize) // ip + port + pid
	_, e3 := io.ReadFull(ckpubconn, nodebts)
	if e3 != nil {
		//fmt.Println(e3)
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	ckpubconn.Close() // 关闭
	// 解析nodebts
	//fmt.Println(nodebts)
	nodes := parseFindNodesFromBytes(nodebts)
	//fmt.Println(nodes[0].ID)
	// 排除掉我已经连接的公网节点，避免死循环，写入关系表
	sortids := make([]PeerID, 0)
	for _, nd := range nodes {
		if p.getPeerByID(nd.ID) == nil {
			if istok, _, newids, _ := insertUpdateTopologyDistancePeerIDTable(p.peerSelf.ID, nd.ID, sortids, 200); istok {
				sortids = newids // 插入
			}
		}
	}
	// 按亲源关系排序节点
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

	// 判断添加
	if addr != nil && tarpid != nil {
		// 避免自己
		if bytes.Compare(p.Config.ID, tarpid) == 0 {
			return // 查找到自己，返回
		}
		hsp := p.getPeerByID(tarpid)
		// 加上未连接的节点
		if hsp == nil {
			// 并且去重
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

	// 数量
	if len(*tarnodes) >= fdndmax {
		return
	}
	// 查询
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
		// tcp连接读取公网节点, 10 秒timeout
		// 发送获取公网节点列表消息
		gotNodes = p.readEffectivePublicNodesFromTcpTimeout(addr, 10)
	}

	// 递归查找
	for _, v := range gotNodes {
		//fmt.Println("p.findNodes()", hex.EncodeToString(v.ID), v.TcpAddr.String())
		p.doFindNearestPublicNodes(v.TcpAddr, v.ID, tarnodes, fdndmax, alradySuckAddrStrs)
		// 检查数量
		if len(*tarnodes) >= fdndmax {
			break
		}
	}

}
