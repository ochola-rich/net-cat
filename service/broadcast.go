package service

import (
	"fmt"
	"strings"
	"time"
	"sync"
)

// NewServer creates a new message engine
func NewServer(maxConn int) *Server {
	return &Server{
		Clients:   make(map[string]*Client),
		Broadcast: make(chan Message),
		Join:      make(chan *Client),
		Leave:     make(chan *Client),
		History:   []string{},
		Mutex:     sync.Mutex{},
	}
}

// Run starts the message engine loop
func (s *Server) Broadcasts() {
	for {
		select {
		// New client joining
		case client := <-s.Join:
			// Send chat history to new client
			for _, msg := range s.History {
				client.Messages <- msg
			}

			// Broadcast join message
			joinMsg := formatSystemMessage(fmt.Sprintf("%s has joined our chat.", client.Name))
			s.addToHistory(joinMsg)
			s.broadcastToOthers(joinMsg, client)

		// Client leaving
		case client := <-s.Leave:
			leaveMsg := formatSystemMessage(fmt.Sprintf("%s has left our chat.", client.Name))
			s.addToHistory(leaveMsg)
			s.broadcastToOthers(leaveMsg, client)

			// Incoming chat message
		case msg := <-s.Broadcast:
				// broadcast channel carries a plain string; trim it and ignore
				// empty payloads. sender information is not available, so pass
				// nil to broadcastToOthers.
			content := strings.TrimSpace(msg.Content)

			
			formatted := formatUserMessage(msg.Sender.Name, content)
			s.addToHistory(formatted)

				// Broadcast to everyone; no sender to exclude.
			s.broadcastToOthers(formatted, nil)
		}
	}
}

// Broadcast to all clients except the sender
func (s *Server) broadcastToOthers(message string, sender *Client) {
	for _, client := range s.Clients {
		if sender != nil && client.Name == sender.Name {
			continue
		}
		client.Messages <- message
	}
}

// Safely add message to history
func (s *Server) addToHistory(message string) {
	s.Mutex.Lock()
	s.History = append(s.History, message)
	s.Mutex.Unlock()
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
