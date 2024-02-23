package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name, Addr string
	C          chan string
	conn       net.Conn
	server     *Server
}

// User factory

func NewUser(conn net.Conn, server *Server) *User {
	puser := &User{
		conn.RemoteAddr().String(),
		conn.RemoteAddr().String(),
		make(chan string),
		conn,
		server,
	}

	// start a goroutine to listen user channel
	go puser.ListenMessage()

	return puser
}

func (this *User) Online() {
	// add user to onlinemap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// broadcast user online
	this.server.Broadcast(this, " is online")
}

func (this *User) Offline() {
	// delete user in usermap
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// broadcast user offline
	this.server.Broadcast(this, " is offline")
}

// send message to connection to this user
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// solve message
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// query online users
		this.server.mapLock.Lock()
		onlinemsg := ""
		for _, user := range this.server.OnlineMap {
			onlinemsg += fmt.Sprintf("[%s]%s\n", user.Addr, user.Name)
		}
		this.server.mapLock.Unlock()
		this.SendMsg(onlinemsg)
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// format: rename|newname
		newName := msg[7:]

		// see if name is already occupied
		this.server.mapLock.Lock()
		_, ok := this.server.OnlineMap[newName]
		if ok {
			//name already exist
			this.server.mapLock.Unlock()
			this.SendMsg(fmt.Sprintf("Username [%s] is occupied! Rename failed\n", newName))
		} else {
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			this.Name = newName
			this.SendMsg(fmt.Sprintf("Success! Your new username is [%s]\n", newName))
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		// format: to|name|message
		// get username
		arr := strings.Split(msg, "|")
		if len(arr) != 3 {
			this.SendMsg("Incorrect format. Usage: to|username|message\n")
			return
		}
		remoteName := arr[1]
		if remoteName == "" {
			this.SendMsg("Incorrect format. Usage: to|username|message\n")
			return
		}
		// get User pointer based on username
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg(fmt.Sprintf("The user [%s] does not exist\n", remoteName))
			return
		}
		// get message and send message
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("Cannot send empty message\n")
			return
		}
		remoteUser.SendMsg(fmt.Sprintf("[PRIVATE MESSAGE]%s %s\n", this.Name, content))

	} else {
		this.server.Broadcast(this, msg)
	}

}

// listen the User channel, if there is a message, send to client
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
