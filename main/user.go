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
	//用户需要和服务进行关联
	server *Server
}

// NewUser 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		// 默认用户名为用户的ip地址
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
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

// 用户消息处理

func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)

}

// Online 用户上线的业务。 需要把用户对象放入到在线用户列表中
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	this.server.BroadCast(this, "已上线")

}

// Offline 用户下线的业务
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	this.server.BroadCast(this, "下线")
}
