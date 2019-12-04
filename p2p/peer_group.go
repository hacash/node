package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"sync"
)

type PeerGroup struct {
	mySelfID []byte

	relationshipPeerIDTable  [][]byte
	relationshipTableMaxSize int

	sequentialPeerIDTable  [][]byte
	sequentialTableMaxSize int

	peers mapset.Set

	peersChangeLock sync.Mutex
}

func NewPeerGroup(selfpid []byte, rptl, sptl int) *PeerGroup {
	return &PeerGroup{
		mySelfID:                 selfpid,
		peers:                    mapset.NewSet(),
		relationshipTableMaxSize: rptl,
		sequentialTableMaxSize:   sptl,
		relationshipPeerIDTable:  make([][]byte, 0, rptl+1),
		sequentialPeerIDTable:    make([][]byte, 0, sptl+1),
	}
}

func (pm *PeerGroup) IsUnfilled() bool {
	if pm.peers.Cardinality() < pm.relationshipTableMaxSize+pm.sequentialTableMaxSize {
		return true
	}
	return false
}

func (pm *PeerGroup) DropPeer(peer *Peer) (bool, error) {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	return pm.dropPeerUnsafe(peer)
}

func (pm *PeerGroup) dropPeerUnsafe(peer *Peer) (bool, error) {
	if pm.peers.Contains(peer) {
		pm.peers.Remove(peer)
		pm.relationshipPeerIDTable = deleteBytesFromList(pm.relationshipPeerIDTable, peer.Id)
		pm.sequentialPeerIDTable = deleteBytesFromList(pm.sequentialPeerIDTable, peer.Id)
		return true, nil
	}
	return false, nil
}

func (pm *PeerGroup) dropPeerUnsafeByID(pid []byte) bool {
	if len(pid) != 16 {
		return false
	}
	ps := pm.peers.ToSlice()
	for _, p := range ps {
		peer := p.(*Peer)
		if bytes.Compare(peer.Id, pid) == 0 {
			pm.peers.Remove(peer)
			return true
		}
	}
	return false
}

func (pm *PeerGroup) addPeerSuccess(peer *Peer) (bool, error) {
	//fmt.Println("addPeerSuccess", hex.EncodeToString(peer.Id))
	publicstr := ""
	if peer.publicIPv4 != nil {
		publicstr = "@public "
	}
	fmt.Println("[Peer] successfully connected "+publicstr+"peer id:", hex.EncodeToString(peer.Id), "name:", peer.Name, "addr:", peer.TcpConn.RemoteAddr().String())

	peer.IsPermitCompleteNode = true
	pm.peers.Add(peer)
	return true, nil
}

func (pm *PeerGroup) AddPeer(peer *Peer) (bool, error) {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	if peer.TcpConn == nil {
		return false, fmt.Errorf("peer connect is closed.")
	}

	// add
	if len(pm.sequentialPeerIDTable) < pm.sequentialTableMaxSize {
		pm.sequentialPeerIDTable = append(pm.sequentialPeerIDTable, peer.Id)
		return pm.addPeerSuccess(peer)
	}
	// move one
	movePeerID := pm.sequentialPeerIDTable[0]
	pm.sequentialPeerIDTable = pm.sequentialPeerIDTable[1:]
	pm.sequentialPeerIDTable = append(pm.sequentialPeerIDTable, peer.Id)
	// check relationship
	if len(pm.relationshipPeerIDTable) < pm.relationshipTableMaxSize {
		pm.relationshipPeerIDTable = InsertIntoRelationshipTable(pm.mySelfID, pm.relationshipPeerIDTable, movePeerID)
		return pm.addPeerSuccess(peer)
	}
	var mustDropPeerId []byte
	pm.relationshipPeerIDTable, mustDropPeerId =
		UpdateRelationshipTable(pm.mySelfID, pm.relationshipPeerIDTable, pm.relationshipTableMaxSize, movePeerID)
	pm.dropPeerUnsafeByID(mustDropPeerId)
	return pm.addPeerSuccess(peer)
}

func (pm *PeerGroup) IsCanAddToRelationshipPeerTable(tarpid []byte) bool {
	if len(pm.relationshipPeerIDTable) < pm.relationshipTableMaxSize {
		return true
	}
	_, dropone := UpdateRelationshipTable(pm.mySelfID, pm.relationshipPeerIDTable, pm.relationshipTableMaxSize, tarpid)
	return bytes.Compare(tarpid, dropone) != 0
}

func (pm *PeerGroup) GetPeerByID(pid []byte) (*Peer, error) {
	if len(pid) != 16 {
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
