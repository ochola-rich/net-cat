package service

import (
	"bufio"
	"fmt"
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

		const changeCmd = "--change-name"
		if len(msg) >= len(changeCmd) && strings.EqualFold(msg[:len(changeCmd)], changeCmd) {
			newName := strings.TrimSpace(msg[len(changeCmd):])
			if newName == "" {
				c.Messages <- "Usage: --change-name <name>"
				continue
			}

			s.Mutex.Lock()
			if _, exists := s.Clients[newName]; exists {
				s.Mutex.Unlock()
				c.Messages <- "Name already taken. Choose a different name."
				continue
			}

			oldName := c.Name
			delete(s.Clients, oldName)
			s.Clients[newName] = c
			c.Name = newName
			s.Mutex.Unlock()

			systemMsg := formatSystemMessage(fmt.Sprintf("%s changed name to %s.", oldName, newName))
			s.addToHistory(systemMsg)
			s.broadcastToOthers(systemMsg, c)
			c.Messages <- "Name updated."
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
