package znet

import (
	"Zinx/ziface"
	"fmt"
	"net"
)

type Connection struct {
	Conn *net.TCPConn

	ConnID uint32

	isClosed bool

	handleAPI ziface.HandFunc

	ExitBuffChan chan bool
}

func NewConnection(conn *net.TCPConn, connID uint32, callbackApi ziface.HandFunc) *Connection {
	c := &Connection{
		conn,
		connID,
		false,
		callbackApi,
		make(chan bool, 1),
	}
	return c

}

func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running ")
	defer fmt.Println(c.RemoteAddr().String(), "conn reader exit!")
	defer c.Stop()

	for {
		buf := make([]byte, 512)
		cnt, err := c.Conn.Read(buf)

		if err != nil {
			fmt.Println("recv buf err", err)
			c.ExitBuffChan <- true
			continue
		}
		// 调用当前链接业务（这里执行的是当前conn的绑定的handle方法
		if err := c.handleAPI(c.Conn, buf, cnt); err != nil {
			fmt.Println("connID", c.ConnID, "handle is error")
			c.ExitBuffChan <- true
			return
		}
	}
}

func (c *Connection) Start() {
	go c.StartReader()

	for {
		select {
		case <-c.ExitBuffChan:
			// 得到退出消息不再阻塞
			return
		}
	}
}

func (c *Connection) Stop() {
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	c.Conn.Close()

	c.ExitBuffChan <- true

	close(c.ExitBuffChan)
}

func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnId() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}
