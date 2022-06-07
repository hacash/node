package p2pv2

import "bytes"

func (p *P2P) getPeerByID(pid PeerID) *Peer {
	v, ok := p.AllNodes.Load(string(pid))
	if ok {
		return v.(*Peer)
	}
	return nil
}

// Join the specified node table
func (p *P2P) addPeerIntoTargetTable(tables *[]PeerID, maxsize int, peer *Peer) {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.addPeerIntoTargetTableUnsafe(tables, maxsize, peer)
}

// Join the specified node table
func (p *P2P) addPeerIntoTargetTableUnsafe(tables *[]PeerID, maxsize int, peer *Peer) bool {
	var addSuccessfully = true // Whether it was added successfully, but not removed immediately
	// insert
	istok, _, newtables, dropid := insertUpdateTopologyDistancePeerIDTable(p.peerSelf.ID, peer.ID, *tables, maxsize)

	if istok {
		*tables = newtables // to update
	}

	// Remove excess
	if dropid != nil {
		droppeer := p.getPeerByID(dropid)
		if droppeer != nil {
			droppeer.RemoveFarawayClose = true // Topology too far
			if bytes.Compare(droppeer.ID, peer.ID) == 0 {
				// The topology is too far away. It was removed immediately
				droppeer.RemoveImmediately = true //立即移除
				addSuccessfully = false           // Add failed
			}
			droppeer.Disconnect()
		}
	}
	return addSuccessfully
}

// Join a strange node
func (p *P2P) addPeerIntoUnfamiliarTable(peer *Peer) {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.addPeerIntoUnfamiliarTableUnsafe(peer)
}

// Upgrade level
func (p *P2P) upgradeOneUnfamiliarNodeLevel() {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.upgradeOneUnfamiliarNodeLevelUnsafe()
}

// Upgrade level
func (p *P2P) upgradeOneUnfamiliarNodeLevelUnsafe() {
	if len(p.UnfamiliarNodesTable) > 0 {
		// Delete first oldest
		olderone := p.UnfamiliarNodesTable[0]
		p.UnfamiliarNodesTable = p.UnfamiliarNodesTable[1:]
		// Whether it is a public network node
		droppeer := p.getPeerByID(olderone)
		if droppeer == nil {
			return // return
		}
		if droppeer.PublicIpPort != nil {
			p.addPeerIntoTargetTableUnsafe(&p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax, droppeer) // Public network node
		} else {
			p.addPeerIntoTargetTableUnsafe(&p.OrdinaryNodeTable, p.Config.OrdinaryNodeTableSizeMax, droppeer) // Common node
		}
	}
}

// Join a strange node
func (p *P2P) addPeerIntoUnfamiliarTableUnsafe(peer *Peer) {
	// join
	p.UnfamiliarNodesTable = append(p.UnfamiliarNodesTable, peer.ID)
	// Judge full
	if len(p.UnfamiliarNodesTable) > p.Config.UnfamiliarNodeTableSizeMax {
		p.upgradeOneUnfamiliarNodeLevelUnsafe() // Raise the level of a node
	}
}

//
func (p *P2P) addPeerAllNodesUnsafe(peer *Peer) {
	// join
	peerId := peer.ID
	_, have := p.AllNodes.LoadOrStore(string(peerId), peer)
	if !have {
		p.AllNodesLen += 1
	}
}

func (p *P2P) dropPeerByConnID(cid uint64) {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.dropPeerByConnIDUnsafe(cid)
}

func (p *P2P) dropPeerByConnIDUnsafe(cid uint64) {
	var peer *Peer = nil
	// query
	p.AllNodes.Range(func(key, value interface{}) bool {
		p := value.(*Peer)
		if p != nil {
			if p.connid == cid {
				peer = p
				return false
			}
		}
		return true
	})
	// remove
	if peer != nil {
		pid := peer.ID
		if rmok, newtbs := removePeerIDFromTable(p.BackboneNodeTable, pid); rmok {
			p.BackboneNodeTable = newtbs // Update table
		}
		if rmok, newtbs := removePeerIDFromTable(p.OrdinaryNodeTable, pid); rmok {
			p.OrdinaryNodeTable = newtbs // Update table
		}
		if rmok, newtbs := removePeerIDFromTable(p.UnfamiliarNodesTable, pid); rmok {
			p.UnfamiliarNodesTable = newtbs // Update table
		}
		p.AllNodes.Delete(string(pid))
		p.AllNodesLen -= 1
		peer.notifyClose()
	}
}
