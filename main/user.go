package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string //用户绑定的channel
	conn net.Conn

	server *Server
}

// 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	go user.ListenMessage()

	return user
}

// 监听当前的User channel方法，一旦有消息，就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线
func (this *User) Online() {
	//用户上线，将用户放入onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线
	this.server.BroadCast(this, "已上线")
}

// 用户下线
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播用户下线
	this.server.BroadCast(this, "已下线")
}

// 用户处理消息的业务
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}
