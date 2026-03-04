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
	Client    map[net.Conn]*Client
	Broadcast chan string
	Join      chan *Client
	Leave     chan *Client
	History   []string
	Mutex     sync.Mutex
}
