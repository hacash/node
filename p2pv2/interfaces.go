package p2pv2

import (
	"bytes"
	"encoding/binary"
	"github.com/hacash/chain/mapset"
	"github.com/hacash/core/interfaces"
	"strconv"
)

/*******************************************************************************/

/**
 * P2P => P2PManager
 */

func (p *P2P) Start() {
	p.doStart()
}

// 返回 false 为已经知晓
func (p *P2P) AddKnowledge(KnowledgeKind string, KnowledgeKey string) bool {
	return p.peerSelf.AddKnowledge(KnowledgeKind, KnowledgeKey)
}

// 返回 true 为已经知晓
func (p *P2P) CheckKnowledge(KnowledgeKind string, KnowledgeKey string) bool {
	return p.peerSelf.CheckKnowledge(KnowledgeKind, KnowledgeKey)
}

func (p *P2P) SetMsgHandler(msghdr interfaces.P2PMsgDataHandler) {
	p.msgHandler = msghdr
	p.peerSelf.msghandler = msghdr
}

/*******************************************************************************/

/**
 * P2P => P2PMsgCommunicator
 */

func (p *P2P) PeerLen() int {
	return len(p.AllNodes)
}

func (p *P2P) GetAllPeers() []interfaces.P2PMsgPeer {
	nodes := []interfaces.P2PMsgPeer{}
	for _, v := range p.AllNodes {
		nodes = append(nodes, v)
	}
	return nodes
}

func (p *P2P) FindAnyOnePeerBetterBePublic() interfaces.P2PMsgPeer {
	var tarnode interfaces.P2PMsgPeer = nil
	if len(p.BackboneNodeTable) > 0 {
		nd := p.getPeerByID(p.BackboneNodeTable[0])
		if nd != nil {
			tarnode = nd
		}
	}
	if tarnode == nil {
		if len(p.OrdinaryNodeTable) > 0 {
			nd := p.getPeerByID(p.OrdinaryNodeTable[0])
			if nd != nil {
				tarnode = nd
			}
		}
	}
	if tarnode == nil {
		for _, nd := range p.AllNodes {
			tarnode = nd // 随机取一个
			break
		}
	}
	// 返回
	return tarnode
}

func (p *P2P) BroadcastDataMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKind string, KnowledgeKey string) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, ty)
	p.broadcastMessageToUnawarePeers(ty, msgbody, KnowledgeKind, KnowledgeKey)
}

/*******************************************************************************/

/**
 * Peer => P2PMsgPeer
 */

func (p *Peer) Disconnect() {
	p.conn.Close()
}

func (p *Peer) Describe() string {
	des := p.Name
	if p.PublicIpPort != nil {
		des += "(" + p.PublicIpPort.IP.String() + ":" + strconv.Itoa(p.PublicIpPort.Port) + ")"
	}
	return des
}

// 返回 false 为已经知晓
func (p *Peer) AddKnowledge(KnowledgeKind string, KnowledgeKey string) bool {
	knval := mapset.NewSet()
	if actual, ldok := p.knownPeerKnowledgeDuplicateRemoval.LoadOrStore(KnowledgeKind, knval); ldok {
		actknow := actual.(mapset.Set)
		if actknow.Contains(KnowledgeKey) {
			return false // known it
		}
		knval = actknow
	}
	knval.Add(KnowledgeKey)
	if knval.Cardinality() > 120 { // max Knowledge of one key
		knval.Pop() // remove one
	}
	return true
}

func (p *Peer) CheckKnowledge(KnowledgeKind string, KnowledgeKey string) bool {
	if actual, ldok := p.knownPeerKnowledgeDuplicateRemoval.Load(KnowledgeKind); ldok {
		actknow := actual.(mapset.Set)
		if actknow.Contains(KnowledgeKey) {
			return true // known it
		}
	}
	return false //not find
}

func (p *Peer) SendDataMsg(msgty uint16, msgbody []byte) error {
	ty := make([]byte, 2)
	binary.BigEndian.PutUint16(ty, msgty)
	buf := bytes.NewBuffer(ty)
	buf.Write(msgbody)
	return sendTcpMsg(p.conn, P2PMsgTypeCustomer, buf.Bytes())
}
