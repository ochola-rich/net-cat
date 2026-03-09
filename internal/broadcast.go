// Package internal contains the server implementation for the net-cat application.
package internal

import (
	"fmt"
	"net-cat/service"
	"strings"
	"time"
)

type MyServer struct {
    *service.Server
}


// Run starts the message engine loop
func (s *MyServer) broadcasts() {
	for {
		select {

			// New client joining
			case client := <-s.Join:
				
				for _, msg := range s.History {
					client.Messages <- msg
				}

				// Broadcast join message
				joinMsg := formatSystemMessage(fmt.Sprintf("%s has joined our chat.", client.Name))
				s.addToHistory(joinMsg)
				s.broadcastToOthers(joinMsg, client)

			// Client leaving
			case client := <-s.Leave:
			
				close(client.Messages)

				leaveMsg := formatSystemMessage(fmt.Sprintf("%s has left our chat.", client.Name))
				s.addToHistory(leaveMsg)
				s.broadcastToOthers(leaveMsg, client)

			// Incoming chat message
			case msg := <-s.Broadcast:
				// broadcast channel carries a plain string; trim it and ignore
				// empty payloads. sender information is not available, so pass
				// nil to broadcastToOthers.
				content := strings.TrimSpace(msg)

			
				formatted := formatUserMessage("", content)
				s.addToHistory(formatted)

				// Broadcast to everyone; no sender to exclude.
				s.broadcastToOthers(formatted, nil)
		}
	}
}

// Broadcast to all clients except the sender
func (s *MyServer) broadcastToOthers(message string, sender *service.Client) {
	for _, client := range s.Clients {
		if sender != nil && client.Name == sender.Name {
			continue
		}
		client.Messages <- message
	}
}

// Safely add message to history
func (s *MyServer) addToHistory(message string) {
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
