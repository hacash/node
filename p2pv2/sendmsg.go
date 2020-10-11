package p2pv2

import (
	"bytes"
	"encoding/binary"
	"net"
)

func sendTcpMsg(conn net.Conn, msgty uint8, body []byte) error {
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(1+len(body)))
	buf := bytes.NewBuffer(size)
	buf.Write([]byte{msgty})
	buf.Write(body)
	//
	data := buf.Bytes()
	//fmt.Println("sendTcpMsg(): ", len(data), data)
	_, e0 := conn.Write(data)
	return e0
}
