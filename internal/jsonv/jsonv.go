package jsonv

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const (
	invalidJSON    = "invalid JSON"
	invalidInteger = "number must be an integer"
	invalidString  = "invalid string"
	tooDeep        = "nesting deeper than 20"
)

var errInvalidJSON = errors.New(invalidJSON)

// Parse decodes JSON while preserving integer lexemes and rejecting ambiguous
// or lossy constructs that encoding/json otherwise accepts.
func Parse(text string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()

	value, err := readValue(decoder, 0)
	if err != nil {
		return nil, err
	}
	_, err = decoder.Token()
	if !errors.Is(err, io.EOF) {
		return nil, errInvalidJSON
	}
	return value, nil
}

func readValue(decoder *json.Decoder, depth int) (any, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, errInvalidJSON
	}

	switch value := token.(type) {
	case json.Delim:
		switch value {
		case '{':
			if depth >= 20 {
				return nil, errors.New(tooDeep)
			}
			return readObject(decoder, depth+1)
		case '[':
			if depth >= 20 {
				return nil, errors.New(tooDeep)
			}
			return readArray(decoder, depth+1)
		default:
			return nil, errInvalidJSON
		}
	case string:
		if strings.ContainsRune(value, '\uFFFD') {
			return nil, errors.New(invalidString)
		}
		return value, nil
	case json.Number:
		if !isInteger(string(value)) {
			return nil, errors.New(invalidInteger)
		}
		if value == "-0" {
			value = "0"
		}
		return value, nil
	case bool, nil:
		return value, nil
	default:
		return nil, errInvalidJSON
	}
}

func readObject(decoder *json.Decoder, depth int) (any, error) {
	object := make(map[string]any)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, errInvalidJSON
		}
		key, ok := token.(string)
		if !ok {
			return nil, errInvalidJSON
		}
		if strings.ContainsRune(key, '\uFFFD') {
			return nil, errors.New(invalidString)
		}
		if _, exists := object[key]; exists {
			return nil, errors.New("duplicate key " + quoteString(key))
		}
		value, err := readValue(decoder, depth)
		if err != nil {
			return nil, err
		}
		object[key] = value
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') {
		return nil, errInvalidJSON
	}
	return object, nil
}

func readArray(decoder *json.Decoder, depth int) (any, error) {
	array := make([]any, 0)
	for decoder.More() {
		value, err := readValue(decoder, depth)
		if err != nil {
			return nil, err
		}
		array = append(array, value)
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim(']') {
		return nil, errInvalidJSON
	}
	return array, nil
}

func isInteger(value string) bool {
	if value == "-0" || value == "0" {
		return true
	}
	if value == "" {
		return false
	}
	index := 0
	if value[0] == '-' {
		index++
	}
	if index >= len(value) || value[index] < '1' || value[index] > '9' {
		return false
	}
	for index++; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

// quoteString exposes the canonical JSON string spelling internally so parser
// diagnostics and canonical output cannot disagree.
func quoteString(value string) string {
	var buffer bytes.Buffer
	writeString(&buffer, value)
	return buffer.String()
}
