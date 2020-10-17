package p2pv2

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/hacash/core/interfaces"
	"net"
	"sync"
	"time"
)

type PeerID []byte

type Peer struct {
	communicator interfaces.P2PMsgCommunicator
	msghandler   interfaces.P2PMsgDataHandler

	connid uint64
	conn   net.Conn

	ID   PeerID
	Name string

	PublicIpPort *net.TCPAddr

	knownPeerKnowledgeDuplicateRemoval *sync.Map // map[string]set[string(byte)]

	createTime *time.Time // 创建时间
	activeTime *time.Time // 上次活跃时间

	// 表示正在发生替换的连接
	ActiveOvertime bool // 失活
	FarawayClose   bool // 遥远
	Replacing      bool
}

func NewEmptyPeer(cmtr interfaces.P2PMsgCommunicator, msgdlr interfaces.P2PMsgDataHandler) *Peer {
	ct := time.Now()
	return &Peer{
		communicator:                       cmtr,
		msghandler:                         msgdlr,
		connid:                             0,
		conn:                               nil,
		ID:                                 nil,
		Name:                               "",
		PublicIpPort:                       nil,
		createTime:                         &ct,
		activeTime:                         &ct,
		Replacing:                          false,
		ActiveOvertime:                     false,
		knownPeerKnowledgeDuplicateRemoval: new(sync.Map),
	}
}

func (p *Peer) notifyConnect() {
	pubinfo := ""
	if p.PublicIpPort != nil {
		pubinfo = " (" + p.PublicIpPort.String() + ")"
	}
	if p.Replacing {
		pubinfo += " connect update"
	} else {
		pubinfo += " id:" + hex.EncodeToString(p.ID) + " connect"
		// 外部消息
		if p.msghandler != nil {
			p.msghandler.OnConnected(p.communicator, p)
		}
	}
	// 打印信息
	fmt.Println("[Peer] " + p.Name + pubinfo + ".")
	p.Replacing = false // reset
}

func (p *Peer) notifyClose() {
	if p.Replacing {
		return // 不打印信息
	}
	pubinfo := ""
	if p.PublicIpPort != nil {
		pubinfo = " (" + p.PublicIpPort.String() + ")"
	}
	if p.ActiveOvertime {
		pubinfo += " active overtime"
	} else if p.FarawayClose {
		pubinfo += " faraway"
	}
	fmt.Println("[Peer] " + p.Name + pubinfo + " disconnect.")
	// 外部消息
	if p.msghandler != nil {
		p.msghandler.OnDisconnected(p)
	}
}

func (p *Peer) NameBytes() []byte {
	nbts := bytes.Repeat([]byte{' '}, PeerNameSize)
	copy(nbts, p.Name)
	return nbts
}

// 拷贝信息
func (p *Peer) ReplacingCopyInfo(oldpeer *Peer) {
	p.Replacing = true
	oldpeer.Replacing = true
	p.knownPeerKnowledgeDuplicateRemoval = oldpeer.knownPeerKnowledgeDuplicateRemoval
}

func (p *Peer) SendUnawareMsg(ty uint16, msgbody []byte, KnowledgeKind string, KnowledgeKey string) error {
	//fmt.Println("SendUnawareMsg:", KnowledgeKey, KnowledgeValue)
	if p.AddKnowledge(KnowledgeKind, KnowledgeKey) {
		// add success and send data
		//fmt.Println("SendUnawareMsg:", p.Name, ", datalen:", len(msgbody))
		return p.SendDataMsg(ty, msgbody)
	}
	return nil
}
