package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	fmt.Println("Logs from your program will appear here!")

	// Parse command line arguments
	cliConfig, err := config.ParseArgs()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Initialize server configuration
	server.Initialize(cliConfig)

	// Create and start the Redis server
	redisServer, err := server.NewRedisServer(cliConfig.Port)
	if err != nil {
		fmt.Printf("Error creating server: %v\n", err)
		os.Exit(1)
	}

	// Start the server (this blocks)
	if err := redisServer.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
