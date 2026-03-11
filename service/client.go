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

		if c.Group == nil {
			c.Group = s.GetOrCreateGroup("lobby")
		}

		const changeCmd = "--change-name"
		const joinCmd = "--join"
		if len(msg) >= len(joinCmd) && strings.EqualFold(msg[:len(joinCmd)], joinCmd) {
			groupName := strings.TrimSpace(msg[len(joinCmd):])
			if groupName == "" {
				c.Messages <- "Usage: --join <group>"
				continue
			}

			if strings.EqualFold(c.Group.Name, groupName) {
				c.Messages <- "You are already in that group."
				continue
			}

			newGroup := s.GetOrCreateGroup(groupName)
			newGroup.Mutex.Lock()
			if _, exists := newGroup.Clients[c.Name]; exists {
				newGroup.Mutex.Unlock()
				c.Messages <- "Name already taken in that group."
				continue
			}
			newGroup.Mutex.Unlock()

			oldGroup := c.Group
			if oldGroup != nil {
				oldGroup.Leave <- c
			}
			c.Group = newGroup
			newGroup.Join <- c
			c.Messages <- fmt.Sprintf("Joined group %s.", newGroup.Name)
			continue
		}

		if len(msg) >= len(changeCmd) && strings.EqualFold(msg[:len(changeCmd)], changeCmd) {
			newName := strings.TrimSpace(msg[len(changeCmd):])
			if newName == "" {
				c.Messages <- "Usage: --change-name <name>"
				continue
			}

			group := c.Group
			group.Mutex.Lock()
			if _, exists := group.Clients[newName]; exists {
				group.Mutex.Unlock()
				c.Messages <- "Name already taken. Choose a different name."
				continue
			}

			oldName := c.Name
			delete(group.Clients, oldName)
			group.Clients[newName] = c
			c.Name = newName
			group.Mutex.Unlock()

			systemMsg := formatSystemMessage(fmt.Sprintf("%s changed name to %s.", oldName, newName))
			group.addToHistory(systemMsg)
			group.broadcastToOthers(systemMsg, c)
			c.Messages <- "Name updated."
			continue
		}

		cmsg := Message{Sender: c, Content: msg}
		c.Group.Broadcast <- cmsg
	}

	if c.Group != nil {
		c.Group.Leave <- c
	}
}

// func NewServer() *Server {
//     return &Server{
//         Clients:   make(map[net.Conn]*Client),
//         Broadcast: make(chan string, 100),
//         Join:      make(chan *Client, 100),
//         Leave:     make(chan *Client, 100),
//     }
// }
