package p2p

import (
	"bytes"
	"encoding/binary"
	mapset "github.com/deckarep/golang-set"
	"net"
)

func (p2p *P2PManager) GetPeerByID(peerID []byte) *Peer {
	if len(peerID) != 16 {
		return nil
	}
	peer := p2p.peerManager.GetPeerByID(peerID)
	if peer != nil {
		return peer
	}
	peers := p2p.lookupPeers.ToSlice()
	for _, p := range peers {
		peer := p.(*Peer)
		if bytes.Compare(peerID, peer.ID) == 0 {
			return peer
		}
	}
	return nil
}

func (p2p *P2PManager) AddPeerToTargetGroup(group *PeerGroup, peer *Peer) error {

	_, err := group.AddPeer(peer)
	if err != nil {
		return err
	}
	p2p.lookupPeers.Remove(peer)
	if p2p.customerDataHandler != nil {
		p2p.customerDataHandler.OnConnected(p2p.peerManager, peer)
	}
	return nil

}

func (p2p *P2PManager) AddOldPublicPeerAddrByBytes(ipport []byte) error {
	return addAddrToMapsetWithMaxSizeByBytes(p2p.recordOldPublicPeerTCPAddrs, 500, ipport)
}

func (p2p *P2PManager) AddOldPublicPeerAddr(ip net.IP, port int) {
	addAddrToMapsetWithMaxSize(p2p.recordOldPublicPeerTCPAddrs, 500, ip, port)
}

func (p2p *P2PManager) AddStaticPublicPeerAddrByString(address string) error {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return err
	}
	addAddrToMapsetWithMaxSize(p2p.recordStaticPublicPeerTCPAddrs, 200, addr.IP, addr.Port)
	return nil
}

///////////////////////////////////////////////////

func addAddrToMapsetWithMaxSize(set mapset.Set, maxlen int, ip net.IP, port int) error {
	ipport := make([]byte, 6)
	copy(ipport[0:4], ip)
	binary.BigEndian.PutUint16(ipport[4:6], uint16(port))
	return addAddrToMapsetWithMaxSizeByBytes(set, maxlen, ipport)
}

func addAddrToMapsetWithMaxSizeByBytes(set mapset.Set, maxlen int, ipport []byte) error {
	set.Add(string(ipport))
	// check size
	if set.Cardinality() > maxlen {
		set.Pop() // MAX LEN == 500
	}
	return nil
}
