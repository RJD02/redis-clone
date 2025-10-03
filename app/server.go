package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port       string
	IsReplica  bool
	MasterHost string
	MasterPort string
}

// parseArgs parses command line arguments and returns ServerConfig
func parseArgs() (*ServerConfig, error) {
	config := &ServerConfig{
		Port:      "6379", // Default port
		IsReplica: false,
	}

	args := os.Args[1:] // Skip program name

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--port requires a value")
			}
			config.Port = args[i+1]
			i++ // Skip the next argument since we consumed it

		case "--replicaof":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--replicaof requires a value")
			}
			// Parse master host and port from the argument
			// Format: "hostname port" or hostname port (with or without quotes)
			replicaOfArg := args[i+1]
			parts := strings.Fields(replicaOfArg)
			if len(parts) != 2 {
				return nil, fmt.Errorf("--replicaof format should be 'host port', got: %s", replicaOfArg)
			}
			config.MasterHost = parts[0]
			config.MasterPort = parts[1]
			config.IsReplica = true
			i++ // Skip the next argument since we consumed it

		default:
			return nil, fmt.Errorf("unknown argument: %s", args[i])
		}
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(config.Port); err != nil {
		return nil, fmt.Errorf("invalid port number: %s", config.Port)
	}

	// Validate master port if replica
	if config.IsReplica {
		if _, err := strconv.Atoi(config.MasterPort); err != nil {
			return nil, fmt.Errorf("invalid master port number: %s", config.MasterPort)
		}
	}

	return config, nil
}

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

// toBulkString formats a string as a Redis RESP bulk string
// Format: $<length>\r\n<data>\r\n

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Parse command line arguments
	config, err := parseArgs()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Set the server role in the commands package
	if config.IsReplica {
		commands.SetServerRole("slave")
		fmt.Printf("Configured as replica of %s:%s\n", config.MasterHost, config.MasterPort)
	} else {
		commands.SetServerRole("master")
		fmt.Println("Configured as master")
	}

	address := "0.0.0.0:" + config.Port
	fmt.Printf("Starting Redis server on %s\n", address)

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Failed to bind to port %s: %v\n", config.Port, err)
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

		// Use the unified command handler
		go commands.HandleCommand(conn, command)
	}
}
