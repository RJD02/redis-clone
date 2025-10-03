package handlers

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// HandlePing handles the PING command
func HandlePing(conn net.Conn, cmd *Command) {
	response := "+PONG\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write PONG response")
		return
	}
	fmt.Println("PONG")
}

// HandleEcho handles the ECHO command
func HandleEcho(conn net.Conn, cmd *Command) {
	if len(cmd.Args) < 1 {
		response := "-ERR wrong number of arguments for 'echo' command\r\n"
		conn.Write([]byte(response))
		return
	}

	message := cmd.Args[0]
	response := parser.ToBulkString(message)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write ECHO response")
		return
	}
	fmt.Println("ECHO:", message)
}
