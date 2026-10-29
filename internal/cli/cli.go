// Package cli implements trailspin's command-line interface.
package cli

import (
	"io"
	"os"
	"strings"

	"github.com/Gulliby-development/trailspin/internal/report"
	"github.com/Gulliby-development/trailspin/internal/scan"
	"github.com/Gulliby-development/trailspin/internal/trace"
)

// Usage is the command synopsis, including its final newline.
const Usage = "usage: trailspin COMMAND [ARGS]\n  scan TRACE [--window N] [--max-repeat N] [--max-errors N] [--cycle-reps N]\n             [--max-tokens N] [--max-calls N]\n  keys TRACE\n  summary TRACE\n"

// Env supplies all input sources used by Run.
type Env struct {
	ReadFile func(path string) ([]byte, error)
	Stdin    io.Reader
}

// Result is the complete, comparable result of a CLI invocation.
type Result struct {
	Code     int
	Out, Err string
}

// Run dispatches one command without accessing process-global IO or the filesystem.
func Run(args []string, env Env) Result {
	if len(args) == 0 {
		return Result{Code: 2, Err: Usage}
	}
	command := args[0]
	if command == "help" || command == "-h" || command == "--help" {
		return Result{Out: Usage}
	}
	if command != "scan" && command != "keys" && command != "summary" {
		return failure("unknown command \"" + command + "\"")
	}

	options, positional, diagnostic := parseOptions(command, args[1:])
	if diagnostic != "" {
		return failure(diagnostic)
	}
	if len(positional) != 1 {
		return failure(command + ": wrong number of arguments")
	}

	path := positional[0]
	displayPath := path
	var data []byte
	var err error
	if path == "-" {
		displayPath = "<stdin>"
		if env.Stdin == nil {
			return failure(displayPath + ": cannot read file")
		}
		data, err = io.ReadAll(env.Stdin)
	} else {
		if env.ReadFile == nil {
			return failure(path + ": cannot read file")
		}
		data, err = env.ReadFile(path)
	}
	if err != nil {
		return failure(displayPath + ": cannot read file")
	}
	calls, err := trace.Parse(data, displayPath)
	if err != nil {
		return failure(err.Error())
	}

	switch command {
	case "keys":
		return Result{Out: report.Keys(calls)}
	case "summary":
		return Result{Out: report.Summary(calls)}
	default:
		findings := scan.Analyze(calls, options)
		if len(findings) == 0 {
			word := "calls"
			if len(calls) == 1 {
				word = "call"
			}
			return Result{Out: "ok, " + itoa(len(calls)) + " " + word + "\n"}
		}
		var output strings.Builder
		for _, finding := range findings {
			output.WriteString(finding)
			output.WriteByte('\n')
		}
		findingWord, callWord := "findings", "calls"
		if len(findings) == 1 {
			findingWord = "finding"
		}
		if len(calls) == 1 {
			callWord = "call"
		}
		output.WriteString(itoa(len(findings)) + " " + findingWord + " in " + itoa(len(calls)) + " " + callWord + "\n")
		return Result{Code: 1, Out: output.String()}
	}
}

func failure(message string) Result {
	return Result{Code: 2, Err: "trailspin: " + message + "\n"}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var buffer [20]byte
	index := len(buffer)
	for value > 0 {
		index--
		buffer[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buffer[index:])
}

// Main binds file input to the operating system and writes exact result bytes.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	result := Run(args, Env{ReadFile: os.ReadFile, Stdin: stdin})
	writer := stdout
	text := result.Out
	if text == "" && result.Err != "" {
		writer = stderr
		text = result.Err
	}
	if text != "" {
		written, err := io.WriteString(writer, text)
		if err != nil || written != len(text) {
			return 2
		}
	}
	return result.Code
}
