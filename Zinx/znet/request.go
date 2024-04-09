package znet

import "Zinx/ziface"

type Request struct {
	conn ziface.IConnection
	data []byte
}

func (r *Request) GetConnect() ziface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.data
}
