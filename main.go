package main

import (
	"fmt"
	// "net"
	"os"
	"net-cat/server"
	// "net-cat/service"
)

// const defaultPort = "8989"


func main() {
	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	port := ""
	if len(args) == 1 {
		port = args[0]
	}

	if err := server.Start(port); err != nil {
		fmt.Println("Error:", err)
	}
}