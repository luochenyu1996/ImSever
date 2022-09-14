package main

import (
	"net"
	"strings"
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
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

// DoMessage 用户消息处理
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		this.QueryOnlineUser()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		this.ReName(msg)
	} else if len(msg) > 4 && msg[:3] == "to|" {
		this.ToOnePerson(msg)
	} else {
		this.server.BroadCast(this, msg)
	}

}

// ToOnePerson 给指定的用户发送消息，私聊功能的实现。  消息格式：to|用户名|消息
func (this *User) ToOnePerson(command string) {
	remoteName := strings.Split(command, "|")[1]

	if remoteName == "" {
		this.SendMsg("消息格式不正确。发送消息的命令格式为：to|用户名|消息\n")
		return
	}
	remoteUser, ok := this.server.OnlineMap[remoteName]
	if !ok {
		this.SendMsg("用户名不存在")
		return
	}
	content := strings.Split(command, "|")[2]
	if content == "" {
		this.SendMsg("消息格式不正确,无消息内容。发送消息的命令格式为：to|用户名|消息\n")
	}
	remoteUser.SendMsg(this.Name + "给你发来消息：" + content)
}

// ReName 修改用户名
func (this *User) ReName(command string) {
	//命令格式 rename|chenyu
	newName := strings.Split(command, "|")[1]
	//判断该用户名是否占用
	_, existUser := this.server.OnlineMap[newName]
	if existUser {
		this.SendMsg("当前用户名已经被占用\n")
	} else {
		this.server.mapLock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()
		this.Name = newName
		this.SendMsg("您的用户名已经更新为：" + this.Name + "\n")
	}
}

// QueryOnlineUser 查询当前在线用户
func (this *User) QueryOnlineUser() {
	this.server.mapLock.Lock()
	for _, user := range this.server.OnlineMap {
		onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
		this.SendMsg(onlineMsg)
	}
	this.server.mapLock.Unlock()
}

// SendMsg 给当前user对应的客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
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
