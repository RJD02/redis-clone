package replication

import (
	"fmt"
	"net"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// ReplicaConnection represents a connection to a replica
type ReplicaConnection struct {
	Conn net.Conn
	ID   string // Could be the remote address
}

// ReplicaManager manages all replica connections
type ReplicaManager struct {
	replicas map[string]*ReplicaConnection
	mu       sync.RWMutex
}

// Global replica manager instance
var Manager = NewReplicaManager()

// NewReplicaManager creates a new replica manager
func NewReplicaManager() *ReplicaManager {
	return &ReplicaManager{
		replicas: make(map[string]*ReplicaConnection),
	}
}

// AddReplica adds a new replica connection
func (rm *ReplicaManager) AddReplica(conn net.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	id := conn.RemoteAddr().String()
	rm.replicas[id] = &ReplicaConnection{
		Conn: conn,
		ID:   id,
	}
	fmt.Printf("Added replica: %s\n", id)
}

// RemoveReplica removes a replica connection
func (rm *ReplicaManager) RemoveReplica(conn net.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	id := conn.RemoteAddr().String()
	delete(rm.replicas, id)
	fmt.Printf("Removed replica: %s\n", id)
}

// PropagateCommand sends a command to all replicas
func (rm *ReplicaManager) PropagateCommand(commandName string, args []string) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if len(rm.replicas) == 0 {
		return // No replicas to propagate to
	}

	// Create RESP array for the command
	respArray := parser.RESPValue{
		Type: "array",
		Array: make([]parser.RESPValue, len(args)+1),
	}

	// First element is the command name
	respArray.Array[0] = parser.RESPValue{Type: "bulk", Str: commandName}

	// Rest are the arguments
	for i, arg := range args {
		respArray.Array[i+1] = parser.RESPValue{Type: "bulk", Str: arg}
	}

	// Encode to RESP format
	respData := parser.EncodeRESP(respArray)

	// Send to all replicas
	for id, replica := range rm.replicas {
		_, err := replica.Conn.Write([]byte(respData))
		if err != nil {
			fmt.Printf("Failed to propagate command to replica %s: %v\n", id, err)
			// Remove failed replica (will be cleaned up in next propagation)
			go rm.RemoveReplica(replica.Conn)
		} else {
			fmt.Printf("Propagated command %s to replica %s\n", commandName, id)
		}
	}
}

// GetReplicaCount returns the number of connected replicas
func (rm *ReplicaManager) GetReplicaCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.replicas)
}