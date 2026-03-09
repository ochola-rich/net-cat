package service

import (
	"bufio"
	"strings"
)

func (c *Client) WriteOutput() {
	for msg := range c.Messages {
		c.Conn.Write([]byte(msg + "\n"))
	}
}

func (c *Client) ReadInput(s *Server) {
	scanner := bufio.NewScanner(c.Conn)
	
	for scanner.Scan() {
		msg := scanner.Text()
		msg = strings.TrimSpace(msg)

		if msg == "" {
			continue
		}

		cmsg := Message{Sender: c, Content: msg}
		s.Broadcast <- cmsg
	}

	s.Leave <- c
	s.Mutex.Lock()
	delete(s.Clients, c.Name)
	s.Mutex.Unlock()
}

// func NewServer() *Server {
//     return &Server{
//         Clients:   make(map[net.Conn]*Client),
//         Broadcast: make(chan string, 100),
//         Join:      make(chan *Client, 100),
//         Leave:     make(chan *Client, 100),
//     }
// }
