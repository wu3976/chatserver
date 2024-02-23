package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// user map
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// main message channel
	Message chan string
}

// create new server
// factory mode
func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

// listen to Message channel
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s:%s\n", user.Addr, user.Name, msg)
	this.Message <- sendMsg
}

func (this *Server) handleConnection(conn net.Conn) {
	fmt.Printf("Connected to remote client %s\n", conn.RemoteAddr().String())
	user := NewUser(conn, this)
	user.Online()

	// channel to detect user alive
	isLive := make(chan bool)

	// receive message from user
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn read error: ", err)
				return
			}

			msg := string(buf[:n-1]) // discard newline character
			// user resolve message
			user.DoMessage(msg)

			isLive <- true // user is active
		}

	}()
	/*
	* A select statement to detect if user is inactive for some time.
	*
	* 1. when user Conn receive message, isLive channel get a message
	* then select end blocking, and slip to 2nd case
	* the time.After returns a new Timer channel(with brand new 30sec)
	*
	* 2. if user is idle, case 1 fail, then evaluate case 2, time.After
	* return a new Timer channel, then the whole select block until timer
	* ends, case 2 get value from Timer channel, or user is active, case 1
	* get value
	 */
	for {
		select {
		case <-isLive:
			// do nothing
		case <-time.After(time.Second * 300):
			// time over
			user.SendMsg("You are kicked because of inactive\n")
			// close resources
			close(user.C)
			conn.Close()
			// end current goroutine
			return
		}
	}
}

// start server
func (this *Server) Start() {
	// tcp protocol
	fmt.Printf("Staring server at %s:%d\n", this.Ip, this.Port)
	// listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
	}

	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("net Accept error: ", err)
			continue
		}
		go this.handleConnection(conn)
	}

	// close main socket

	defer listener.Close()
}
