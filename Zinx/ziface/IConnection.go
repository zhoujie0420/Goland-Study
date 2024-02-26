package ziface

import "net"

type IConnection interface {
	Start()

	Stop()

	GetTCPConnection() *net.TCPConn

	GetConnId() uint32

	RemoteAddr() net.Addr
}

// HandFunc 统一处理链接业务的接口
type HandFunc func(*net.TCPConn, []byte, int) error
