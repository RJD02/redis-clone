package server

import (
	"fmt"
	"net"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/config"
	"github.com/codecrafters-io/redis-starter-go/app/replication"
)

// RedisServer represents the Redis server instance
type RedisServer struct {
	listener net.Listener
	address  string
}

// NewRedisServer creates a new Redis server instance
func NewRedisServer(port string) (*RedisServer, error) {
	address := "0.0.0.0:" + port

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to bind to port %s: %v", port, err)
	}

	return &RedisServer{
		listener: listener,
		address:  address,
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

		go HandleConnection(conn)
	}
}

// Close closes the server listener
func (s *RedisServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Initialize sets up the server configuration based on CLI args
func Initialize(cliConfig *config.CLIConfig) {
	if cliConfig.IsReplica {
		config.SetServerRole("slave")
		config.SetReplicaConfig(cliConfig.MasterHost, cliConfig.MasterPort)
		fmt.Printf("Configured as replica of %s:%s\n", cliConfig.MasterHost, cliConfig.MasterPort)

		// Start replication handshake in background
		go startReplicationHandshake(cliConfig.MasterHost, cliConfig.MasterPort)
	} else {
		config.SetServerRole("master")
		fmt.Println("Configured as master")
	}
}

// startReplicationHandshake initiates the handshake with master server
func startReplicationHandshake(masterHost, masterPort string) {
	// Give the server a moment to start up
	time.Sleep(100 * time.Millisecond)

	client := replication.NewReplicaClient(masterHost, masterPort)

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
