package handlers

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
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

	// Send empty RDB file after FULLRESYNC response
	// This is the hex representation of an empty RDB file
	emptyRDBHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

	// Decode hex to binary
	rdbData, err := hex.DecodeString(emptyRDBHex)
	if err != nil {
		fmt.Printf("Failed to decode RDB hex: %v\n", err)
		return
	}

	// Send RDB file in the format: $<length>\r\n<binary_contents>
	// Note: This is NOT a RESP bulk string, so no trailing \r\n
	rdbResponse := fmt.Sprintf("$%d\r\n", len(rdbData))
	_, err = conn.Write([]byte(rdbResponse))
	if err != nil {
		fmt.Printf("Failed to write RDB length: %v\n", err)
		return
	}

	// Send the binary RDB data
	_, err = conn.Write(rdbData)
	if err != nil {
		fmt.Printf("Failed to write RDB data: %v\n", err)
		return
	}

	fmt.Printf("Sent empty RDB file (%d bytes)\n", len(rdbData))

	// Register this connection as a replica for command propagation
	replication.Manager.AddReplica(conn)
}
