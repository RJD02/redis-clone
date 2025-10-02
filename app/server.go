package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

var (
	ECHO_COMMAND_PREFIX = "*2\r\n$4\r\nECHO\r\n"
	PING_COMMAND        = "*1\r\n$4\r\nPING\r\n"
	SET_COMMAND_PREFIX  = "*3\r\n$3\r\nSET\r\n"
	GET_COMMAND_PREFIX  = "*2\r\n$3\r\nGET\r\n"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

// toBulkString formats a string as a Redis RESP bulk string
// Format: $<length>\r\n<data>\r\n

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

		if command == PING_COMMAND {
			go commands.HandlePing(conn)
		} else if strings.HasPrefix(command, ECHO_COMMAND_PREFIX) {
			go commands.HandleEcho(conn, command)
		} else if strings.HasPrefix(command, SET_COMMAND_PREFIX) {
			go commands.HandleSet(conn, command)
		} else if strings.HasPrefix(command, GET_COMMAND_PREFIX) {
			go commands.HandleGet(conn, command)
		}
	}
}
