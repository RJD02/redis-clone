package config

import "fmt"

// ServerConfig holds all server configuration
type ServerConfig struct {
	// Server role and replication info
	Role       string
	MasterHost string
	MasterPort string
	IsReplica  bool

	// Network configuration
	Port string

	// Replication constants
	MasterReplId     string
	MasterReplOffset int
}

// Global server configuration instance
var Server = &ServerConfig{
	Role:             "master",
	Port:             "6379",
	MasterReplId:     "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
	MasterReplOffset: 0,
}

// SetServerRole sets the server's role (master or slave)
func SetServerRole(role string) {
	Server.Role = role
	fmt.Printf("Server role set to: %s\n", role)
}

// GetServerRole returns the current server role
func GetServerRole() string {
	return Server.Role
}

// SetReplicaConfig configures the server as a replica
func SetReplicaConfig(masterHost, masterPort string) {
	Server.IsReplica = true
	Server.MasterHost = masterHost
	Server.MasterPort = masterPort
	Server.Role = "slave"
}

// IsServerMaster returns true if the server is a master
func IsServerMaster() bool {
	return Server.Role == "master"
}
