package server

import (
	_ "bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net-cat/cmd"
	"net-cat/service"
	"os"
	_ "strings"
)

const maxClients = 10

type Options struct {
	InfoWriter  io.Writer
	ErrorLogger *log.Logger
}

// Start initializes the TCP listener and accepts clients forever.
func Start(port string) error {
	return StartWithOptions(port, Options{})
}

// StartWithOptions initializes the TCP listener with configurable output.
func StartWithOptions(port string, opts Options) error {
	// If no CLI port is provided, use the shared model default.
	if port == "" {
		port = service.DefaultPort
	}

	infoWriter := opts.InfoWriter
	if infoWriter == nil {
		infoWriter = os.Stdout
	}
	errorLogger := opts.ErrorLogger
	if errorLogger == nil {
		errorLogger = log.Default()
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Fprintln(infoWriter, "Listening on the port :"+port)

	server := service.NewServer(maxClients)

	for {
		conn, err := listener.Accept()
		if err != nil {
			errorLogger.Println("Error accepting connection:", err)
			continue
		}

		// Each client gets its own goroutine so connections are concurrent.
		go handleConnection(server, conn)
		go server.Broadcasts()
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
