package parser

import (
	"fmt"
	"strconv"
	"strings"
)

func ToBulkString(s string) string {
	length := len(s)
	return "$" + strconv.Itoa(length) + "\r\n" + s + "\r\n"
}

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
		end := strings.Index(data, "\r\n")
		if end == -1 {
			return RESPValue{}, fmt.Errorf("invalid simple string")
		}
		return RESPValue{Type: "simple", Str: data[1:end]}, nil

	case '-': // Error
		end := strings.Index(data, "\r\n")
		if end == -1 {
			return RESPValue{}, fmt.Errorf("invalid error")
		}
		return RESPValue{Type: "error", Str: data[1:end]}, nil

	case ':': // Integer
		end := strings.Index(data, "\r\n")
		if end == -1 {
			return RESPValue{}, fmt.Errorf("invalid integer")
		}
		val, err := strconv.Atoi(data[1:end])
		if err != nil {
			return RESPValue{}, fmt.Errorf("invalid integer value")
		}
		return RESPValue{Type: "integer", Int: val}, nil

	case '$': // Bulk String
		return ParseBulkString(data)

	case '*': // Array
		return ParseArray(data)

	default:
		return RESPValue{}, fmt.Errorf("unknown RESP type: %c", data[0])
	}
}

// ParseBulkString parses a bulk string from RESP data
func ParseBulkString(data string) (RESPValue, error) {
	firstCRLF := strings.Index(data, "\r\n")
	if firstCRLF == -1 {
		return RESPValue{}, fmt.Errorf("invalid bulk string format")
	}

	lengthStr := data[1:firstCRLF]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return RESPValue{}, fmt.Errorf("invalid bulk string length")
	}

	if length == -1 {
		return RESPValue{Type: "bulk", Str: ""}, nil // null bulk string
	}

	start := firstCRLF + 2
	if len(data) < start+length+2 {
		return RESPValue{}, fmt.Errorf("incomplete bulk string")
	}

	content := data[start : start+length]
	return RESPValue{Type: "bulk", Str: content}, nil
}

// ParseArray parses an array from RESP data
func ParseArray(data string) (RESPValue, error) {
	firstCRLF := strings.Index(data, "\r\n")
	if firstCRLF == -1 {
		return RESPValue{}, fmt.Errorf("invalid array format")
	}

	countStr := data[1:firstCRLF]
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return RESPValue{}, fmt.Errorf("invalid array count")
	}

	if count == -1 {
		return RESPValue{Type: "array", Array: nil}, nil // null array
	}

	array := make([]RESPValue, 0, count)
	remaining := data[firstCRLF+2:]

	for i := 0; i < count; i++ {
		if len(remaining) == 0 {
			return RESPValue{}, fmt.Errorf("incomplete array")
		}

		var element RESPValue
		var consumed int

		switch remaining[0] {
		case '$':
			element, err = ParseBulkString(remaining)
			if err != nil {
				return RESPValue{}, err
			}
			// Calculate how much we consumed
			firstCRLF := strings.Index(remaining, "\r\n")
			lengthStr := remaining[1:firstCRLF]
			length, _ := strconv.Atoi(lengthStr)
			consumed = firstCRLF + 2 + length + 2
		default:
			return RESPValue{}, fmt.Errorf("unsupported array element type: %c", remaining[0])
		}

		array = append(array, element)
		remaining = remaining[consumed:]
	}

	return RESPValue{Type: "array", Array: array}, nil
}

// EncodeRESP encodes a RESPValue back to RESP format
func EncodeRESP(value RESPValue) string {
	switch value.Type {
	case "simple":
		return "+" + value.Str + "\r\n"
	case "error":
		return "-" + value.Str + "\r\n"
	case "integer":
		return ":" + strconv.Itoa(value.Int) + "\r\n"
	case "bulk":
		if value.Str == "" {
			return "$-1\r\n" // null bulk string
		}
		return "$" + strconv.Itoa(len(value.Str)) + "\r\n" + value.Str + "\r\n"
	case "array":
		if value.Array == nil {
			return "*-1\r\n" // null array
		}
		result := "*" + strconv.Itoa(len(value.Array)) + "\r\n"
		for _, element := range value.Array {
			result += EncodeRESP(element)
		}
		return result
	default:
		return "-ERR unknown type\r\n"
	}
}
