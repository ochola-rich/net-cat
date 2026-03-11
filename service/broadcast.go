package service

import (
	"fmt"
	"strings"
	"time"
)

// NewServer creates a new message engine
func NewServer(maxConn int) *Server {
	s := &Server{
		Groups:  make(map[string]*Group),
		MaxConn: maxConn,
	}
	s.GetOrCreateGroup("lobby")
	return s
}

func (s *Server) GetOrCreateGroup(name string) *Group {
	s.Mutex.Lock()
	if group, ok := s.Groups[name]; ok {
		s.Mutex.Unlock()
		return group
	}
	group := &Group{
		Name:      name,
		Clients:   make(map[string]*Client),
		Broadcast: make(chan Message),
		Join:      make(chan *Client),
		Leave:     make(chan *Client),
		History:   []string{},
	}
	s.Groups[name] = group
	s.Mutex.Unlock()

	go group.Broadcasts()
	return group
}

func (s *Server) TotalClients() int {
	s.Mutex.Lock()
	groups := make([]*Group, 0, len(s.Groups))
	for _, group := range s.Groups {
		groups = append(groups, group)
	}
	s.Mutex.Unlock()

	total := 0
	for _, group := range groups {
		group.Mutex.Lock()
		total += len(group.Clients)
		group.Mutex.Unlock()
	}
	return total
}

// Run starts the group message loop
func (g *Group) Broadcasts() {
	for {
		select {
		// New client joining
		case client := <-g.Join:
			g.Mutex.Lock()
			g.Clients[client.Name] = client
			g.Mutex.Unlock()

			// Send chat history to new client
			for _, msg := range g.History {
				client.Messages <- msg
			}

			// Broadcast join message
			joinMsg := formatSystemMessage(fmt.Sprintf("%s has joined our chat.", client.Name))
			g.addToHistory(joinMsg)
			g.broadcastToOthers(joinMsg, client)

		// Client leaving
		case client := <-g.Leave:
			g.Mutex.Lock()
			delete(g.Clients, client.Name)
			g.Mutex.Unlock()

			leaveMsg := formatSystemMessage(fmt.Sprintf("%s has left our chat.", client.Name))
			g.addToHistory(leaveMsg)
			g.broadcastToOthers(leaveMsg, client)

			// Incoming chat message
		case msg := <-g.Broadcast:
			// Trim input and ignore blank messages.
			content := strings.TrimSpace(msg.Content)
			if content == "" {
				continue
			}
			if msg.Sender == nil {
				continue
			}

			formatted := formatUserMessage(msg.Sender.Name, content)
			g.addToHistory(formatted)

			// Broadcast to everyone except the sender.
			g.broadcastToOthers(formatted, msg.Sender)
		}
	}
}

// Broadcast to all clients except the sender
func (g *Group) broadcastToOthers(message string, sender *Client) {
	for _, client := range g.Clients {
		if sender != nil && client.Name == sender.Name {
			continue
		}
		client.Messages <- message
	}
}

// Safely add message to history
func (g *Group) addToHistory(message string) {
	g.Mutex.Lock()
	g.History = append(g.History, message)
	g.Mutex.Unlock()
}

// Format user message
func formatUserMessage(name, msg string) string {
	return fmt.Sprintf("[%s][%s]: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		name,
		msg,
	)
}

// Format system message
func formatSystemMessage(msg string) string {
	return fmt.Sprintf("[%s][System]: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		msg,
	)
}
