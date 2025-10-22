package handlers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/repository"
)

// ReplicaCommandHandler handles commands from master without sending responses
type ReplicaCommandHandler struct {
	dataHandler *DataHandler
	conn        net.Conn // Connection to master for sending ACK responses
	offset      int      // Tracks total bytes of commands processed from master
}

// NewReplicaCommandHandler creates a new replica command handler with repository dependency
func NewReplicaCommandHandler(repo repository.KeyValueRepository) *ReplicaCommandHandler {
	return &ReplicaCommandHandler{
		dataHandler: NewDataHandler(repo),
		conn:        nil, // Will be set later via SetConnection
		offset:      0,   // Start with 0 offset
	}
}

// SetConnection sets the connection to master (used for sending ACK responses)
func (rch *ReplicaCommandHandler) SetConnection(conn net.Conn) {
	rch.conn = conn
}

// ProcessCommand processes a command from master without sending any response
func (rch *ReplicaCommandHandler) ProcessCommand(respData string) error {
	cmd, err := ParseCommand(respData)
	if err != nil {
		fmt.Printf("Replica command parse error: %v\n", err)
		return err
	}

	fmt.Printf("Replica processing command: %s, Args: %v\n", cmd.Name, cmd.Args)

	// Process commands without sending responses back, except for REPLCONF GETACK
	switch cmd.Name {
	case "SET":
		// Use a nil connection to indicate no response should be sent
		return rch.processSilentSet(cmd)
	case "REPLCONF":
		// Handle REPLCONF commands (specifically GETACK)
		return rch.processReplconf(cmd)
	case "DEL":
		// Could add delete command handling here
		fmt.Printf("Replica processed DEL command (not yet implemented)\n")
	default:
		fmt.Printf("Replica received unknown command: %s\n", cmd.Name)
	}

	return nil
}

// processSilentSet processes a SET command without sending any response
func (rch *ReplicaCommandHandler) processSilentSet(cmd *Command) error {
	if len(cmd.Args) < 2 {
		return fmt.Errorf("wrong number of arguments for 'set' command")
	}

	key := cmd.Args[0]
	value := cmd.Args[1]

	// Parse expiration options if present (same logic as regular SET)
	var expiration *time.Duration
	for i := 2; i < len(cmd.Args); i += 2 {
		if i+1 >= len(cmd.Args) {
			return fmt.Errorf("syntax error in SET command")
		}

		option := strings.ToUpper(cmd.Args[i])
		timeStr := cmd.Args[i+1]

		timeVal, err := strconv.Atoi(timeStr)
		if err != nil {
			return fmt.Errorf("invalid expiration time: %v", err)
		}

		switch option {
		case "EX":
			// Expiration in seconds
			duration := time.Duration(timeVal) * time.Second
			expiration = &duration
		case "PX":
			// Expiration in milliseconds
			duration := time.Duration(timeVal) * time.Millisecond
			expiration = &duration
		default:
			return fmt.Errorf("unknown SET option: %s", option)
		}
	}

	err := rch.dataHandler.repo.Set(key, value, expiration)
	if err != nil {
		return fmt.Errorf("failed to set key in replica: %v", err)
	}

	if expiration != nil {
		fmt.Printf("Replica SET: %s = %s (expires in %v)\n", key, value, *expiration)
	} else {
		fmt.Printf("Replica SET: %s = %s (no expiration)\n", key, value)
	}
	return nil
}

// processReplconf handles REPLCONF commands from master
func (rch *ReplicaCommandHandler) processReplconf(cmd *Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("wrong number of arguments for 'replconf' command")
	}

	subcommand := strings.ToUpper(cmd.Args[0])

	switch subcommand {
	case "GETACK":
		// REPLCONF GETACK * - master is requesting acknowledgment
		// We should respond with REPLCONF ACK <offset>
		return rch.sendAck()
	default:
		fmt.Printf("Replica received unknown REPLCONF subcommand: %s\n", subcommand)
	}

	return nil
}

// sendAck sends REPLCONF ACK response back to master
func (rch *ReplicaCommandHandler) sendAck() error {
	if rch.conn == nil {
		return fmt.Errorf("no connection to master for sending ACK")
	}

	// For now, hardcode offset to 0 as specified in the requirements
	offset := "0"

	// Create REPLCONF ACK response: *3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n
	ackCommand := parser.RESPValue{
		Type: "array",
		Array: []parser.RESPValue{
			{Type: "bulk", Str: "REPLCONF"},
			{Type: "bulk", Str: "ACK"},
			{Type: "bulk", Str: offset},
		},
	}

	// Encode to RESP format
	respData := parser.EncodeRESP(ackCommand)

	// Send to master
	_, err := rch.conn.Write([]byte(respData))
	if err != nil {
		return fmt.Errorf("failed to write REPLCONF ACK response: %v", err)
	}

	fmt.Printf("Sent REPLCONF ACK %s to master: %s", offset, respData)
	return nil
}
