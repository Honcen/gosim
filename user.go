package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

//创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	//创建user对象
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel 消息的goroutine
	go user.ListenMessage()
	return user
}

//监听 当前user channel 的方法，一旦有消息。就直接发生消息给连接客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}

//用户上线
func (this *User) Online() {
	//用户上线，将用户加入到 OnlineMap 中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//广播当前用户上线消息
	this.server.BroadCast(this, "上线了")
}

//用户下线
func (this *User) Offline() {
	//用户上线，将用户从OnlineMap 中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "下线了")
}

//发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

//用户消息出来
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户数 打印出在线信息
		// this.sendMsg("当前共有" + strconv.Itoa(len(this.server.OnlineMap)) + "人在线")
		this.sendMsg(fmt.Sprintf("当前共有 %d 人在线\n", len(this.server.OnlineMap)))
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "：在线。。。\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()

	} else {
		this.server.BroadCast(this, msg)
	}

}
