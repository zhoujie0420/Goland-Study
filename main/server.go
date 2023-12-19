package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

func (this *Server) Handler(accept net.Conn) {

	// ... 业务逻辑
	fmt.Printf("Success")
}

// 启动服务器的地址
func (this *Server) Start() {

	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Printf("net.Listen err", err)
		return
	}

	// close listen
	defer listen.Close()

	for {
		//accept
		accept, err := listen.Accept()
		if err != nil {
			fmt.Printf("listener accept err", err)
			continue
		}

		//do handler
		go this.Handler(accept)
	}

}
