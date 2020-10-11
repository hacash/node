package p2pv2

import (
	"bytes"
	"io"
	"net"
	"time"
)

/**
 * 搜寻离自己最近的节点
 */
func (p *P2P) findNodes() {

	tarnodes := make([]*fdNodes, 0, 9)
	fdndmax := 8

	// 查找
	p.doFindNearestNode(nil, nil, &tarnodes, fdndmax)

	// 按顺序连接
	for i := len(tarnodes) - 1; i >= 0; i-- {
		node := tarnodes[i]
		haspeer := p.getPeerByID(node.ID)
		// 判断是否已经连接
		if haspeer == nil {
			p.ConnectNodeInitiative(node.TcpAddr)
		}
	}

}

func (p *P2P) readEffectivePublicNodesFromTcp(addr *net.TCPAddr) []*fdNodes {

	gotNodes := make([]*fdNodes, 0)
	// 尝试连接
	isclosed := false
	ckpubconn, e1 := net.DialTimeout("tcp", addr.String(), time.Second*5)
	if e1 != nil {
		return gotNodes
	}
	go func() {
		// 10 秒后关闭
		time.Sleep(time.Second * 10)
		if !isclosed {
			ckpubconn.Close()
		}
	}()
	e0 := sendTcpMsg(ckpubconn, P2PMsgTypeRequestNearestPublicNodes, nil) // 请求最近的公网节点
	if e0 != nil {
		//fmt.Println(e0)
		isclosed = true
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	ndn := make([]byte, 1)
	_, e2 := io.ReadFull(ckpubconn, ndn)
	if e2 != nil {
		//fmt.Println(e2)
		isclosed = true
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	ndnum := int(ndn[0])
	if ndnum == 0 || ndnum > 200 {
		isclosed = true
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	nodebts := make([]byte, ndnum*FindNodeSize) // ip + port + pid
	_, e3 := io.ReadFull(ckpubconn, nodebts)
	if e3 != nil {
		//fmt.Println(e3)
		isclosed = true
		ckpubconn.Close() // 关闭
		return gotNodes
	}
	isclosed = true
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

func (p *P2P) doFindNearestNode(addr *net.TCPAddr, tarpid PeerID, tarnodes *[]*fdNodes, fdndmax int) {
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
		// tcp连接读取公网节点
		// 发送获取公网节点列表消息
		gotNodes = p.readEffectivePublicNodesFromTcp(addr)
	}

	// 递归查找
	for _, v := range gotNodes {
		//fmt.Println("p.findNodes()", hex.EncodeToString(v.ID), v.TcpAddr.String())
		p.doFindNearestNode(v.TcpAddr, v.ID, tarnodes, fdndmax)
		// 检查数量
		if len(*tarnodes) >= fdndmax {
			break
		}
	}

}
