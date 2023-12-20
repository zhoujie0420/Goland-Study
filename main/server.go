package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,

		//在线用户的列表
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) Handler(conn net.Conn) {

	// ... 业务逻辑
	//fmt.Printf("Success")

	user := NewUser(conn)

	//用户上线，将用户放入onlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	//广播当前用户上线消息
	this.BroadCast(user, "已上线")

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4086)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Printf("conn read err:", err)
				return
			}

			// 提取消息，驱除 \n
			msg := string(buf[:n-1])
			this.BroadCast(user, msg)
		}
	}()

	// 当前handler阻塞
	select {}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

// 监听Message广播消息的channel的goroutine ,一旦有消息就发送给全部的在线user
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		//将msg发送给全部的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
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

	// 启动监听message的goroutine
	go this.ListenMessage()
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
