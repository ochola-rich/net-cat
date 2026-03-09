package service

import (
	"bufio"
)

func (c *Client) WriteOutput() {
	for msg := range c.Messages {
		c.Conn.Write([]byte(msg))
	}
}

func (c *Client) ReadeInput(s *Server) {
	scanner := bufio.NewScanner(c.Conn)
	for scanner.Scan() {
		msg := scanner.Text()
		if msg == "" {
			continue
		}
		s.Broadcast <- msg
	}
	s.Leave <- c.Name
}// package main

// import (
// 	"fmt"

// 	ac "net-cat/internal"
// )

// func main() {
// 	server := ac.NewServer(5)
// 	go server.Run()

// 	// Simulated clients
// 	client1 := &ac.Client{Conn: nil, Name: "Alice", Messages: make(chan string, 10)}
// 	client2 := &ac.Client{Conn: nil, Name: "Bob", Messages: mak// package main

// import (
// 	"fmt"

// 	ac "net-cat/internal"
// )

// func main() {
// 	server := ac.NewServer(5)
// 	go server.Run()

// 	// Simulated clients
// 	client1 := &ac.Client{Conn: nil, Name: "Alice", Messages: make(chan string, 10)}
// 	client2 := &ac.Client{Conn: nil, Name: "Bob", Messages: make(chan string, 10)}

// 	server.Join <- client1
// 	server.Join <- client2

// 	server.Broadcast <- ac.Message{Sender: client1, Content: "Hello Bob!"}
// 	server.Broadcast <- ac.Message{Sender: client2, Content: "Hi Alice!"}

// 	// Read messages
// 	for i := 0; i < 2; i++ {
// 		fmt.Println(<-client1.Messages)
// 		fmt.Println(<-client2.Messages)
// 	}
// }
// e(chan string, 10)}

// 	server.Join <- client1
// 	server.Join <- client2

// 	server.Broadcast <- ac.Message{Sender: client1, Content: "Hello Bob!"}
// 	server.Broadcast <- ac.Message{Sender: client2, Content: "Hi Alice!"}

// 	// Read messages
// 	for i := 0; i < 2; i++ {
// 		fmt.Println(<-client1.Messages)
// 		fmt.Println(<-client2.Messages)
// 	}
// }

