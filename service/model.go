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
	Group    *Group
}

type Server struct {
	Groups  map[string]*Group
	MaxConn int
	Mutex   sync.Mutex
}

type Group struct {
	Name      string
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
