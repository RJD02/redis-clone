package parser

import "strconv"

// ToBulkString formats a string as a Redis RESP bulk string
// Format: $<length>\r\n<data>\r\n
func ToBulkString(s string) string {
	length := len(s)
	return "$" + strconv.Itoa(length) + "\r\n" + s + "\r\n"
}

// ToSimpleString formats a string as a Redis RESP simple string
// Format: +<data>\r\n
func ToSimpleString(s string) string {
	return "+" + s + "\r\n"
}

// ToError formats a string as a Redis RESP error
// Format: -<error>\r\n
func ToError(s string) string {
	return "-" + s + "\r\n"
}

// ToInteger formats an integer as a Redis RESP integer
// Format: :<number>\r\n
func ToInteger(i int) string {
	return ":" + strconv.Itoa(i) + "\r\n"
}

// ToNullBulkString returns a RESP null bulk string
func ToNullBulkString() string {
	return "$-1\r\n"
}
