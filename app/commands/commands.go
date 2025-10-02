package commands

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

var Dictionary = make(map[string]string)

func HandlePing(conn net.Conn) {

	response := "+PONG\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write PONG response")
		return
	}
	fmt.Println("PONG")
}

func HandleEcho(conn net.Conn, command string) {

	// Parse ECHO command: *2\r\n$4\r\nECHO\r\n$5\r\napple\r\n
	parts := strings.Split(command, "\r\n")
	if len(parts) >= 6 {
		// parts[4] contains the echoed message
		message := parts[4]
		response := parser.ToBulkString(message)
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Failed to write ECHO response")
			return
		}
		fmt.Println("ECHO:", message)
	}
}

func HandleSet(conn net.Conn, command string) {

	response := "+OK\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write SET response")
		return
	}
	// add to the dictionary
	parts := strings.Split(command, "\r\n")
	if len(parts) >= 8 {
		key := parts[4]
		value := parts[6]
		Dictionary[key] = value
		fmt.Println("SET:", key, "=", value)
	}
	fmt.Println("SET command received")
}

func HandleGet(conn net.Conn, command string) {
	// get from dictionary
	parts := strings.Split(command, "\r\n")
	if len(parts) >= 6 {
		key := parts[4]
		value, exists := Dictionary[key]
		var response string
		if exists {
			response = parser.ToBulkString(value)
			fmt.Println("GET:", key, "=", value)
		} else {
			response = "$-1\r\n" // RESP Null Bulk String
			fmt.Println("GET:", key, "= (not found)")
		}
		_, err := conn.Write([]byte(response))
		if err != nil {
			fmt.Println("Failed to write GET response")
			return
		}
	}
	fmt.Println("GET command received")
}
