package replication

import (
	"fmt"
	"net"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// CommandProcessor defines a function type for processing commands from master
type CommandProcessor func(respData string) error

// ConnectionSetter defines a function for setting the connection on command processor
type ConnectionSetter func(conn net.Conn)

// ReplicaClient handles the connection from replica to master
type ReplicaClient struct {
	masterHost       string
	masterPort       string
	replicaPort      string
	conn             net.Conn
	commandProcessor CommandProcessor
	connectionSetter ConnectionSetter // New field to set connection on command processor
	buffer           []byte           // Buffer for accumulating partial data
	rdbReceived      bool             // Flag to track if RDB file has been fully received
}

// NewReplicaClient creates a new replica client
func NewReplicaClient(masterHost, masterPort, replicaPort string, processor CommandProcessor, setter ConnectionSetter) *ReplicaClient {
	return &ReplicaClient{
		masterHost:       masterHost,
		masterPort:       masterPort,
		replicaPort:      replicaPort,
		commandProcessor: processor,
		connectionSetter: setter,
		buffer:           make([]byte, 0, 4096),
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

	// Set the connection on the command processor so it can send ACK responses
	if r.connectionSetter != nil {
		r.connectionSetter(r.conn)
	}

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
	readBuffer := make([]byte, 1024)
	for {
		n, err := r.conn.Read(readBuffer)
		if err != nil {
			fmt.Printf("Master connection closed: %v\n", err)
			return
		}

		if n > 0 {
			// Accumulate data in buffer
			r.buffer = append(r.buffer, readBuffer[:n]...)

			// Process complete commands from the buffer
			r.processBufferedCommands()
		}
	}
}

// processBufferedCommands processes complete RESP commands from the buffer
func (r *ReplicaClient) processBufferedCommands() {
	data := string(r.buffer)

	// First, process any handshake responses (+PONG, +OK, +FULLRESYNC)
	for len(data) > 0 && (data[0] == '+' || data[0] == '-' || data[0] == ':') {
		complete, consumed := r.isCompleteRESPMessage(data)
		if !complete {
			break
		}

		message := data[:consumed]
		fmt.Printf("Received from master: %s", message)

		data = data[consumed:]
		r.buffer = r.buffer[consumed:]
	}

	// Handle RDB file if not yet received
	if !r.rdbReceived && len(data) > 0 {
		consumed := r.handleRDBFile(data)
		if consumed > 0 {
			data = data[consumed:]
			r.buffer = r.buffer[consumed:]
			r.rdbReceived = true
		} else if data[0] == '$' {
			// We see the start of RDB but don't have it all yet
			return
		}
		// If it doesn't start with $, we might not have the RDB start yet
	}

	// Process commands after RDB file
	for len(data) > 0 {
		// Check if we have a complete RESP message
		complete, consumed := r.isCompleteRESPMessage(data)
		if !complete {
			// Not enough data for a complete message, wait for more
			break
		}

		// Extract the complete message
		message := data[:consumed]

		// Check if this looks like a command (starts with *) vs a response (+, -, :, $)
		if len(message) > 0 && message[0] == '*' {
			// This is a command array, process it
			if r.commandProcessor != nil {
				err := r.commandProcessor(message)
				if err != nil {
					fmt.Printf("Error processing replica command: %v\n", err)
				}
			}
		} else {
			// This is a response (PONG, OK, FULLRESYNC, etc.), just log it
			fmt.Printf("Received from master: %s", message)
		}

		// Remove processed data from buffer
		data = data[consumed:]
		r.buffer = r.buffer[consumed:]
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleRDBFile handles the RDB file reception and returns how many bytes were consumed
func (r *ReplicaClient) handleRDBFile(data string) int {
	// Look for the RDB file marker (starts with $)
	if len(data) == 0 || data[0] != '$' {
		return 0
	}

	// Find the end of the length specification
	firstCRLF := findCRLF(data)
	if firstCRLF == -1 {
		return 0 // Not enough data yet
	}

	// Parse the RDB file length
	lengthStr := data[1:firstCRLF]
	length := 0
	for _, char := range lengthStr {
		if char >= '0' && char <= '9' {
			length = length*10 + int(char-'0')
		} else {
			// Invalid length
			return 0
		}
	}

	// Check if we have the complete RDB file
	headerLen := firstCRLF + 2
	totalNeeded := headerLen + length
	if len(data) < totalNeeded {
		return 0 // Not enough data yet
	}

	// We have the complete RDB file, consume it
	return totalNeeded
}

// isCompleteRESPMessage checks if we have a complete RESP message and returns how many bytes it consumes
func (r *ReplicaClient) isCompleteRESPMessage(data string) (complete bool, consumed int) {
	if len(data) == 0 {
		return false, 0
	}

	switch data[0] {
	case '+', '-', ':': // Simple string, error, integer
		end := findCRLF(data)
		if end == -1 {
			return false, 0
		}
		return true, end + 2

	case '$': // Bulk string (but we should have handled RDB already)
		return r.isCompleteBulkString(data)

	case '*': // Array
		return r.isCompleteArray(data)

	default:
		// Unknown type, skip this byte and continue
		return true, 1
	}
}

// findCRLF finds the first occurrence of \r\n in data
func findCRLF(data string) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			return i
		}
	}
	return -1
}

// isCompleteBulkString checks if we have a complete bulk string
func (r *ReplicaClient) isCompleteBulkString(data string) (bool, int) {
	firstCRLF := findCRLF(data)
	if firstCRLF == -1 {
		return false, 0
	}

	// Parse length
	lengthStr := data[1:firstCRLF]
	length := 0
	for _, char := range lengthStr {
		if char >= '0' && char <= '9' {
			length = length*10 + int(char-'0')
		} else if char == '-' && length == 0 {
			// Null bulk string
			return true, firstCRLF + 2
		} else {
			return false, 0
		}
	}

	// Check if we have enough data for the content + final CRLF
	needed := firstCRLF + 2 + length + 2
	if len(data) < needed {
		return false, 0
	}

	return true, needed
}

// isCompleteArray checks if we have a complete array
func (r *ReplicaClient) isCompleteArray(data string) (bool, int) {
	firstCRLF := findCRLF(data)
	if firstCRLF == -1 {
		return false, 0
	}

	// Parse array count
	countStr := data[1:firstCRLF]
	count := 0
	for _, char := range countStr {
		if char >= '0' && char <= '9' {
			count = count*10 + int(char-'0')
		} else if char == '-' && count == 0 {
			// Null array
			return true, firstCRLF + 2
		} else {
			return false, 0
		}
	}

	// Process each element in the array
	consumed := firstCRLF + 2
	remaining := data[consumed:]

	for i := 0; i < count; i++ {
		if len(remaining) == 0 {
			return false, 0
		}

		// For now, only handle bulk string elements (most common in commands)
		if remaining[0] == '$' {
			complete, elementConsumed := r.isCompleteBulkString(remaining)
			if !complete {
				return false, 0
			}
			consumed += elementConsumed
			remaining = remaining[elementConsumed:]
		} else {
			// Handle other types if needed
			return false, 0
		}
	}

	return true, consumed
}

// Close closes the connection to master
func (r *ReplicaClient) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
