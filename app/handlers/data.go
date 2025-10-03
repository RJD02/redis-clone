package handlers

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/parser"
	"github.com/codecrafters-io/redis-starter-go/app/storage"
)

// HandleSet handles the SET command with optional expiration
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
	storage.Dictionary.Set(key, value, expiration)

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

// HandleGet handles the GET command
func HandleGet(conn net.Conn, cmd *Command) {
	if len(cmd.Args) < 1 {
		response := "-ERR wrong number of arguments for 'get' command\r\n"
		conn.Write([]byte(response))
		return
	}

	key := cmd.Args[0]
	value, exists := storage.Dictionary.Get(key)

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
