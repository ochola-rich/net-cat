package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

const maxClients = 10
const defaultPort = "8989"

type Client struct {
	Conn net.Conn
	Name string
}

type Server struct {
	Clients map[net.Conn]*Client
	Mutex   sync.Mutex
}

// Start initializes and runs the TCP server
func Start(port string) error {
	if port == "" {
		port = defaultPort
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Println("Listening on the port :" + port)

	server := &Server{
		Clients: make(map[net.Conn]*Client),
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go server.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	// Enforce max clients
	s.Mutex.Lock()
	if len(s.Clients) >= maxClients {
		s.Mutex.Unlock()
		conn.Write([]byte("Server full. Maximum 10 clients allowed.\n"))
		conn.Close()
		return
	}
	s.Mutex.Unlock()

	sendWelcome(conn)

	name, err := askName(conn)
	if err != nil {
		conn.Close()
		return
	}

	client := &Client{
		Conn: conn,
		Name: name,
	}

	s.registerClient(client)
	defer s.removeClient(client)

	log.Printf("Client connected: %s\n", client.Name)

	// Keep connection alive
	buffer := make([]byte, 1)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("Client disconnected: %s\n", client.Name)
			} else {
				log.Printf("Read error from %s: %v\n", client.Name, err)
			}
			return
		}
	}
}

func (s *Server) registerClient(client *Client) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Clients[client.Conn] = client
}

func (s *Server) removeClient(client *Client) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	delete(s.Clients, client.Conn)
	client.Conn.Close()
}

func sendWelcome(conn net.Conn) {
	welcome := "Welcome to TCP-Chat!\n" +
		"         _nnnn_\n" +
		"        dGGGGMMb\n" +
		"       @p~qp~~qMb\n" +
		"       M|@||@) M|\n" +
		"       @,----.JM|\n" +
		"      JS^\\__/  qKL\n" +
		"     dZP        qKRb\n" +
		"    dZP          qKKb\n" +
		"   fZP            SMMb\n" +
		"   HZM            MMMM\n" +
		"   FqM            MMMM\n" +
		" __| \".        |\\dS\"qML\n" +
		" |    `.       | ` \\Zq\n" +
		"_)      \\.___.,|     .'\n" +
		"\\____   )MMMMMP|   .'\n" +
		"     `-'       `--'\n" +
		"[ENTER YOUR NAME]: "

	conn.Write([]byte(welcome))
}

func askName(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)

	for {
		name, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		name = strings.TrimSpace(name)

		if name == "" {
			conn.Write([]byte("Name cannot be empty. Try again: "))
			continue
		}

		return name, nil
	}
}