package p2p

import (
	"encoding/binary"
	"net"
	"sync"
	"time"
)

type Peer struct {
	Name string
	Id   []byte // len 32

	PublicListenTcpAddr net.Addr // public tcp listen addr

	TcpConn net.Conn

	tcpListenPort int
	udpListenPort int

	knownPeerKnowledgeDuplicateRemoval sync.Map // map[string]set[string(byte)]

	activeTime time.Time // check live
}

func NewPeer(id []byte, name string) *Peer {
	return &Peer{
		Id:                  id,
		Name:                name,
		PublicListenTcpAddr: nil,
		tcpListenPort:       0,
		udpListenPort:       0,
		activeTime:          time.Now(),
	}
}

func (p *Peer) Close() {
	if p.TcpConn != nil {
		p.TcpConn.Close()
		p.TcpConn = nil
	}
}

func (p *Peer) SendMsg(ty uint16, msgbody []byte) error {
	if msgbody == nil {
		msgbody = []byte{}
	}
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, ty)
	var dtlen []byte = nil
	if msgbody != nil {
		if ty == MsgTypeData {
			dtlen = make([]byte, 4)
			binary.BigEndian.PutUint32(dtlen, uint32(len(msgbody)))
		} else {
			dtlen = make([]byte, 2)
			binary.BigEndian.PutUint16(dtlen, uint16(len(msgbody)))
		}
	}
	data = append(data, dtlen...)
	data = append(data, msgbody...)

	// send data
	if p.TcpConn != nil {
		_, e := p.TcpConn.Write(data)
		if e != nil {
			return e
		}
	}
	return nil

}
