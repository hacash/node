package p2pv2

import (
	"encoding/binary"
	"net"
	"time"
)

/**
 * 握手信号连接
 */
func dialTimeoutWithHandshakeSignal(network, address string, timeout time.Duration) (net.Conn, error) {

	// 拨号连接
	conn, e0 := net.DialTimeout(network, address, timeout)
	if e0 != nil {
		return nil, e0
	}

	// 发送握手信号
	e1 := sendTcpMsgHandshakeSignal(conn)
	if e1 != nil {
		conn.Close()
		return nil, e1
	}

	return conn, nil
}

/**
 * 发送握手连接
 */
func sendTcpMsgHandshakeSignal(conn net.Conn) error {

	// 发送握手信号
	hdsksglbts := make([]byte, 4)
	binary.BigEndian.PutUint32(hdsksglbts, P2PHandshakeSignal)
	_, e1 := conn.Write(hdsksglbts)
	if e1 != nil {
		return e1
	}

	return nil
}
