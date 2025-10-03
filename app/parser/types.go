package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// RESPValue represents a parsed RESP value
type RESPValue struct {
	Type  string      // "simple", "error", "integer", "bulk", "array"
	Str   string      // for simple strings, errors, and bulk strings
	Int   int         // for integers
	Array []RESPValue // for arrays
}

// ParseRESP parses a RESP message and returns the parsed value
func ParseRESP(data string) (RESPValue, error) {
	if len(data) == 0 {
		return RESPValue{}, fmt.Errorf("empty data")
	}

	switch data[0] {
	case '+': // Simple String
		return parseSimpleString(data)
	case '-': // Error
		return parseError(data)
	case ':': // Integer
		return parseInteger(data)
	case '$': // Bulk String
		return ParseBulkString(data)
	case '*': // Array
		return ParseArray(data)
	default:
		return RESPValue{}, fmt.Errorf("unknown RESP type: %c", data[0])
	}
}

// parseSimpleString parses a simple string from RESP data
func parseSimpleString(data string) (RESPValue, error) {
	end := strings.Index(data, "\r\n")
	if end == -1 {
		return RESPValue{}, fmt.Errorf("invalid simple string")
	}
	return RESPValue{Type: "simple", Str: data[1:end]}, nil
}

// parseError parses an error from RESP data
func parseError(data string) (RESPValue, error) {
	end := strings.Index(data, "\r\n")
	if end == -1 {
		return RESPValue{}, fmt.Errorf("invalid error")
	}
	return RESPValue{Type: "error", Str: data[1:end]}, nil
}

// parseInteger parses an integer from RESP data
func parseInteger(data string) (RESPValue, error) {
	end := strings.Index(data, "\r\n")
	if end == -1 {
		return RESPValue{}, fmt.Errorf("invalid integer")
	}
	val, err := strconv.Atoi(data[1:end])
	if err != nil {
		return RESPValue{}, fmt.Errorf("invalid integer value")
	}
	return RESPValue{Type: "integer", Int: val}, nil
}
