package main

import (
	"fmt"
	"net"
)

/**
* server 结构体
 */
type Server struct {
	Ip   string
	Port int
}

//创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}

	return server
}

//handler
func (this *Server) Handler(conn net.Conn) {
	//当前连接业务\
	fmt.Println("连接建立成功")
}

//启动服务器的接口
func (this *Server) Start() {
	// socket 监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net. Listen err:", err)
	}
	defer listener.Close()

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
