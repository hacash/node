package p2pv2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

func sendTcpMsg(conn net.Conn, msgty uint8, body []byte) error {
	if conn == nil {
		return fmt.Errorf("conn is nil")
	}
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(1+len(body)))
	buf := bytes.NewBuffer(size)
	buf.Write([]byte{msgty})
	buf.Write(body)
	//
	data := buf.Bytes()
	// fmt.Println("sendTcpMsg(): ", conn.RemoteAddr().String(), ":", len(data), data)
	_, e0 := conn.Write(data)
	return e0
}
