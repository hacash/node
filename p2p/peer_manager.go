package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/deckarep/golang-set"
	"sync"
	"time"
)

type PeerManagerConfig struct {
	RelationshipPeerTableMaxLen int
	SequentialPeerTableMaxLen   int
}

func NewPeerManagerConfig() *PeerManagerConfig {
	cnf := &PeerManagerConfig{
		RelationshipPeerTableMaxLen: 12,
		SequentialPeerTableMaxLen:   8,
	}
	return cnf
}

type PeerManager struct {
	p2p    *P2PManager
	config *PeerManagerConfig

	RelationshipPeerIDTable [][]byte
	SequentialPeerIDTable   [][]byte

	peers mapset.Set

	// manager
	knownPeerIds mapset.Set // set[[]byte] // id.len=32

	waitToConnectNode sync.Map // map[string(target_peer_id)]*net.Addr // local addr

	peersChangeLock sync.Mutex
}

func NewPeerManager(cnf *PeerManagerConfig, p2p *P2PManager) *PeerManager {
	pm := &PeerManager{
		p2p:                     p2p,
		config:                  cnf,
		RelationshipPeerIDTable: make([][]byte, 0, cnf.RelationshipPeerTableMaxLen),
		SequentialPeerIDTable:   make([][]byte, 0, cnf.SequentialPeerTableMaxLen),
		peers:                   mapset.NewSet(),
		knownPeerIds:            mapset.NewSet(),
		peersChangeLock:         sync.Mutex{},
	}
	return pm
}

func (pm *PeerManager) Start() {
	go pm.loop()
}

func (pm *PeerManager) loop() {
	aaa := time.NewTicker(time.Minute * 1)
	for {
		select {
		case <-aaa.C:
			//fmt.Println("5 Minute Ticker")
			break
		}
	}
}

func (pm *PeerManager) AddKnownPeerId(pid []byte) {
	pm.knownPeerIds.Add(string(pid))
	if pm.knownPeerIds.Cardinality() > 200 {
		pm.knownPeerIds.Pop() // remove one
	}
}

func (pm *PeerManager) SendFindNewNodeMsgToUnawarePeers(peer *Peer) error {
	pidstr := string(peer.Id)
	if pm.knownPeerIds.Contains(pidstr) {
		return nil // im already known
	}
	pm.AddKnownPeerId(peer.Id)
	// msg body
	data := bytes.NewBuffer(peer.Id)
	// send
	pm.peers.Each(func(i interface{}) bool {
		p := i.(*Peer)
		if !p.knownPeerIds.Contains(pidstr) {
			p.AddKnownPeerId(peer.Id)
			//fmt.Println("p.SendMsg(MsgTypeDiscoverNewNodeJoin, data.Bytes())  to  ", p.Name)
			p.SendMsg(MsgTypeDiscoverNewNodeJoin, data.Bytes())
		}
		return false
	})
	return nil
}

func (pm *PeerManager) DropPeer(peer *Peer) (bool, error) {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	return pm.dropPeerUnsafe(peer)
}

func (pm *PeerManager) dropPeerUnsafe(peer *Peer) (bool, error) {
	if pm.peers.Contains(peer) {
		pm.peers.Remove(peer)
		pm.RelationshipPeerIDTable = deleteBytesFromList(pm.RelationshipPeerIDTable, peer.Id)
		pm.SequentialPeerIDTable = deleteBytesFromList(pm.SequentialPeerIDTable, peer.Id)
		return true, nil
	}
	return false, nil
}

func (pm *PeerManager) dropPeerUnsafeByID(pid []byte) (bool, error) {
	ps := pm.peers.ToSlice()
	for _, p := range ps {
		peer := p.(*Peer)
		if bytes.Compare(peer.Id, pid) == 0 {
			pm.peers.Remove(peer)
			return true, nil
		}
	}
	return false, nil
}

func (pm *PeerManager) addPeerSuccess(peer *Peer) (bool, error) {
	fmt.Println("addPeerSuccess", hex.EncodeToString(peer.Id))
	pm.peers.Add(peer)
	return true, nil
}

func (pm *PeerManager) AddPeer(peer *Peer) (bool, error) {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	// check have
	havp, err := pm.GetPeerByID(peer.Id)
	if err != nil {
		return false, err
	}
	if havp != nil {
		// already have id
		peer.SendMsg(MsgTypeConnectRefuse, nil)
		peer.Close()
		return false, nil
	}

	// add
	if len(pm.SequentialPeerIDTable) < pm.config.SequentialPeerTableMaxLen {
		pm.SequentialPeerIDTable = append(pm.SequentialPeerIDTable, peer.Id)
		return pm.addPeerSuccess(peer)
	}
	// move one
	movePeerID := pm.SequentialPeerIDTable[0]
	pm.SequentialPeerIDTable = pm.SequentialPeerIDTable[1:]
	pm.SequentialPeerIDTable = append(pm.SequentialPeerIDTable, peer.Id)
	// check relationship
	if len(pm.RelationshipPeerIDTable) < pm.config.RelationshipPeerTableMaxLen {
		pm.RelationshipPeerIDTable = InsertIntoRelationshipTable(pm.p2p.selfPeerId, pm.RelationshipPeerIDTable, movePeerID)
		return pm.addPeerSuccess(peer)
	}
	var mustDropPeerId []byte
	pm.RelationshipPeerIDTable, mustDropPeerId =
		UpdateRelationshipTable(pm.p2p.selfPeerId, pm.RelationshipPeerIDTable, pm.config.RelationshipPeerTableMaxLen, movePeerID)
	pm.dropPeerUnsafeByID(mustDropPeerId)
	return pm.addPeerSuccess(peer)
}

func (pm *PeerManager) IsCanAddToRelationshipPeerTable(pid []byte) bool {
	if len(pm.RelationshipPeerIDTable) < pm.config.RelationshipPeerTableMaxLen {
		return true
	}
	_, dropone := UpdateRelationshipTable(pm.p2p.selfPeerId, pm.RelationshipPeerIDTable, pm.config.RelationshipPeerTableMaxLen, pid)
	return bytes.Compare(pid, dropone) != 0
}

func (pm *PeerManager) GetPeerByID(pid []byte) (*Peer, error) {
	if len(pid) != 32 {
		return nil, fmt.Errorf("id len not is 32.")
	}

	ps := pm.peers.ToSlice()
	for _, p := range ps {
		peer := p.(*Peer)
		if bytes.Compare(peer.Id, pid) == 0 {
			return peer, nil
		}
	}
	return nil, nil
}
