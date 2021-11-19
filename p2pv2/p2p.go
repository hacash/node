package p2pv2

import (
	"encoding/hex"
	"github.com/hacash/core/interfaces"
	"sync"
)

type P2P struct {
	Config *P2PConfig

	msgHandler interfaces.P2PMsgDataHandler

	BackboneNodeTable    []PeerID // 公网节点
	OrdinaryNodeTable    []PeerID // 私网节点
	UnfamiliarNodesTable []PeerID // 临时节点

	AllNodesLen int
	AllNodes    sync.Map //       [string]*Peer // 全部节点池

	TemporaryConnsLen int
	TemporaryConns    sync.Map // [uint64]net.Conn // 临时连接池

	PeerChangeMux sync.RWMutex

	// my peer
	peerSelf           *Peer
	MyselfIsPublicPeer bool // 我自己是不是公网节点

	// 状态
	isInFindingNode uint32
}

func NewP2P(cnf *P2PConfig) *P2P {

	p2pobj := &P2P{
		Config:               cnf,
		BackboneNodeTable:    []PeerID{},
		OrdinaryNodeTable:    []PeerID{},
		UnfamiliarNodesTable: []PeerID{},
		AllNodesLen:          0,
		TemporaryConnsLen:    0,
		PeerChangeMux:        sync.RWMutex{},
		msgHandler:           nil,
		MyselfIsPublicPeer:   false,
		isInFindingNode:      0,
	}

	var peerSelf = NewEmptyPeer(p2pobj, p2pobj.msgHandler)
	peerSelf.ID = cnf.ID
	if len(cnf.Name) == 0 {
		cnf.Name = "hn_" + string([]byte(hex.EncodeToString(cnf.ID))[0:10])
	}
	peerSelf.Name = cnf.Name

	// set myself peer
	p2pobj.peerSelf = peerSelf

	return p2pobj
}

func (p *P2P) doStart() error {

	go p.loop()

	go p.tryConnectToStaticBootNodes()

	return p.listen(p.Config.TCPListenPort)
}
