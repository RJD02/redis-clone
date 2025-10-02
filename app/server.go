package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
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

		if command == "*1\r\n$4\r\nPING\r\n" {
			response := "+PONG\r\n"
			_, err = conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Failed to write PONG response")
				return
			}
			fmt.Println("PONG")
		} else if strings.HasPrefix(command, "*2\r\n$4\r\nECHO\r\n") {
			// Parse ECHO command: *2\r\n$4\r\nECHO\r\n$5\r\napple\r\n
			parts := strings.Split(command, "\r\n")
			if len(parts) >= 6 {
				// parts[4] contains the echoed message
				message := parts[4]
				response := "+" + message + "\r\n"
				_, err = conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to write ECHO response")
					return
				}
				fmt.Println("ECHO:", message)
			}
		}
	}
}
