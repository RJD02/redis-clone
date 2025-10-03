package handlers

import (
	"fmt"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// Command represents a parsed Redis command
type Command struct {
	Name string   // Command name (PING, SET, GET, etc.)
	Args []string // Command arguments
}

// ParseCommand parses a RESP command into a Command structure
func ParseCommand(respData string) (*Command, error) {
	parsed, err := parser.ParseRESP(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RESP: %v", err)
	}

	if parsed.Type != "array" || len(parsed.Array) == 0 {
		return nil, fmt.Errorf("invalid command format")
	}

	// First element is the command name
	commandName := strings.ToUpper(parsed.Array[0].Str)

	// Rest are arguments
	args := make([]string, len(parsed.Array)-1)
	for i := 1; i < len(parsed.Array); i++ {
		args[i-1] = parsed.Array[i].Str
	}

	return &Command{
		Name: commandName,
		Args: args,
	}, nil
}

// HandleCommand routes commands to appropriate handlers
func HandleCommand(conn net.Conn, respData string) {
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
		HandleSet(conn, cmd)
	case "GET":
		HandleGet(conn, cmd)
	case "INFO":
		HandleInfo(conn, cmd)
	default:
		response := "-ERR unknown command '" + cmd.Name + "'\r\n"
		conn.Write([]byte(response))
		fmt.Printf("Unknown command: %s\n", cmd.Name)
	}
}
