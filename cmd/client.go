package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net-cat/service"
	"net-cat/utils"
	"strings"
)

func HandleClient(c net.Conn, s *service.Server) {
	client := &service.Client{Conn: c, Messages: make(chan string)}

	c.Write([]byte(utils.Banner))
	reader := bufio.NewReader(c)

	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		c.Write([]byte("Invalid input, use a valid name\n"))
		c.Close()
		return
	}

	group := s.GetOrCreateGroup("lobby")
	group.Mutex.Lock()
	if _, exists := group.Clients[name]; exists {
		group.Mutex.Unlock()
		c.Write([]byte("Name already taken in this group.\n"))
		c.Close()
		return
	}
	group.Mutex.Unlock()

	client.Name = name
	client.Group = group
	group.Join <- client

	log.Printf("Client connected: %s (group: %s)\n", client.Name, group.Name)

	go client.ReadInput(s)
	go client.WriteOutput()

	fmt.Printf("total clients: %d\n", s.TotalClients())

}
