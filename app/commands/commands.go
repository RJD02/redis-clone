package commands

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
)

// Command represents a parsed Redis command
type Command struct {
	Name string   // Command name (PING, SET, GET, etc.)
	Args []string // Command arguments
}

// ParseCommand parses a RESP command into a Command structure
func ParseCommand(respData string) (*Command, error) {
	parsed, err := parser.ParseRESP(respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RESP: %v", err)
	}

	if parsed.Type != "array" || len(parsed.Array) == 0 {
		return nil, fmt.Errorf("invalid command format")
	}

	// First element is the command name
	commandName := strings.ToUpper(parsed.Array[0].Str)

	// Rest are arguments
	args := make([]string, len(parsed.Array)-1)
	for i := 1; i < len(parsed.Array); i++ {
		args[i-1] = parsed.Array[i].Str
	}

	return &Command{
		Name: commandName,
		Args: args,
	}, nil
}

// HandleCommand routes commands to appropriate handlers
func HandleCommand(conn net.Conn, respData string) {
	cmd, err := ParseCommand(respData)
	if err != nil {
		response := "-ERR " + err.Error() + "\r\n"
		conn.Write([]byte(response))
		fmt.Printf("Command parse error: %v\n", err)
		return
	}

	fmt.Printf("Command: %s, Args: %v\n", cmd.Name, cmd.Args)

	switch cmd.Name {
	case "PING":
		HandlePing(conn, cmd)
	case "ECHO":
		HandleEcho(conn, cmd)
	case "SET":
		HandleSet(conn, cmd)
	case "GET":
		HandleGet(conn, cmd)
	case "INFO":
		HandleInfo(conn, cmd)
	default:
		response := "-ERR unknown command '" + cmd.Name + "'\r\n"
		conn.Write([]byte(response))
		fmt.Printf("Unknown command: %s\n", cmd.Name)
	}
}

// KeyValue represents a value with expiration
type KeyValue struct {
	Value     string
	ExpiresAt *time.Time
}

// ExpiringDict is a thread-safe dictionary with expiration support
type ExpiringDict struct {
	data map[string]KeyValue
	mu   sync.RWMutex
}

// NewExpiringDict creates a new expiring dictionary
func NewExpiringDict() *ExpiringDict {
	return &ExpiringDict{
		data: make(map[string]KeyValue),
	}
}

// Set stores a key-value pair with optional expiration
func (ed *ExpiringDict) Set(key, value string, expiration *time.Duration) {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	kv := KeyValue{Value: value}
	if expiration != nil {
		expiresAt := time.Now().Add(*expiration)
		kv.ExpiresAt = &expiresAt

		// Start a goroutine to delete the key after expiration
		go func(k string, expireTime time.Time) {
			time.Sleep(time.Until(expireTime))
			ed.Delete(k)
		}(key, expiresAt)
	}

	ed.data[key] = kv
}

// Get retrieves a value by key, checking expiration
func (ed *ExpiringDict) Get(key string) (string, bool) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	kv, exists := ed.data[key]
	if !exists {
		return "", false
	}

	// Check if expired
	if kv.ExpiresAt != nil && time.Now().After(*kv.ExpiresAt) {
		// Key has expired, delete it
		go ed.Delete(key)
		return "", false
	}

	return kv.Value, true
}

// Delete removes a key from the dictionary
func (ed *ExpiringDict) Delete(key string) {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	delete(ed.data, key)
}

var Dictionary = NewExpiringDict()

// Global server configuration
var serverRole = "master" // Default to master

// SetServerRole sets the server's role (master or slave)
func SetServerRole(role string) {
	serverRole = role
	fmt.Printf("Server role set to: %s\n", role)
}

// GetServerRole returns the current server role
func GetServerRole() string {
	return serverRole
}

func HandlePing(conn net.Conn, cmd *Command) {
	response := "+PONG\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write PONG response")
		return
	}
	fmt.Println("PONG")
}

func HandleEcho(conn net.Conn, cmd *Command) {
	if len(cmd.Args) < 1 {
		response := "-ERR wrong number of arguments for 'echo' command\r\n"
		conn.Write([]byte(response))
		return
	}

	message := cmd.Args[0]
	response := parser.ToBulkString(message)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write ECHO response")
		return
	}
	fmt.Println("ECHO:", message)
}

func HandleSet(conn net.Conn, cmd *Command) {
	if len(cmd.Args) < 2 {
		response := "-ERR wrong number of arguments for 'set' command\r\n"
		conn.Write([]byte(response))
		return
	}

	key := cmd.Args[0]
	value := cmd.Args[1]

	var expiration *time.Duration

	// Parse expiration options: EX seconds or PX milliseconds
	for i := 2; i < len(cmd.Args); i += 2 {
		if i+1 >= len(cmd.Args) {
			response := "-ERR syntax error\r\n"
			conn.Write([]byte(response))
			return
		}

		option := strings.ToUpper(cmd.Args[i])
		timeStr := cmd.Args[i+1]

		timeVal, err := strconv.Atoi(timeStr)
		if err != nil {
			response := "-ERR invalid expiration time\r\n"
			conn.Write([]byte(response))
			return
		}

		switch option {
		case "EX":
			// Expiration in seconds
			duration := time.Duration(timeVal) * time.Second
			expiration = &duration
		case "PX":
			// Expiration in milliseconds
			duration := time.Duration(timeVal) * time.Millisecond
			expiration = &duration
		default:
			response := "-ERR syntax error\r\n"
			conn.Write([]byte(response))
			return
		}
	}

	// Store in dictionary
	Dictionary.Set(key, value, expiration)

	response := "+OK\r\n"
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write SET response")
		return
	}

	if expiration != nil {
		fmt.Printf("SET: %s = %s (expires in %v)\n", key, value, *expiration)
	} else {
		fmt.Printf("SET: %s = %s (no expiration)\n", key, value)
	}
}

func HandleGet(conn net.Conn, cmd *Command) {
	if len(cmd.Args) < 1 {
		response := "-ERR wrong number of arguments for 'get' command\r\n"
		conn.Write([]byte(response))
		return
	}

	key := cmd.Args[0]
	value, exists := Dictionary.Get(key)

	var response string
	if exists {
		response = parser.ToBulkString(value)
		fmt.Println("GET:", key, "=", value)
	} else {
		response = "$-1\r\n" // RESP Null Bulk String
		fmt.Println("GET:", key, "= (not found)")
	}

	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write GET response")
		return
	}
}

func HandleInfo(conn net.Conn, cmd *Command) {
	// Default to all sections if no argument provided
	section := "all"
	if len(cmd.Args) > 0 {
		section = strings.ToLower(cmd.Args[0])
	}

	var infoContent string
	role := GetServerRole()

	switch section {
	case "replication":
		// Only return replication information
		infoContent = "role:" + role
	case "all":
		// For now, just return replication info for "all" as well
		// In a full implementation, this would include all sections
		infoContent = "# Replication\nrole:" + role
	default:
		// Unknown section, return empty (Redis behavior)
		infoContent = ""
	}

	// Encode as bulk string
	response := parser.ToBulkString(infoContent)
	_, err := conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to write INFO response")
		return
	}

	fmt.Printf("INFO: section=%s, role=%s\n", section, role)
}
