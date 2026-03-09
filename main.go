package main

import (
	"fmt"
	"net-cat/server"
	"os"
)

func main() {
	// Accepts zero or one argument: optional TCP port.
	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	// Empty port delegates to server default.
	port := ""
	if len(args) == 1 {
		port = args[0]
	}

	// Start blocks forever unless startup fails.
	if err := server.Start(port); err != nil {
		fmt.Println("Error:", err)
	}
}
