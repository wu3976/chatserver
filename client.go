package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // client mode
}

func NewClient(server_ip string, server_port int) *Client {
	cli := &Client{
		ServerIP:   server_ip,
		ServerPort: server_port,
		flag:       -1,
	}
	cli.Connect()
	return cli
}

func (client *Client) DealResponse() {
	// as soon as conn has message, copy it to stdout
	// iteratively and blocking
	io.Copy(os.Stdout, client.conn)
}

func (this *Client) Connect() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", this.ServerIP, this.ServerPort))
	if err != nil {
		fmt.Println("net.Dial error")
		os.Exit(1)
	}
	this.conn = conn
}

func (this *Client) PublicChat() {
	var chatMsg string
	// prompt user to enter message
	fmt.Println("Enter chat message, or enter exit to exit. Press enter to send.")
	//fmt.Scanln(&chatMsg)
	reader := bufio.NewReader(os.Stdin)
	chatMsg, _ = reader.ReadString(byte('\n'))
	for chatMsg != "exit\r\n" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write error: ", err)
				fmt.Println("Message not sent")
			} else {
				fmt.Println("Message sent")
			}
		}
		fmt.Scanln(&chatMsg)
	}
}

func (this *Client) QueryUsers() {
	_, err := this.conn.Write([]byte("who\n"))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
	}
}

func (this *Client) PrivateChat() {
	this.QueryUsers()
	fmt.Println("Enter the username you want to chat, or enter exit to quit")
	var remoteName, chatMsg string
	fmt.Scanln(&remoteName)
	fmt.Println("Now you can enter your messages. Enter exit to quit")
	reader := bufio.NewReader(os.Stdin)
	for remoteName != "exit" {
		//fmt.Scanln(&chatMsg)
		temp, _ := reader.ReadString(byte('\n'))
		chatMsg = temp
		if chatMsg == "exit\r\n" {
			break
		}
		sendMsg := fmt.Sprintf("to|%s|%s", remoteName, chatMsg)
		fmt.Println(sendMsg)
		_, err := this.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn.Write error: ", err)
			break
		}
	}

}

func (this *Client) UpdateName() bool {
	fmt.Println("Enter username: ")
	fmt.Scanln(&this.Name)
	send_msg := fmt.Sprintf("rename|%s\n", this.Name)
	_, err := this.conn.Write([]byte(send_msg))
	if err != nil {
		fmt.Println("conn.Write error: ", err)
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for !client.menu() {
		}

		// behave differently based on flag
		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.UpdateName()
			break
		}
	}
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1. public chat")
	fmt.Println("2. private chat")
	fmt.Println("3. rename")
	fmt.Println("0. exit")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("Please enter a valid option")
		return false
	}
}

var sip string
var sport int

func init() { // automatically called by flag.Parse()
	flag.StringVar(&sip, "ip", "127.0.0.1", "set server ip (default 127.0.0.1)")
	flag.IntVar(&sport, "port", 8400, "set server port (default 8400)")
}

func main() {
	// cmd parsing
	flag.Parse()

	client := NewClient(sip, sport)

	// open a goroutine to print server response
	go client.DealResponse()

	fmt.Println("Successfully connecting to server")
	client.Run()
}
