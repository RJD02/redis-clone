package replication

import (
	"fmt"
	"net"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// ReplicaClient handles the connection from replica to master
type ReplicaClient struct {
	masterHost  string
	masterPort  string
	replicaPort string
	conn        net.Conn
}

// NewReplicaClient creates a new replica client
func NewReplicaClient(masterHost, masterPort, replicaPort string) *ReplicaClient {
	return &ReplicaClient{
		masterHost:  masterHost,
		masterPort:  masterPort,
		replicaPort: replicaPort,
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

	// Wait for PONG response
	time.Sleep(50 * time.Millisecond)

	// Step 2: Send REPLCONF listening-port
	if err := r.sendReplconfListeningPort(); err != nil {
		return fmt.Errorf("failed to send REPLCONF listening-port: %v", err)
	}

	// Wait for OK response
	time.Sleep(50 * time.Millisecond)

	// Step 3: Send REPLCONF capa eof capa psync2
	if err := r.sendReplconfCapa(); err != nil {
		return fmt.Errorf("failed to send REPLCONF capa: %v", err)
	}

	// Wait for OK response
	time.Sleep(50 * time.Millisecond)

	// Step 4: Send PSYNC ? -1
	if err := r.sendPsync(); err != nil {
		return fmt.Errorf("failed to send PSYNC: %v", err)
	}

	// Wait for FULLRESYNC response
	time.Sleep(50 * time.Millisecond)

	fmt.Println("Replication handshake completed")
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

	// Start background response reader if not already started
	go r.readResponse()

	return nil
}

// sendReplconfListeningPort sends REPLCONF listening-port command to the master
func (r *ReplicaClient) sendReplconfListeningPort() error {
	// Create REPLCONF listening-port command as RESP array
	// *3\r\n$8\r\nREPLCONF\r\n$14\r\nlistening-port\r\n$<port_len>\r\n<port>\r\n
	replconfCommand := parser.RESPValue{
		Type: "array",
		Array: []parser.RESPValue{
			{Type: "bulk", Str: "REPLCONF"},
			{Type: "bulk", Str: "listening-port"},
			{Type: "bulk", Str: r.replicaPort},
		},
	}

	// Encode to RESP format
	respData := parser.EncodeRESP(replconfCommand)

	// Send to master
	_, err := r.conn.Write([]byte(respData))
	if err != nil {
		return fmt.Errorf("failed to write REPLCONF listening-port command: %v", err)
	}

	fmt.Printf("Sent REPLCONF listening-port to master: %s", respData)
	return nil
}

// sendReplconfCapa sends REPLCONF capa eof capa psync2 command to the master
func (r *ReplicaClient) sendReplconfCapa() error {
	// Create REPLCONF capa eof capa psync2 command as RESP array
	// *5\r\n$8\r\nREPLCONF\r\n$4\r\ncapa\r\n$3\r\neof\r\n$4\r\ncapa\r\n$6\r\npsync2\r\n
	replconfCommand := parser.RESPValue{
		Type: "array",
		Array: []parser.RESPValue{
			{Type: "bulk", Str: "REPLCONF"},
			{Type: "bulk", Str: "capa"},
			{Type: "bulk", Str: "eof"},
			{Type: "bulk", Str: "capa"},
			{Type: "bulk", Str: "psync2"},
		},
	}

	// Encode to RESP format
	respData := parser.EncodeRESP(replconfCommand)

	// Send to master
	_, err := r.conn.Write([]byte(respData))
	if err != nil {
		return fmt.Errorf("failed to write REPLCONF capa command: %v", err)
	}

	fmt.Printf("Sent REPLCONF capa eof capa psync2 to master: %s", respData)
	return nil
}

// sendPsync sends PSYNC ? -1 command to the master
func (r *ReplicaClient) sendPsync() error {
	// Create PSYNC ? -1 command as RESP array
	// *3\r\n$5\r\nPSYNC\r\n$1\r\n?\r\n$2\r\n-1\r\n
	psyncCommand := parser.RESPValue{
		Type: "array",
		Array: []parser.RESPValue{
			{Type: "bulk", Str: "PSYNC"},
			{Type: "bulk", Str: "?"},
			{Type: "bulk", Str: "-1"},
		},
	}

	// Encode to RESP format
	respData := parser.EncodeRESP(psyncCommand)

	// Send to master
	_, err := r.conn.Write([]byte(respData))
	if err != nil {
		return fmt.Errorf("failed to write PSYNC command: %v", err)
	}

	fmt.Printf("Sent PSYNC ? -1 to master: %s", respData)
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
