package p2pv2

func (p *P2P) getPeerByID(pid PeerID) *Peer {
	v, ok := p.AllNodes.Load(string(pid))
	if ok {
		return v.(*Peer)
	}
	return nil
}

// 加入指定的节点表
func (p *P2P) addPeerIntoTargetTable(tables *[]PeerID, maxsize int, peer *Peer) {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.addPeerIntoTargetTableUnsafe(tables, maxsize, peer)
}

// 加入指定的节点表
func (p *P2P) addPeerIntoTargetTableUnsafe(tables *[]PeerID, maxsize int, peer *Peer) {

	// 插入
	istok, _, newtables, dropid := insertUpdateTopologyDistancePeerIDTable(p.peerSelf.ID, peer.ID, *tables, maxsize)
	if istok {
		*tables = newtables // 更新
	}
	// 移除多余的
	if dropid != nil {
		droppeer := p.getPeerByID(dropid)
		if droppeer != nil {
			droppeer.FarawayClose = true
			droppeer.Disconnect()
		}
	}
}

// 加入陌生节点
func (p *P2P) addPeerIntoUnfamiliarTable(peer *Peer) {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.addPeerIntoUnfamiliarTableUnsafe(peer)
}

// 提升等级
func (p *P2P) upgradeOneUnfamiliarNodeLevel() {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.upgradeOneUnfamiliarNodeLevelUnsafe()
}

// 提升等级
func (p *P2P) upgradeOneUnfamiliarNodeLevelUnsafe() {
	if len(p.UnfamiliarNodesTable) > 0 {
		// 删除第一个最早的
		olderone := p.UnfamiliarNodesTable[0]
		p.UnfamiliarNodesTable = p.UnfamiliarNodesTable[1:]
		// 是否为公网节点
		droppeer := p.getPeerByID(olderone)
		if droppeer.PublicIpPort != nil {
			p.addPeerIntoTargetTableUnsafe(&p.BackboneNodeTable, p.Config.BackboneNodeTableSizeMax, droppeer) // 公网节点
		} else {
			p.addPeerIntoTargetTableUnsafe(&p.OrdinaryNodeTable, p.Config.OrdinaryNodeTableSizeMax, droppeer) // 普通节点
		}
	}
}

// 加入陌生节点
func (p *P2P) addPeerIntoUnfamiliarTableUnsafe(peer *Peer) {
	// 加入
	p.UnfamiliarNodesTable = append(p.UnfamiliarNodesTable, peer.ID)
	// 判断满员
	if len(p.UnfamiliarNodesTable) > p.Config.UnfamiliarNodeTableSizeMax {
		p.upgradeOneUnfamiliarNodeLevelUnsafe() // 提升一个节点的等级
	}
}

func (p *P2P) dropPeerByConnID(cid uint64) {
	p.PeerChangeMux.Lock()
	defer p.PeerChangeMux.Unlock()

	p.dropPeerByConnIDUnsafe(cid)
}

func (p *P2P) dropPeerByConnIDUnsafe(cid uint64) {
	var peer *Peer = nil
	// 查询
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
	// 移除
	if peer != nil {
		pid := peer.ID
		if rmok, newtbs := removePeerIDFromTable(p.BackboneNodeTable, pid); rmok {
			p.BackboneNodeTable = newtbs // 更新表
		}
		if rmok, newtbs := removePeerIDFromTable(p.OrdinaryNodeTable, pid); rmok {
			p.OrdinaryNodeTable = newtbs // 更新表
		}
		if rmok, newtbs := removePeerIDFromTable(p.UnfamiliarNodesTable, pid); rmok {
			p.UnfamiliarNodesTable = newtbs // 更新表
		}
		p.AllNodes.Delete(string(pid))
		p.AllNodesLen -= 1
		peer.notifyClose()
	}
}
