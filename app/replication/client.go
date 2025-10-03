package replication

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// ReplicaClient handles the connection from replica to master
type ReplicaClient struct {
	masterHost string
	masterPort string
	conn       net.Conn
}

// NewReplicaClient creates a new replica client
func NewReplicaClient(masterHost, masterPort string) *ReplicaClient {
	return &ReplicaClient{
		masterHost: masterHost,
		masterPort: masterPort,
	}
}

// Connect establishes connection to the master server
func (r *ReplicaClient) Connect() error {
	address := r.masterHost + ":" + r.masterPort
	fmt.Printf("Connecting to master at %s\n", address)

	var err error
	r.conn, err = net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %v", err)
	}

	fmt.Printf("Connected to master at %s\n", address)
	return nil
}

// StartHandshake initiates the replication handshake with the master
func (r *ReplicaClient) StartHandshake() error {
	if r.conn == nil {
		return fmt.Errorf("not connected to master")
	}

	// Step 1: Send PING command as RESP array
	if err := r.sendPing(); err != nil {
		return fmt.Errorf("failed to send PING: %v", err)
	}

	fmt.Println("Replication handshake initiated with PING")
	return nil
}

// sendPing sends a PING command to the master
func (r *ReplicaClient) sendPing() error {
	// Create PING command as RESP array: *1\r\n$4\r\nPING\r\n
	pingCommand := parser.RESPValue{
		Type: "array",
		Array: []parser.RESPValue{
			{Type: "bulk", Str: "PING"},
		},
	}

	// Encode to RESP format
	respData := parser.EncodeRESP(pingCommand)

	// Send to master
	_, err := r.conn.Write([]byte(respData))
	if err != nil {
		return fmt.Errorf("failed to write PING command: %v", err)
	}

	fmt.Printf("Sent PING to master: %s", respData)

	// Read response (optional for this stage, but good practice)
	go r.readResponse()

	return nil
}

// readResponse reads responses from master (runs in background)
func (r *ReplicaClient) readResponse() {
	buffer := make([]byte, 1024)
	for {
		n, err := r.conn.Read(buffer)
		if err != nil {
			fmt.Printf("Master connection closed: %v\n", err)
			return
		}

		if n > 0 {
			response := string(buffer[:n])
			fmt.Printf("Received from master: %s", response)
		}
	}
}

// Close closes the connection to master
func (r *ReplicaClient) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
