package main

import (
	"net"
)

// User 用户结构
type User struct {
	Name string
	Addr string
	//每个用户都有一个 chan
	C chan string
	//与客户端连接的句柄
	conn net.Conn
}

// NewUser 创建一个用户
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		// 默认用户名为用户的ip地址
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}
	go user.ListenMessage()
	return user
}

// ListenMessage 监听当前User channel ，一旦有消息，就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		//fmt.Println("客户端消息测试")
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
