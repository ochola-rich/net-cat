package server

import (
	_ "bufio"
	"fmt"
	_ "io"
	"log"
	"net"
	"net-cat/cmd"
	"net-cat/service"
	_ "strings"
)

const maxClients = 10

// Start initializes the TCP listener and accepts clients forever.
func Start(port string) error {
	// If no CLI port is provided, use the shared model default.
	if port == "" {
		port = service.DefaultPort
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Listening on the port :" + port)

	server := service.NewServer(maxClients)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Each client gets its own goroutine so connections are concurrent.
		go handleConnection(server, conn)
		go server.Run()
	}
}

// handleConnection manages one client from connect to disconnect.
func handleConnection(s *service.Server, conn net.Conn) {
	// Enforce max clients (best-effort under concurrent connects).
	s.Mutex.Lock()
	if len(s.Clients) >= maxClients {
		s.Mutex.Unlock()
		conn.Write([]byte("Server full. Maximum 10 clients allowed.\n"))
		conn.Close()
		return
	}
	s.Mutex.Unlock()

	go cmd.HandleClient(conn, s)
}