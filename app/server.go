package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

// toBulkString formats a string as a Redis RESP bulk string
// Format: $<length>\r\n<data>\r\n

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Default port
	port := "6379"

	// Check for command line arguments
	if len(os.Args) > 1 {
		// Look for --port flag
		for i := 1; i < len(os.Args); i++ {
			if os.Args[i] == "--port" && i+1 < len(os.Args) {
				port = os.Args[i+1]
				break
			}
		}
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(port); err != nil {
		fmt.Printf("Invalid port number: %s\n", port)
		os.Exit(1)
	}

	address := "0.0.0.0:" + port
	fmt.Printf("Starting Redis server on %s\n", address)

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", port, err)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection")
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Connection closed or error:", err)
			return
		}

		// Only process if we received data
		command := string(buffer[:n])

		// Use the unified command handler
		go commands.HandleCommand(conn, command)
	}
}
