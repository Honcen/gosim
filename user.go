package main

import (
	"fmt"
	"net"
	"strings"
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
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) SendToUser(msg string) {
	//消息私聊协议 to|zhang3|nihao
	toName := strings.Split(msg, "|")[1]
	toUser, ok := this.server.OnlineMap[toName]
	if !ok {
		this.SendMsg("用户不存在")
	} else {
		content := strings.Split(msg, "|")[2]
		toUser.SendMsg("[" + this.Addr + "]" + this.Name + " 对你说:" + content)
	}
}

//在线用户查询
func (this *User) Who() {
	//查询当前在线用户数 打印出在线信息
	// this.SendMsg("当前共有" + strconv.Itoa(len(this.server.OnlineMap)) + "人在线")
	this.SendMsg(fmt.Sprintf("当前共有 %d 人在线\n", len(this.server.OnlineMap)))
	this.server.mapLock.Lock()
	for _, user := range this.server.OnlineMap {
		onlineMsg := "[" + user.Addr + "]" + user.Name + "：在线。。。\n"
		this.SendMsg(onlineMsg)
	}
	this.server.mapLock.Unlock()
}

func (this *User) Rename(msg string) {
	newName := strings.Split(msg, "|")[1]
	_, ok := this.server.OnlineMap[newName]
	if ok {
		this.SendMsg("用户名 " + newName + " 已经存在")
	} else {
		this.server.mapLock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()
		this.Name = newName
		this.SendMsg("您的用户名已经更新为：" + newName)
	}
}

//用户消息出来
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户数 打印出在线信息
		// this.SendMsg("当前共有" + strconv.Itoa(len(this.server.OnlineMap)) + "人在线")
		// this.SendMsg(fmt.Sprintf("当前共有 %d 人在线\n", len(this.server.OnlineMap)))
		// this.server.mapLock.Lock()
		// for _, user := range this.server.OnlineMap {
		// 	onlineMsg := "[" + user.Addr + "]" + user.Name + "：在线。。。\n"
		// 	this.SendMsg(onlineMsg)
		// }
		// this.server.mapLock.Unlock()
		this.Who()

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 rename|张三
		// newName := strings.Split(msg, "|")[1]
		// _, ok := this.server.OnlineMap[newName]
		// if ok {
		// 	this.SendMsg("用户名 " + newName + " 已经存在")
		// } else {
		// 	this.server.mapLock.Lock()
		// 	delete(this.server.OnlineMap, this.Name)
		// 	this.server.OnlineMap[newName] = this
		// 	this.server.mapLock.Unlock()
		// 	this.Name = newName
		// 	this.SendMsg("您的用户名已经更新为：" + newName)
		// }
		this.Rename(msg)
	} else if len(msg) > 3 && msg[:3] == "to|" {

		this.SendToUser(msg)
	} else {
		this.server.BroadCast(this, msg)
	}

}
