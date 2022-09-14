package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// 在线用户列表
	OnlineMap map[string]*User
	// 同步锁
	mapLock sync.RWMutex
	// 消息广播 channel
	Message chan string
}

// NewSever 创建一个server
func NewSever(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息的channe的goroutine , 一旦有消息就发送给全部的在线User

func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		// 将从管道中读取的msg 发送给全部的子在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			//写入到每个用户的channel中
			cli.C <- msg
		}
		this.mapLock.Unlock()

	}

}

// BroadCast 广播用户上线消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	// 把消息放入到管道中
	this.Message <- sendMsg

}

// Handler 针对连接 进行业务的处理
func (this *Server) Handler(conn net.Conn) {

	fmt.Println("连接成功")
	fmt.Printf("%s连接建立成功\n", conn.RemoteAddr())

	// 用户上线
	user := NewUser(conn, this)
	user.Online()

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn read err", err)
				return
			}
			//提取用户发送的消息（去除 '\n'） 并进行广播
			msg := string(buf[:n-1])

			user.DoMessage(msg)

		}

		//这里要加上括号调用
	}()

	// 当前handle 进行阻塞
	select {}

}

// Start 启动server
func (this *Server) Start() {
	//socket listen  fmt.Sprint("%s:%d", this.Ip, this.Port)-> 127.0.0.1
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err", err)
		return
	}
	// 启动监听 message 的goroutine
	go this.ListenMessager()

	//close  listen   socket
	defer listener.Close()

	for {
		//accept  这里接受成功之后  可以代表一个用户上线
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept  err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}

}
