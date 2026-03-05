package internal

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"net"
)

// Client represents a connected user (logic-only, no TCP here)
type Client struct {
	Conn 	 net.Conn
	Name     string
	Messages chan string // outbound messages to this client
}

// Server represents the message engine
type Server struct {
	Clients   map[string]*Client
	Broadcast chan Message
	Join      chan *Client
	Leave     chan *Client
	History   []string
	Mutex     sync.Mutex
	MaxConn   int
}

// Message structure sent through Broadcast channel
type Message struct {
	Sender  *Client
	Content string
}

// NewServer creates a new message engine
func NewServer(maxConn int) *Server {
	return &Server{
		Clients:   make(map[string]*Client),
		Broadcast: make(chan Message),
		Join:      make(chan *Client),
		Leave:     make(chan *Client),
		History:   []string{},
		MaxConn:   maxConn,
	}
}

// Run starts the message engine loop
func (s *Server) Run() {
	for {
		select {

		// New client joining
		case client := <-s.Join:
			if len(s.Clients) >= s.MaxConn {
				client.Messages <- "Server is full. Connection rejected."
				close(client.Messages)
				continue
			}

			s.Clients[client.Name] = client

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
			if _, exists := s.Clients[client.Name]; exists {
				delete(s.Clients, client.Name)
				close(client.Messages)

				leaveMsg := formatSystemMessage(fmt.Sprintf("%s has left our chat.", client.Name))
				s.addToHistory(leaveMsg)
				s.broadcastToOthers(leaveMsg, client)
			}

		// Incoming chat message
		case msg := <-s.Broadcast:
			content := strings.TrimSpace(msg.Content)

			// Do not broadcast empty messages
			if content == "" {
				continue
			}

			formatted := formatUserMessage(msg.Sender.Name, content)
			s.addToHistory(formatted)

			// Broadcast to everyone except sender
			s.broadcastToOthers(formatted, msg.Sender)
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
