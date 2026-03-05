package cmd

import (
	"bufio"
	"net"
	"net-cat/service"
	"net-cat/utils"
	"strings"
)

func HandleClient(c net.Conn, s *service.Server) {
	client := &service.Client{Conn: c}

	c.Write([]byte(utils.Banner))
	reader := bufio.NewReader(c)

	name, _ := reader.ReadString('\n')
	name = strings.Trim(name, "\n")

	if name == "" {
		c.Write([]byte("Invalid input, use a valid name"))
	}
	client.Name = name
	s.Clients[c] = client

	go client.WriteOutput()
	go client.ReadeInput(s)

}


