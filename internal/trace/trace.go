// Package trace parses strict JSON-lines tool-call traces.
package trace

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Gulliby-development/trailspin/internal/jsonv"
)

// Call is one validated tool invocation in a trace.
type Call struct {
	Step   int
	Tool   string
	Args   string
	OK     bool
	Tokens int64
}

// Key returns the canonical identity used to compare calls.
func (call Call) Key() string {
	return call.Tool + " " + call.Args
}

// Parse validates and parses a complete trace. Any failure returns no calls.
func Parse(data []byte, path string) ([]Call, error) {
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("%s: input is not valid UTF-8", path)
	}
	text := string(data)
	if strings.HasPrefix(text, "\ufeff") {
		text = text[len("\ufeff"):]
	}

	var calls []Call
	lineNumber := 0
	for len(text) > 0 {
		lineNumber++
		line := text
		if newline := strings.IndexByte(text, '\n'); newline >= 0 {
			line = text[:newline]
			text = text[newline+1:]
		} else {
			text = ""
		}
		line = strings.TrimSuffix(line, "\r")
		if line == "" || strings.Trim(line, " \t") == "" || strings.HasPrefix(line, "#") {
			continue
		}
		value, err := jsonv.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %s", path, lineNumber, err)
		}
		object, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s:%d: not a JSON object", path, lineNumber)
		}
		call, err := parseCall(object)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %s", path, lineNumber, err)
		}
		call.Step = len(calls) + 1
		calls = append(calls, call)
	}
	return calls, nil
}

func parseCall(object map[string]any) (Call, error) {
	toolValue, exists := object["tool"]
	if !exists {
		return Call{}, fmt.Errorf("missing \"tool\"")
	}
	tool, ok := toolValue.(string)
	if !ok || !validTool(tool) {
		return Call{}, fmt.Errorf("\"tool\" must be 1 to 64 characters of A-Z a-z 0-9 _ . : -")
	}

	args := map[string]any{}
	if value, exists := object["args"]; exists {
		var ok bool
		args, ok = value.(map[string]any)
		if !ok {
			return Call{}, fmt.Errorf("\"args\" must be an object")
		}
	}

	okValue := true
	if value, exists := object["ok"]; exists {
		var ok bool
		okValue, ok = value.(bool)
		if !ok {
			return Call{}, fmt.Errorf("\"ok\" must be true or false")
		}
	}

	tokens := int64(0)
	if value, exists := object["tokens"]; exists {
		number, ok := value.(json.Number)
		if !ok {
			return Call{}, fmt.Errorf("\"tokens\" must be an integer from 0 to 1000000000")
		}
		parsed, err := strconv.ParseInt(string(number), 10, 64)
		if err != nil || parsed < 0 || parsed > 1000000000 {
			return Call{}, fmt.Errorf("\"tokens\" must be an integer from 0 to 1000000000")
		}
		tokens = parsed
	}
	return Call{Tool: tool, Args: jsonv.Canonical(args), OK: okValue, Tokens: tokens}, nil
}

func validTool(tool string) bool {
	if len(tool) < 1 || len(tool) > 64 {
		return false
	}
	for index := 0; index < len(tool); index++ {
		character := tool[index]
		if (character >= 'A' && character <= 'Z') ||
			(character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("_.:-", rune(character)) {
			continue
		}
		return false
	}
	return true
}
