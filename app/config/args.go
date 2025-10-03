package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// CLIConfig holds command line configuration
type CLIConfig struct {
	Port       string
	IsReplica  bool
	MasterHost string
	MasterPort string
}

// ParseArgs parses command line arguments and returns CLIConfig
func ParseArgs() (*CLIConfig, error) {
	config := &CLIConfig{
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
