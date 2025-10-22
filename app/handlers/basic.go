package handlers

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/config"
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
	// Get the replication ID and offset from server configuration
	replId := config.Server.MasterReplId
	replOffset := config.Server.MasterReplOffset

	// Respond with FULLRESYNC using the actual server configuration
	response := fmt.Sprintf("+FULLRESYNC %s %d\r\n", replId, replOffset)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write PSYNC response")
		return
	}
	fmt.Printf("PSYNC: %v -> FULLRESYNC %s %d\n", cmd.Args, replId, replOffset)
}
