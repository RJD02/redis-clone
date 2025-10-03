package server

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/handlers"
)

// HandleConnection handles a single client connection
func HandleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Connection closed: %s (%v)\n", conn.RemoteAddr(), err)
			return
		}

		// Only process if we received data
		if n > 0 {
			command := string(buffer[:n])
			// Use the unified command handler
			go handlers.HandleCommand(conn, command)
		}
	}
}
