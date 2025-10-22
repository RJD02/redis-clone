package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/repository"
)

// ReplicaCommandHandler handles commands from master without sending responses
type ReplicaCommandHandler struct {
	dataHandler *DataHandler
}

// NewReplicaCommandHandler creates a new replica command handler with repository dependency
func NewReplicaCommandHandler(repo repository.KeyValueRepository) *ReplicaCommandHandler {
	return &ReplicaCommandHandler{
		dataHandler: NewDataHandler(repo),
	}
}

// ProcessCommand processes a command from master without sending any response
func (rch *ReplicaCommandHandler) ProcessCommand(respData string) error {
	cmd, err := ParseCommand(respData)
	if err != nil {
		fmt.Printf("Replica command parse error: %v\n", err)
		return err
	}

	fmt.Printf("Replica processing command: %s, Args: %v\n", cmd.Name, cmd.Args)

	// Process commands without sending responses back
	switch cmd.Name {
	case "SET":
		// Use a nil connection to indicate no response should be sent
		return rch.processSilentSet(cmd)
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