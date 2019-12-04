package p2p

import (
	"encoding/binary"
	"net"
)

func ParseIPPortToTCPAddrByByte(data []byte) *net.TCPAddr {
	if len(data) != 6 {
		return nil
	}
	return &net.TCPAddr{
		IP:   data[0:4],
		Port: int(binary.BigEndian.Uint16(data[4:6])),
	}
}
