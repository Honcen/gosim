package main

import (
	"fmt"
	"net"
	"sync"
)

/**
* server 结构体
 */
type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	//用户锁
	mapLock sync.RWMutex
	//消息广播 channel
	Message chan string
}

//创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

//监听Message广播消息channel的goroutine
//一旦有消息就发送给全部的在线User
func (this *Server) ListenMessager() {
	for {
		//从消息 channel中读出信息
		msg := <-this.Message
		//将消息发送给全部在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

//handler
func (this *Server) Handler(conn net.Conn) {
	//当前连接业务\
	// fmt.Println("连接建立成功")
	user := NewUser(conn)

	//用户上线，将用户加入到 OnlineMap 中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	//广播当前用户上线消息
	this.BroadCast(user, "上线了")

	//当前handler 阻塞
	select {}
}

//启动服务器的接口
func (this *Server) Start() {
	// socket 监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net. Listen err:", err)
	}
	defer listener.Close()
	//启动监听message的goroutine
	go this.ListenMessager()
	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(" listener accept err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)
	}

}
