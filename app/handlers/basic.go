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

// HandleReplconf handles the REPLCONF command
func HandleReplconf(conn net.Conn, cmd *Command) {
	// For now, REPLCONF always responds with +OK regardless of the arguments
	// In the future, we might want to handle specific REPLCONF subcommands
	response := "+OK\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write REPLCONF response")
		return
	}
	fmt.Printf("REPLCONF: %v\n", cmd.Args)
}

// HandlePsync handles the PSYNC command
func HandlePsync(conn net.Conn, cmd *Command) {
	// For now, PSYNC always responds with FULLRESYNC with a dummy replication ID and offset 0
	// In a real implementation, this would handle partial resync logic
	replId := "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb" // Example replication ID
	response := fmt.Sprintf("+FULLRESYNC %s 0\r\n", replId)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write PSYNC response")
		return
	}
	fmt.Printf("PSYNC: %v\n", cmd.Args)
}
