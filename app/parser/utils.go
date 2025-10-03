package parser

import "strconv"

// EncodeRESP encodes a RESPValue back to RESP format
func EncodeRESP(value RESPValue) string {
	switch value.Type {
	case "simple":
		return ToSimpleString(value.Str)
	case "error":
		return ToError(value.Str)
	case "integer":
		return ToInteger(value.Int)
	case "bulk":
		if value.Str == "" {
			return ToNullBulkString()
		}
		return ToBulkString(value.Str)
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
		return ToError("ERR unknown type")
	}
}
