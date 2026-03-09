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
	name = strings.Trim(name, "\n")

	if strings.TrimSpace(name) == "" {
		c.Write([]byte("Invalid input, use a valid name"))
	}

	client.Name = name
	s.Join <- client

	log.Printf("Client connected: %s\n", client.Name)

	s.Mutex.Lock()
	s.Clients[name] = client
	s.Mutex.Unlock()

	go client.ReadInput(s)
	go client.WriteOutput()

	fmt.Printf("total clients: %d\n", len(s.Clients))

}
