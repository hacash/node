package p2p

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"net"
	"sync"
	"time"
)

type PeerManagerConfig struct {
	PublicPeerGroupMaxLen   int
	InteriorPeerGroupMaxLen int
	LookupConnectMaxLen     int
}

func NewPeerManagerConfig() *PeerManagerConfig {
	cnf := &PeerManagerConfig{
		PublicPeerGroupMaxLen:   15,
		InteriorPeerGroupMaxLen: 60,
	}
	return cnf
}

type PeerManager struct {
	p2p    *P2PManager
	config *PeerManagerConfig

	publicPeerGroup   *PeerGroup
	interiorPeerGroup *PeerGroup

	// manager
	knownPeerIds mapset.Set // set[[]byte] // id.len=32

	waitToConnectNode sync.Map // map[string(target_peer_id)]*net.Addr // local addr

	peersChangeLock sync.Mutex
}

func NewPeerManager(cnf *PeerManagerConfig, p2p *P2PManager) *PeerManager {
	if cnf.PublicPeerGroupMaxLen < 3 || cnf.InteriorPeerGroupMaxLen < 3 {
		panic("PublicPeerGroupMaxLen or InteriorPeerGroupMaxLen cannot less than 3.")
	}
	ppgmlr := cnf.PublicPeerGroupMaxLen / 3 * 2
	ppgmls := cnf.PublicPeerGroupMaxLen - ppgmlr
	ipgmlr := cnf.InteriorPeerGroupMaxLen / 3 * 2
	ipgmls := cnf.InteriorPeerGroupMaxLen - ipgmlr
	pm := &PeerManager{
		p2p:               p2p,
		config:            cnf,
		publicPeerGroup:   NewPeerGroup(p2p.selfPeerId, ppgmlr, ppgmls),
		interiorPeerGroup: NewPeerGroup(p2p.selfPeerId, ipgmlr, ipgmls),
		knownPeerIds:      mapset.NewSet(),
		peersChangeLock:   sync.Mutex{},
	}
	return pm
}

func (pm *PeerManager) Start() {
	go pm.loop()
}

func (pm *PeerManager) loop() {

	activePeerSendPingTiker := time.NewTicker(time.Minute * 15)
	dropNotActivePeerTicker := time.NewTicker(time.Minute * 20)

	for {
		select {

		case <-dropNotActivePeerTicker.C:
			curt := time.Now()
			peers := pm.publicPeerGroup.peers.ToSlice()
			peers = append(pm.publicPeerGroup.peers.ToSlice(), peers...)
			go func() {
				for _, p := range peers {
					peer := p.(*Peer)
					if peer.activeTime.Add(time.Minute * 25).Before(curt) {
						pm.DropPeer(peer)
						peer.Close()
					}
				}
			}()

		case <-activePeerSendPingTiker.C:
			curt := time.Now()
			peers := pm.publicPeerGroup.peers.ToSlice()
			peers = append(pm.publicPeerGroup.peers.ToSlice(), peers...)
			go func() {
				for _, p := range peers {
					peer := p.(*Peer)
					if peer.activeTime.Add(time.Minute * 17).Before(curt) {
						peer.SendMsg(TCPMsgTypePing, nil)
					}
				}
			}()
		}
	}
}

func (pm *PeerManager) DropPeer(peer *Peer) error {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	_, err := pm.publicPeerGroup.dropPeerUnsafe(peer)
	if err != nil {
		return err
	}
	_, err = pm.interiorPeerGroup.dropPeerUnsafe(peer)
	if err != nil {
		return err
	}
	return nil
}

func (pm *PeerManager) dropPeerUnsafeByID(pid []byte) error {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	pm.publicPeerGroup.dropPeerUnsafeByID(pid)
	pm.interiorPeerGroup.dropPeerUnsafeByID(pid)
	return nil
}

func (pm *PeerManager) AddPeer(peer *Peer) error {
	pm.peersChangeLock.Lock()
	defer pm.peersChangeLock.Unlock()

	if peer.publicIPv4 == nil {
		pm.interiorPeerGroup.AddPeer(peer)
	} else {
		pm.publicPeerGroup.AddPeer(peer)
	}
	return nil
}

func (pm *PeerManager) GetPeerByID(pid []byte) *Peer {

	peer, _ := pm.publicPeerGroup.GetPeerByID(pid)
	if peer != nil {
		return peer
	}
	peer, _ = pm.interiorPeerGroup.GetPeerByID(pid)
	if peer != nil {
		return peer
	}
	return nil // not find
}

func (pm *PeerManager) AddKnownPeerId(pid []byte) {
	pm.knownPeerIds.Add(string(pid))
	if pm.knownPeerIds.Cardinality() > 200 {
		pm.knownPeerIds.Pop() // remove one
	}
}

func (pm *PeerManager) GetAllCurrentConnectedPublicPeerAddressBytes() []byte {
	peers := pm.publicPeerGroup.peers.ToSlice()
	addrsbytes := make([]byte, 6*len(peers))
	for i, p := range peers {
		peer := p.(*Peer)
		addrbts := peer.ParseRemotePublicTCPAddress()
		copy(addrsbytes[i*6:i*6+6], addrbts)
	}
	return addrsbytes
}

func (pm *PeerManager) CheckHasConnectedWithRemotePublicAddr(tcpaddr *net.TCPAddr) bool {

	tartcpaddr := make([]byte, 6)
	copy(tartcpaddr[0:4], tcpaddr.IP.To4())
	binary.BigEndian.PutUint16(tartcpaddr[4:6], uint16(tcpaddr.Port))
	//
	peers := pm.publicPeerGroup.peers.ToSlice()
	for _, p := range peers {
		pipport := p.(*Peer).ParseRemotePublicTCPAddress()
		if pipport != nil {
			if bytes.Compare(tartcpaddr, pipport) == 0 {
				return true
			}
		}
	}
	return false
}

func (pm *PeerManager) BroadcastFindNewNodeMsgToUnawarePublicPeers(peer *Peer) error {
	//fmt.Println("BroadcastFindNewNodeMsgToUnawarePublicPeers", peer.ParseRemotePublicTCPAddress())
	return pm.BroadcastFindNewNodeMsgToUnawarePublicPeersByBytes(peer.Id, peer.ParseRemotePublicTCPAddress())
}

func (pm *PeerManager) BroadcastFindNewNodeMsgToUnawarePublicPeersByBytes(peerId []byte, ipport []byte) error {
	if len(peerId) != 16 || len(ipport) != 6 {
		return fmt.Errorf("data len error.")
	}
	pidstr := string(peerId)
	if pm.knownPeerIds.Contains(pidstr) {
		return nil // im already known
	}
	pm.AddKnownPeerId(peerId)
	// msg body
	data := bytes.NewBuffer(peerId) // len + 16
	data.Write(ipport)              // len + 6
	// send publicPeerGroup
	pm.publicPeerGroup.peers.Each(func(i interface{}) bool {
		p := i.(*Peer)
		if bytes.Compare(p.Id, peerId) != 0 && !p.knownPeerIds.Contains(pidstr) {
			p.AddKnownPeerId(peerId)
			fmt.Println("p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())  to  ", hex.EncodeToString(peerId), p.Name)
			p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())
		}
		return false
	})
	// send interiorPeerGroup
	pm.interiorPeerGroup.peers.Each(func(i interface{}) bool {
		p := i.(*Peer)
		if bytes.Compare(p.Id, peerId) != 0 && !p.knownPeerIds.Contains(pidstr) {
			p.AddKnownPeerId(peerId)
			fmt.Println("p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())  to  ", hex.EncodeToString(peerId), p.Name)
			p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())
		}
		return false
	})
	return nil
}
