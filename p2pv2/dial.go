package p2pv2

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

/**
 * 握手信号连接
 */
func dialTimeoutWithHandshakeSignal(network, address string, timeout time.Duration) (net.Conn, error) {

	// 拨号连接
	//fmt.Println("DialTimeout", address, timeout)
	conn, e0 := net.DialTimeout(network, address, timeout)
	//fmt.Println("DialTimeout", address, "return", conn, e0)
	if e0 != nil {
		return nil, e0
	}

	// 执行握手
	e1 := doTcpMsgHandshakeSignalIfErrorClose(conn, time.Second*10)
	if e1 != nil {
		return nil, e1
	}

	return conn, nil
}

/**
 * 发送和接受握手连接
 */
func doTcpMsgHandshakeSignalIfErrorClose(conn net.Conn, timeout time.Duration) error {
	//return nil

	timein := time.NewTimer(timeout)
	reterrCh := make(chan error, 1)

	go func() {
		// 发送握手信号
		e1 := sendTcpMsgHandshakeSignal(conn)
		if e1 != nil {
			conn.Close()
			reterrCh <- e1
			return
		}

		// 收到握手信号
		hdsksgl := make([]byte, 4)
		_, e2 := io.ReadFull(conn, hdsksgl)
		if e2 != nil {
			conn.Close()
			reterrCh <- e2
			return
		}
		signal := binary.BigEndian.Uint32(hdsksgl)
		if signal != P2PHandshakeSignal {
			conn.Close()
			reterrCh <- fmt.Errorf("p2p handshake fail")
			return
		}

		// ok
		//fmt.Println("reterrCh <- nil")
		reterrCh <- nil
		return
	}()

	// return
	select {
	case <-timein.C:
		conn.Close() // close time conn
		return fmt.Errorf("sendTcpMsgHandshakeSignal timeout")
	case err := <-reterrCh:
		timein.Stop()
		return err
	}

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
