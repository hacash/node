package p2pv2

import (
	"bytes"
	"net"
)

const FindNodeSize = 4 + 2 + 16

type fdNodes struct {
	TcpAddr *net.TCPAddr
	ID      PeerID
}

func parseFindNodesFromBytes(datas []byte) []*fdNodes {
	var nodes = []*fdNodes{}
	if len(datas)%FindNodeSize != 0 {
		return nodes
	}
	for i := 0; i < len(datas); i += FindNodeSize {
		node := &fdNodes{
			ID: PeerID(datas[i+6 : i+FindNodeSize]),
		}
		node.TcpAddr = parseIPPortToTCPAddrByByte(datas[i : i+6])
		nodes = append(nodes, node)
	}
	return nodes
}

func serializeFindNodesToBytes(nodes []*fdNodes) []byte {
	buf := bytes.NewBuffer(nil)
	for _, v := range nodes {
		ippbts := parseTCPAddrToIPPortBytes(v.TcpAddr)
		buf.Write(ippbts)
		buf.Write(v.ID)
	}
	return buf.Bytes()
}
