package p2p

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/interfaces"
	"github.com/hacash/core/sys"
	"github.com/hacash/node/mapset"
	"net"
	"sync"
	"time"
)

type PeerManagerConfig struct {
	PublicPeerGroupMaxLen   int
	InteriorPeerGroupMaxLen int
}

func NewEmptyPeerManagerConfig() *PeerManagerConfig {
	cnf := &PeerManagerConfig{
		PublicPeerGroupMaxLen:   15,
		InteriorPeerGroupMaxLen: 60,
	}
	return cnf
}

func NewPeerManagerConfig(cnffile *sys.Inicnf) *PeerManagerConfig {
	cnf := NewEmptyPeerManagerConfig()

	return cnf
}

////////////////////////////////////////

type PeerManager struct {
	p2p    *P2PManager
	config *PeerManagerConfig

	publicPeerGroup   *PeerGroup
	interiorPeerGroup *PeerGroup

	// manager
	knownPeerIds mapset.Set // set[[]byte] // id.len=32

	waitToConnectNode sync.Map // map[string(target_peer_id)]*net.Addr // local addr

	//currentConnectedPeerIDs           mapset.Set // set[string(id)] len = 16
	//currentConnectedPublicPeerIPPorts mapset.Set // set[string(ipport)] len = 6

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

	activePeerSendPingTiker := time.NewTicker(time.Minute * 5)
	dropNotActivePeerTicker := time.NewTicker(time.Minute * 7)

	for {
		select {

		case <-activePeerSendPingTiker.C:
			curt := time.Now()
			peers := pm.publicPeerGroup.peers.ToSlice()
			peers = append(pm.interiorPeerGroup.peers.ToSlice(), peers...)
			go func() {
				for _, p := range peers {
					peer := p.(*Peer)
					if peer.activeTime.Add(time.Minute * 11).Before(curt) {
						peer.SendMsg(TCPMsgTypePing, nil)
					}
				}
			}()

		case <-dropNotActivePeerTicker.C:
			curt := time.Now()
			peers := pm.publicPeerGroup.peers.ToSlice()
			peers = append(pm.interiorPeerGroup.peers.ToSlice(), peers...)
			go func() {
				for _, p := range peers {
					peer := p.(*Peer)
					if peer.activeTime.Add(time.Minute * 15).Before(curt) {
						pm.DropPeer(peer)
						peer.Close()
					}
				}
			}()
		}

	}
}

// interface api
func (pm *PeerManager) PeerLen() int {
	return pm.publicPeerGroup.peers.Cardinality() + pm.interiorPeerGroup.peers.Cardinality()
}

// interface api
func (pm *PeerManager) FindAnyOnePeerBetterBePublic() interfaces.MsgPeer {
	return pm.FindRandomOnePeerBetterBePublic()
}

func (pm *PeerManager) FindRandomOnePeerBetterBePublic() *Peer {
	var pp = pm.publicPeerGroup
	if pp.peers.Cardinality() == 0 {
		pp = pm.interiorPeerGroup
	}
	// pop
	tarpeer := pp.peers.Pop()
	if tarpeer != nil {
		pp.peers.Add(tarpeer)
	}
	return tarpeer.(*Peer)
}

func (pm *PeerManager) BroadcastMessageToUnawarePeers(ty uint16, msgbody []byte, KnowledgeKey string, KnowledgeValue string) {
	pm.publicPeerGroup.BroadcastMessageToUnawarePeers(ty, msgbody, KnowledgeKey, KnowledgeValue)
	pm.interiorPeerGroup.BroadcastMessageToUnawarePeers(ty, msgbody, KnowledgeKey, KnowledgeValue)
}

func (pm *PeerManager) BroadcastMessageToAllPeers(ty uint16, msgbody []byte) {
	pm.publicPeerGroup.BroadcastMessageToAllPeers(ty, msgbody)
	pm.interiorPeerGroup.BroadcastMessageToAllPeers(ty, msgbody)
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

func (pm *PeerManager) CheckHasConnectedToRemotePublicAddr(tcpaddr *net.TCPAddr) bool {
	tartcpaddr := ParseTCPAddrToIPPortBytes(tcpaddr)
	return pm.CheckHasConnectedToRemotePublicAddrByByte(tartcpaddr)
}

func (pm *PeerManager) CheckHasConnectedToRemotePublicAddrByByte(tartcpaddr []byte) bool {
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
	return pm.BroadcastFindNewNodeMsgToUnawarePublicPeersByBytes(peer.ID, peer.ParseRemotePublicTCPAddress())
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
		if bytes.Compare(p.ID, peerId) != 0 && !p.knownPeerIds.Contains(pidstr) {
			p.AddKnownPeerId(peerId)
			fmt.Println("p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())  to  ", hex.EncodeToString(peerId), p.Name)
			p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())
		}
		return false
	})
	// send interiorPeerGroup
	pm.interiorPeerGroup.peers.Each(func(i interface{}) bool {
		p := i.(*Peer)
		if bytes.Compare(p.ID, peerId) != 0 && !p.knownPeerIds.Contains(pidstr) {
			p.AddKnownPeerId(peerId)
			fmt.Println("p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())  to  ", hex.EncodeToString(peerId), p.Name)
			p.SendMsg(TCPMsgTypeDiscoverPublicPeerJoin, data.Bytes())
		}
		return false
	})
	return nil
}
