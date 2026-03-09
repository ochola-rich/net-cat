package service

import (
	"net"
	"sync"
)

const DefaultPort = "8989"

type Client struct {
	Conn     net.Conn
	Name     string
	Messages chan string
}

type Server struct {
	Clients   map[string]*Client
	Broadcast chan Message
	Join      chan *Client
	Leave     chan *Client
	History   []string
	Mutex     sync.Mutex
}

type Message struct {
	Sender  *Client
	Content string
}
