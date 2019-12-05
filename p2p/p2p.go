package p2p

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"sync"
	"time"
)

type P2PManagerConfig struct {
	TCPListenPort       int
	UDPListenPort       int
	LookupConnectMaxLen int
}

func NewP2PManagerConfig() *P2PManagerConfig {
	cnf := &P2PManagerConfig{
		TCPListenPort:       3337,
		UDPListenPort:       3336,
		LookupConnectMaxLen: 128,
	}
	return cnf
}

type P2PManager struct {
	config *P2PManagerConfig

	peerManager *PeerManager
	lookupPeers mapset.Set // []*Peer

	selfPeerName       string
	selfPeerId         []byte // len = 16
	selfRemotePublicIP []byte // is public ?

	//selfRemoteAddr net.Addr

	changeStatusLock sync.Mutex

	waitToReplyIsPublicPeer map[*Peer]struct {
		curt time.Time
		code uint32
	}

	recordOldPublicPeerTCPAddrs    mapset.Set // old public peer addrs set[string(byte(ip_port))]
	recordStaticPublicPeerTCPAddrs mapset.Set // static setting

	// handler
	customerDataHandler P2PMsgDataHandler
}

func NewP2PManager(cnf *P2PManagerConfig, pmcnf *PeerManagerConfig) (*P2PManager, error) {

	p2p := &P2PManager{
		config:             cnf,
		selfRemotePublicIP: nil,
		lookupPeers:        mapset.NewSet(), //make([]*Peer, 0),
		waitToReplyIsPublicPeer: make(map[*Peer]struct {
			curt time.Time
			code uint32
		}, 0),
		recordOldPublicPeerTCPAddrs:    mapset.NewSet(),
		recordStaticPublicPeerTCPAddrs: mapset.NewSet(),
		customerDataHandler:            nil,
	}

	// -------- TEST START --------
	p2p.selfPeerId = make([]byte, 16)
	rand.Read(p2p.selfPeerId) // test
	nnn := []byte(hex.EncodeToString(p2p.selfPeerId))
	p2p.selfPeerName = "hcx_test_peer_" + string(nnn[:8])
	//fmt.Println("im: ", p2p.selfPeerName, string(nnn))
	// -------- TEST END --------

	// pmcnf := &PeerManagerConfig{}
	p2p.peerManager = NewPeerManager(pmcnf, p2p)

	return p2p, nil
}

func (p2p *P2PManager) SetMsgHandler(handler P2PMsgDataHandler) {
	p2p.customerDataHandler = handler
}

func (p2p *P2PManager) Start() {

	go p2p.startListenTCP()
	go p2p.startListenUDP()

	go p2p.loop()

	go p2p.peerManager.Start()

	fmt.Println("[Peer] Start p2p manager id:", hex.EncodeToString(p2p.selfPeerId), "name:", p2p.selfPeerName,
		"listen on port TCP:", p2p.config.TCPListenPort, "UDP:", p2p.config.UDPListenPort)

}

func (p2p *P2PManager) loop() {

	dropUnHandShakeTiker := time.NewTicker(time.Second * 13)
	dropNotReplyPublicTiker := time.NewTicker(time.Second * 8)
	reconnectSomePublicPeerTiker := time.NewTicker(time.Second * 21)
	getAddrsFromPublicPeersTiker := time.NewTicker(time.Minute * 23)

	for {

		select {

		case <-getAddrsFromPublicPeersTiker.C:
			go func() {
				pubpeers := p2p.peerManager.publicPeerGroup.peers.ToSlice()
				for _, p := range pubpeers {
					peer := p.(*Peer)
					peer.SendMsg(TCPMsgTypeGetPublicConnectedPeerAddrs, nil)
					<-time.Tick(time.Second * 7)
				}
			}()

		case <-reconnectSomePublicPeerTiker.C:

			// -------- TEST START --------
			//peers := p2p.peerManager.publicPeerGroup.peers.ToSlice()
			//peers = append(peers, p2p.peerManager.interiorPeerGroup.peers.ToSlice()...)
			//allpnames := ""
			//for _, p := range peers {
			//	allpnames += p.(*Peer).Name + "  "
			//}
			//fmt.Println("current peers num ", len(peers), "(  "+allpnames+")")
			// -------- TEST END   --------

			go func() {
				publicconnCount := p2p.peerManager.publicPeerGroup.peers.Cardinality()
				if publicconnCount < 3 {
					ipport := p2p.recordOldPublicPeerTCPAddrs.Pop()
					if ipport == nil && publicconnCount == 0 {
						ipport = p2p.recordStaticPublicPeerTCPAddrs.Pop()
						if ipport != nil {
							p2p.recordStaticPublicPeerTCPAddrs.Add(ipport) // reput in
						}
					}
					if ipport != nil {
						ipports := []byte(ipport.(string))
						addr := ParseIPPortToTCPAddrByByte(ipports)
						if addr != nil {
							go func() {
								connerr := p2p.TryConnectToPeer(nil, addr)
								if connerr == nil { // reput in
									p2p.AddOldPublicPeerAddrByBytes(ipports)
								}
							}()
						}
					} else {
						//fmt.Println(fmt.Errorf("Cannot get any peer addr to connect."))
					}
				}
			}()

		case <-dropNotReplyPublicTiker.C:
			go func() {
				p2p.changeStatusLock.Lock()
				tnow := time.Now()
				for peer, res := range p2p.waitToReplyIsPublicPeer {
					if res.curt.Add(time.Second * 7).Before(tnow) {
						delete(p2p.waitToReplyIsPublicPeer, peer)
						if peer.publicIPv4 == nil {
							p2p.AddPeerToTargetGroup(p2p.peerManager.interiorPeerGroup, peer)
						}
					}
				}
				p2p.changeStatusLock.Unlock()
			}()

		case <-dropUnHandShakeTiker.C:
			tnow := time.Now()
			peers := p2p.lookupPeers.ToSlice()
			go func() {
				for _, p := range peers {
					peer := p.(*Peer)
					if len(peer.ID) == 0 {
						if peer.connTime.Add(time.Second * 9).Before(tnow) {
							peer.TcpConn.Close() // drop it over time not do hand shake
						}
					}
				}
			}()

		}

	}

}
