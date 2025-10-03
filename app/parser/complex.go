package parser

import (
	"fmt"
	"strconv"
	"strings"
)

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
