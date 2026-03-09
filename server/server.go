package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net-cat/service"
	"strings"
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

	server := service.NewServer() // &service.Server{
	// 	// Active clients keyed by connection object.
	// 	Clients: make(map[net.Conn]*service.Client),
	// 	Mutex:  sync.Mutex{},
	// }
	go consumeClientLeaves(server) // starts the goroutine that checks for disconnects

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Each client gets its own goroutine so connections are concurrent.
		go handleConnection(server, conn) // 
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

	sendWelcome(conn)

	name, err := askName(conn)
	if err != nil {
		conn.Close()
		return
	}

	client := &service.Client{
		Conn: conn,
		Name: name,
	}

	// Register now; always remove on any later return path.
	registerClient(s, client)
	defer func() {
		s.Leave <- client.Name // does not remove the client directry, it instead sends his/ her name to leaves
	}()

	log.Printf("Client connected: %s\n", client.Name)

	// Keep connection alive by reading until the peer disconnects.
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

// registerClient stores a connected client in shared state.
func registerClient(s *service.Server, client *service.Client) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	s.Clients[client.Conn] = client
}

// removeClient deletes a client by name from shared state and closes its socket.
func removeClient(s *service.Server, clientName string) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	for conn, client := range s.Clients {
		if client.Name != clientName {
			continue
		}

		delete(s.Clients, conn)
		conn.Close()
		return
	}
}

// consumeClientLeaves removes clients when disconnect events are received.
func consumeClientLeaves(s *service.Server) { // go routine that constantry checks if a client has left
	for clientName := range s.Leave {
		removeClient(s, clientName)
	}
}

// sendWelcome writes the banner and prompts the user for their name.
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

// askName reads newline-terminated input and rejects empty names.
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
