package service

import (
	"bufio"
	"net"
)

func (c *Client)WriteOutput(){
	for msg := range c.Messages{
		c.Conn.Write([]byte(msg))
	}
}

func (c *Client)ReadeInput(s *Server){
 scanner := bufio.NewScanner(c.Conn)
 for scanner.Scan(){
	msg := scanner.Text()
	if msg == ""{
		continue
	}
	s.Broadcast <- msg
 }
 s.Leave <- c
}

func NewServer() *Server {
    return &Server{
        Clients:   make(map[net.Conn]*Client),
        Broadcast: make(chan string, 100),
        Join:      make(chan *Client, 100),
        Leave:     make(chan *Client, 100),
    }
}
