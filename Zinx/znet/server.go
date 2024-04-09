package znet

import (
	"Zinx/ziface"
	"errors"
	"fmt"
	"net"
	"time"
)

type Server struct {
	Name string

	IPVersion string

	IP string

	Port int
}

func CallBackToClient(conn *net.TCPConn, data []byte, cnt int) error {
	// 回显业务
	fmt.Println("[Conn Handle] CallBackToClient")
	if _, err := conn.Write(data[:cnt]); err != nil {
		fmt.Println("write back buf err", err)
		return errors.New("CallBackToClient err")
	}
	return nil
}

func (s *Server) Start() {
	fmt.Printf("[Start] server listen at IP %s ,Port %d\n", s.IP, s.Port)

	//开启一个go去做服务端的listen
	go func() {
		// 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		// 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		// 监听成功
		fmt.Println("start server", s.Name, "success ,now listen...")

		var cid uint32
		cid = 0

		// 启动server网络连接业务
		for {
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			dealConn := NewConnection(conn, cid, CallBackToClient)
			cid++

			go dealConn.Start()
		}
	}()
}

func (s *Server) Serve() {
	s.Start()

	for true {
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) Stop() {

}

func NewServer(name string) ziface.IServer {
	s := &Server{
		name,
		"tcp4",
		"127.0.0.1",
		7777,
	}
	return s
}
