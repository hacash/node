package p2pv2

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"time"
)

func (p *P2P) handleConnMsg(connid uint64, conn net.Conn, peer *Peer, msg []byte) {

	//fmt.Println("handleConnMsg", msg)

	ct := time.Now()
	peer.activeTime = &ct // 活跃时间

	// 处理消息
	msgty := msg[0]
	msgbody := msg[1:]

	switch msgty {

	case P2PMsgTypeCustomer:
		// 客户端消息到达
		if len(msgbody) < 2 {
			break
		}
		if p.msgHandler != nil {
			mty := binary.BigEndian.Uint16(msgbody[0:2])
			mbody := msgbody[2:]
			//fmt.Println("P2PMsgTypeCustomer", mty, mbody)
			p.msgHandler.OnMsgData(p, peer, mty, mbody)
		}

	case P2PMsgTypePing:
		// ping pong
		sendTcpMsg(conn, P2PMsgTypePong, nil)
		break

	case P2PMsgTypePong:
		// 收到 pong 消息，do nothing
		break

	case P2PMsgTypeRequestNearestPublicNodes:
		// 发送我连接的全部公网节点
		nodes := make([]*fdNodes, 0)
		for _, pid := range p.BackboneNodeTable {
			if peer := p.getPeerByID(pid); peer != nil && peer.PublicIpPort != nil {
				nodes = append(nodes, &fdNodes{
					TcpAddr: peer.PublicIpPort,
					ID:      peer.ID,
				})
				if len(nodes) >= 100 {
					break
				}
			}
		}
		buf := bytes.NewBuffer([]byte{uint8(len(nodes))}) // 第一位为数量 <= 200
		buf.Write(serializeFindNodesToBytes(nodes))
		conn.Write(buf.Bytes()) // 发送全部公网节点
		conn.Close()            // 立即关闭连接
		break

	case P2PMsgTypeRequestIDForPublicNodeCheck:
		conn.Write(p.peerSelf.ID) // 发送我的id
		conn.Close()              // 立即关闭连接
		break

	case P2PMsgTypeAnswerIdKeepConnectAsPeer:

		// 对方也愿意持久连接
		if len(msgbody) != PeerIDSize+PeerNameSize {
			break
		}
		if peer.ID != nil {
			break
		}
		// 我主动连接的对方，对方肯定是公网 node，不用检测判断
		peerId := msgbody[0:PeerIDSize]
		peerName := string(msgbody[PeerIDSize:])
		peer.ID = peerId
		peer.Name = strings.TrimRight(peerName, " ")
		// 添加进公网节点表
		p.PeerChangeMux.Lock()
		p.AllNodes[string(peerId)] = peer
		p.addPeerIntoTargetTableUnsafe(&p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax, peer)
		p.PeerChangeMux.Unlock()
		// 通知连接成功
		peer.notifyConnect()
		// 判断是否需要执行第一次查找节点
		if len(p.BackboneNodeTable) == 1 {
			p.findNodes()
		}
		break

	case P2PMsgTypeReportIdKeepConnectAsPeer:
		// 请求对方持久连接
		//fmt.Println("P2PMsgTypeReportIdKeepConnectAsPeer")

		if len(msgbody) != 4+PeerIDSize+PeerNameSize {
			break
		}
		if peer.ID != nil {
			break
		}
		port := binary.BigEndian.Uint32(msgbody[0:4])
		peerId := msgbody[4 : 4+PeerIDSize]
		peerName := string(msgbody[4+PeerIDSize:])
		peer.ID = peerId
		peer.Name = strings.TrimRight(peerName, " ")
		// 连接加入节点
		if oldpeer, hasp := p.AllNodes[string(peerId)]; hasp {
			// 已经存在如何处理？
			peer.ReplacingCopyInfo(oldpeer)
			oldpeer.Disconnect()        // 直接断开旧的连接
			time.Sleep(time.Second * 1) // 休眠 1 秒
		}
		// 回复我也愿意连接的消息
		rplidbuf := bytes.NewBuffer(p.peerSelf.ID)
		rplidbuf.Write(p.peerSelf.NameBytes())
		e3 := sendTcpMsg(conn, P2PMsgTypeAnswerIdKeepConnectAsPeer, rplidbuf.Bytes())
		if e3 != nil {
			// 发送消息出错
			conn.Close()
			break
		}
		// 添加为新的节点
		p.PeerChangeMux.Lock()
		p.AllNodes[string(peerId)] = peer
		p.addPeerIntoUnfamiliarTableUnsafe(peer)
		p.PeerChangeMux.Unlock()

		go func() {
			// 通知正式连接上
			defer peer.notifyConnect()
			// 发送判断是否为公网节点的消息
			tcp, e0 := net.ResolveTCPAddr("tcp", conn.RemoteAddr().String())
			if e0 != nil {
				return
			}
			tcp.Port = int(port) // 公网监听端口
			// 尝试连接
			ckpubconn, e1 := net.DialTimeout("tcp", tcp.String(), time.Second*10)
			if e1 != nil {
				return
			}
			sendTcpMsg(ckpubconn, P2PMsgTypeRequestIDForPublicNodeCheck, nil)
			checkpid := make([]byte, PeerIDSize)
			rdn, e2 := io.ReadFull(ckpubconn, checkpid)
			if e2 != nil {
				return
			}
			if rdn == PeerIDSize {
				// 判断节点为公网节点
				if bytes.Compare(checkpid, peerId) == 0 {
					//fmt.Println("OK PublicIpPort:", hex.EncodeToString(peerId), tcp.String())
					peer.PublicIpPort = tcp // 写入公网节点
				}
			}
			ckpubconn.Close()
		}()

		break

	}

}
