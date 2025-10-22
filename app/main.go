package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
	"github.com/codecrafters-io/redis-starter-go/app/repository"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// RedisServer represents the Redis server instance
type RedisServer struct {
	listener          net.Listener
	address           string
	connectionHandler *ConnectionHandler
	repository        repository.KeyValueRepository
}

// ConnectionHandler handles connections with injected dependencies
type ConnectionHandler struct {
	handlerManager *handlers.HandlerManager
}

// NewConnectionHandler creates a new connection handler with dependencies
func NewConnectionHandler(handlerManager *handlers.HandlerManager) *ConnectionHandler {
	return &ConnectionHandler{
		handlerManager: handlerManager,
	}
}

// HandleConnection handles a single client connection
func (ch *ConnectionHandler) HandleConnection(conn net.Conn) {
	defer func() {
		// Clean up replica connection if it was registered
		replication.Manager.RemoveReplica(conn)
		conn.Close()
	}()

	fmt.Printf("New connection from %s\n", conn.RemoteAddr())

	for {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Printf("Connection closed: %s (%v)\n", conn.RemoteAddr(), err)
			return
		}

		// Only process if we received data
		if n > 0 {
			command := string(buffer[:n])
			// Use the handler manager with dependency injection
			go ch.handlerManager.HandleCommand(conn, command)
		}
	}
}

// NewRedisServer creates a new Redis server instance
func NewRedisServer(port string) (*RedisServer, error) {
	address := "0.0.0.0:" + port

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to port %s: %v", port, err)
	}

	// Create repository with existing global dictionary for backward compatibility
	repo := repository.NewMemoryRepositoryWithStorage(storage.Dictionary)

	// Create handler manager with dependencies
	handlerManager := handlers.NewHandlerManager(repo)

	// Create connection handler
	connectionHandler := NewConnectionHandler(handlerManager)

	return &RedisServer{
		listener:          listener,
		address:           address,
		connectionHandler: connectionHandler,
		repository:        repo,
	}, nil
}

// Start starts the Redis server and accepts connections
func (s *RedisServer) Start() error {
	fmt.Printf("Starting Redis server on %s\n", s.address)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}

		go s.connectionHandler.HandleConnection(conn)
	}
}

// Close closes the server listener
func (s *RedisServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// GetRepository returns the server's repository (for testing or other uses)
func (s *RedisServer) GetRepository() repository.KeyValueRepository {
	return s.repository
}

// Initialize sets up the server configuration based on CLI args
func Initialize(cliConfig *config.CLIConfig) {
	if cliConfig.IsReplica {
		config.SetServerRole("slave")
		config.SetReplicaConfig(cliConfig.MasterHost, cliConfig.MasterPort)
		fmt.Printf("Configured as replica of %s:%s\n", cliConfig.MasterHost, cliConfig.MasterPort)

		// Start replication handshake in background
		go startReplicationHandshake(cliConfig.MasterHost, cliConfig.MasterPort, cliConfig.Port)
	} else {
		config.SetServerRole("master")
		fmt.Println("Configured as master")
	}
}

// startReplicationHandshake initiates the handshake with master server
func startReplicationHandshake(masterHost, masterPort, replicaPort string) {
	// Give the server a moment to start up
	time.Sleep(100 * time.Millisecond)

	client := replication.NewReplicaClient(masterHost, masterPort, replicaPort)

	// Connect to master
	if err := client.Connect(); err != nil {
		fmt.Printf("Failed to connect to master: %v\n", err)
		return
	}

	// Start handshake
	if err := client.StartHandshake(); err != nil {
		fmt.Printf("Failed to start handshake: %v\n", err)
		return
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	// Parse command line arguments
	cliConfig, err := config.ParseArgs()
	if err != nil {
		fmt.Printf("Error parsing arguments: %v\n", err)
		os.Exit(1)
	}

	// Initialize server configuration
	Initialize(cliConfig)

	// Create and start the Redis server
	redisServer, err := NewRedisServer(cliConfig.Port)
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
