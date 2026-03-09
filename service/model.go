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
	Clients   map[net.Conn]*Client
	Broadcast chan string
	Join      chan *Client
	Leave     chan string
	History   []string
	Mutex     sync.Mutex
}

func NewServer() *Server {
	return &Server{
		Clients:   make(map[net.Conn]*Client),
		Broadcast: make(chan string, 100),
		Join:      make(chan *Client, 100),
		Leave:     make(chan string, 100),
		History:   []string{},
	}
}
