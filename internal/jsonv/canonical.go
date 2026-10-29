package jsonv

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

// Canonical serializes a value returned by Parse using deterministic JSON text.
func Canonical(value any) string {
	var buffer bytes.Buffer
	writeValue(&buffer, value)
	return buffer.String()
}

func writeValue(buffer *bytes.Buffer, value any) {
	switch value := value.(type) {
	case nil:
		buffer.WriteString("null")
	case bool:
		if value {
			buffer.WriteString("true")
		} else {
			buffer.WriteString("false")
		}
	case string:
		writeString(buffer, value)
	case json.Number:
		if value == "-0" {
			buffer.WriteByte('0')
		} else {
			buffer.WriteString(string(value))
		}
	case []any:
		buffer.WriteByte('[')
		for index, element := range value {
			if index != 0 {
				buffer.WriteByte(',')
			}
			writeValue(buffer, element)
		}
		buffer.WriteByte(']')
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Slice(keys, func(i, j int) bool {
			return strings.Compare(keys[i], keys[j]) < 0
		})
		buffer.WriteByte('{')
		for index, key := range keys {
			if index != 0 {
				buffer.WriteByte(',')
			}
			writeString(buffer, key)
			buffer.WriteByte(':')
			writeValue(buffer, value[key])
		}
		buffer.WriteByte('}')
	}
}

func writeString(buffer *bytes.Buffer, value string) {
	buffer.WriteByte('"')
	for _, character := range value {
		switch character {
		case '"':
			buffer.WriteString("\\\"")
		case '\\':
			buffer.WriteString("\\\\")
		case '\n':
			buffer.WriteString("\\n")
		case '\r':
			buffer.WriteString("\\r")
		case '\t':
			buffer.WriteString("\\t")
		default:
			if character < 0x20 {
				const hexadecimal = "0123456789abcdef"
				buffer.WriteString("\\u00")
				buffer.WriteByte(hexadecimal[byte(character)>>4])
				buffer.WriteByte(hexadecimal[byte(character)&0x0f])
			} else {
				buffer.WriteRune(character)
			}
		}
	}
	buffer.WriteByte('"')
}
