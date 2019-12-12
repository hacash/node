package p2p

import (
	"encoding/binary"
	"net"
)

func ParseIPToString(p4 []byte) string {

	// If IPv4, use dotted notation.
	if len(p4) != 4 {
		return "_._._._"
	}
	b := make([]byte, 16)
	n := ubtoa(b, 0, p4[0])
	b[n] = '.'
	n++
	n += ubtoa(b, n, p4[1])
	b[n] = '.'
	n++
	n += ubtoa(b, n, p4[2])
	b[n] = '.'
	n++
	n += ubtoa(b, n, p4[3])
	return string(b[:n])

}

func ubtoa(dst []byte, start int, v byte) int {
	if v < 10 {
		dst[start] = v + '0'
		return 1
	} else if v < 100 {
		dst[start+1] = v%10 + '0'
		dst[start] = v/10 + '0'
		return 2
	}

	dst[start+2] = v%10 + '0'
	dst[start+1] = (v/10)%10 + '0'
	dst[start] = v/100 + '0'
	return 3
}

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
