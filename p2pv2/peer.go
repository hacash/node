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

	createTime *time.Time // Creation time
	activeTime *time.Time // Last active time

	// Indicates that a replacement connection is occurring
	RemoveImmediately    bool // Removed immediately because the relationship table is full and the topology is too far away
	RemoveActiveOvertime bool // Inactivation
	RemoveFarawayClose   bool // distant
	RemoveReplacing      bool // replace
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
		RemoveReplacing:                    false,
		RemoveFarawayClose:                 false,
		RemoveActiveOvertime:               false,
		RemoveImmediately:                  false,
		knownPeerKnowledgeDuplicateRemoval: new(sync.Map),
	}
}

func (p *Peer) notifyConnect() {
	if p.RemoveImmediately {
		return // Remove nodes immediately without printing or notifying
	}
	pubinfo := ""
	if p.PublicIpPort != nil {
		pubinfo = " (" + p.PublicIpPort.String() + ")"
	}
	if p.RemoveReplacing {
		pubinfo += " connect update"
	} else {
		pubinfo += " id:" + hex.EncodeToString(p.ID) + " connect"
		// External messages
		if p.msghandler != nil {
			p.msghandler.OnConnected(p.communicator, p)
		}
	}
	// Print information
	fmt.Println("[Peer] " + p.Name + pubinfo + ".")
	p.RemoveReplacing = false // reset
}

func (p *Peer) notifyClose() {
	if p.RemoveImmediately {
		return // Remove nodes immediately without printing or notifying
	}
	if p.RemoveReplacing {
		return // For the replaced node, the closing information is not printed, and only the output is printed
	}
	pubinfo := ""
	if p.PublicIpPort != nil {
		pubinfo = " (" + p.PublicIpPort.String() + ")"
	}
	if p.RemoveActiveOvertime {
		pubinfo += " active overtime"
	} else if p.RemoveFarawayClose {
		pubinfo += " topology faraway"
	}
	fmt.Println("[Peer] " + p.Name + pubinfo + " disconnect.")
	// External messages
	if p.msghandler != nil {
		p.msghandler.OnDisconnected(p)
	}
}

func (p *Peer) NameBytes() []byte {
	nbts := bytes.Repeat([]byte{' '}, PeerNameSize)
	copy(nbts, p.Name)
	return nbts
}

// Copy information
func (p *Peer) ReplacingCopyInfo(oldpeer *Peer) {
	p.RemoveReplacing = true
	oldpeer.RemoveReplacing = true
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
