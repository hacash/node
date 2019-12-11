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

func ParseTCPAddrToIPPortBytes(ipport *net.TCPAddr) []byte {
	bts := make([]byte, 6)
	copy(bts[0:4], ipport.IP.To4())
	binary.BigEndian.PutUint16(bts[4:6], uint16(ipport.Port))
	return bts
}
