package handlers

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/repository"
)

// HandlerManager manages all command handlers with their dependencies
type HandlerManager struct {
	dataHandler *DataHandler
}

// NewHandlerManager creates a new handler manager with all dependencies
func NewHandlerManager(repo repository.KeyValueRepository) *HandlerManager {
	return &HandlerManager{
		dataHandler: NewDataHandler(repo),
	}
}

// HandleCommand routes commands to appropriate handlers with dependency injection
func (hm *HandlerManager) HandleCommand(conn net.Conn, respData string) {
	cmd, err := ParseCommand(respData)
	if err != nil {
		response := "-ERR " + err.Error() + "\r\n"
		conn.Write([]byte(response))
		fmt.Printf("Command parse error: %v\n", err)
		return
	}

	fmt.Printf("Command: %s, Args: %v\n", cmd.Name, cmd.Args)

	switch cmd.Name {
	case "PING":
		HandlePing(conn, cmd)
	case "ECHO":
		HandleEcho(conn, cmd)
	case "SET":
		hm.dataHandler.HandleSet(conn, cmd)
	case "GET":
		hm.dataHandler.HandleGet(conn, cmd)
	case "INFO":
		HandleInfo(conn, cmd)
	case "REPLCONF":
		HandleReplconf(conn, cmd)
	case "PSYNC":
		HandlePsync(conn, cmd)
	default:
		response := "-ERR unknown command '" + cmd.Name + "'\r\n"
		conn.Write([]byte(response))
		fmt.Printf("Unknown command: %s\n", cmd.Name)
	}
}
