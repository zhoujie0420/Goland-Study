package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Printf("net Dial err", err)
		return nil
	}
	client.conn = conn

	return client
}

var serverIp string
var serverPort int

func (client *Client) menu() bool {
	var flag int
	fmt.Printf("1.公聊模式")
	fmt.Printf("2.私聊模式")
	fmt.Printf("3.更新用户名")
	fmt.Printf("0.公聊模式")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Printf(">>>请输入合法的操作")
		return false
	}
}

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址")
	flag.IntVar(&serverPort, "ip", 8888, "设置服务器端口")
}

// 查询在线
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err", err)
		return
	}
}
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string
	client.SelectUsers()
	fmt.Println(">>输入聊天用户")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			//消息不为空则发送
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn wirte err", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>请输入消息内容")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println(">>>>输入聊天用户")
		fmt.Scanln(&remoteName)
	}
}
func (client *Client) PublicChat() {

	//提示用户输入信息
	var chatMsg string
	fmt.Println(">>>输入内容")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {

		// 发送服务端
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>输入内容")
		fmt.Scanln(&chatMsg)
	}
}
func (client *Client) UpdateName() bool {

	fmt.Printf("请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Printf("conn write err", err)
		return false
	}
	return true
}

func (client *Client) DealResponse() {
	//一旦有数据，永久拷贝
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		//根据不同的模式处理业务
		switch client.flag {
		case 1:
			//公聊
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		//私聊
		case 3:
			client.UpdateName()
			break

		}
	}
}
func main() {
	//命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Printf(">>>>>>链接服务器失败")
		return
	}
	//单独开启一个goroutine 去接受消息
	go client.DealResponse()
	fmt.Printf(">>>>>>链接服务器成功")
	client.Run()
}
