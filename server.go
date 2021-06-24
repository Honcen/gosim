package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
	user := NewUser(conn, this)

	//用户上线，将用户加入到 OnlineMap 中
	// this.mapLock.Lock()
	// this.OnlineMap[user.Name] = user
	// this.mapLock.Unlock()
	// //广播当前用户上线消息
	// this.BroadCast(user, "上线了")
	user.Online()

	//创建一个channel 用于监听用户是否活跃  往channel发送 true 时表示用户活跃
	isLive := make(chan bool)

	//接收用户端发送的消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用户消息（去除 ’\n‘）
			msg := string(buf[:n-1])
			//将得到的消息广播
			// this.BroadCast(user, msg)
			user.DoMessage(msg)

			//用户的任意消息都能激活
			isLive <- true
		}
	}()

	//当前handler 阻塞
	for {
		select {
		case <-isLive:
			//当前用户活跃的
			//什么都不写会自动执行下个case 的条件，定时器完成自动重置
		case <-time.After(time.Second * 30):
			//定时器超过 10秒 则超时
			//将当前 User 强制关闭
			user.SendMsg("你已超时被踢")
			//销毁用户
			close(user.C)
			//关闭连接
			conn.Close()
			//退出当前handler
			return //或者 runtime.Goexit()
		}
	}

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
