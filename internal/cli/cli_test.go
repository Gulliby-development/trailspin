package cli_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/Gulliby-development/trailspin/internal/cli"
)

func TestCliHelpAndUsage(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"help"}, want: cli.Result{Code: 0, Out: "usage: trailspin COMMAND [ARGS]\n  scan TRACE [--window N] [--max-repeat N] [--max-errors N] [--cycle-reps N]\n             [--max-tokens N] [--max-calls N]\n  keys TRACE\n  summary TRACE [--top N]\n"}},
		{args: []string{"-h"}, want: cli.Result{Code: 0, Out: "usage: trailspin COMMAND [ARGS]\n  scan TRACE [--window N] [--max-repeat N] [--max-errors N] [--cycle-reps N]\n             [--max-tokens N] [--max-calls N]\n  keys TRACE\n  summary TRACE [--top N]\n"}},
		{args: []string{"--help", "x"}, want: cli.Result{Code: 0, Out: "usage: trailspin COMMAND [ARGS]\n  scan TRACE [--window N] [--max-repeat N] [--max-errors N] [--cycle-reps N]\n             [--max-tokens N] [--max-calls N]\n  keys TRACE\n  summary TRACE [--top N]\n"}},
		{args: []string{}, want: cli.Result{Code: 2, Err: "usage: trailspin COMMAND [ARGS]\n  scan TRACE [--window N] [--max-repeat N] [--max-errors N] [--cycle-reps N]\n             [--max-tokens N] [--max-calls N]\n  keys TRACE\n  summary TRACE [--top N]\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestCliUnknownCommand(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"frob"}, want: cli.Result{Code: 2, Err: "trailspin: unknown command \"frob\"\n"}},
		{args: []string{"frob", "--x"}, want: cli.Result{Code: 2, Err: "trailspin: unknown command \"frob\"\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestCliArgumentCounts(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan"}, want: cli.Result{Code: 2, Err: "trailspin: scan: wrong number of arguments\n"}},
		{args: []string{"keys"}, want: cli.Result{Code: 2, Err: "trailspin: keys: wrong number of arguments\n"}},
		{args: []string{"summary"}, want: cli.Result{Code: 2, Err: "trailspin: summary: wrong number of arguments\n"}},
		{args: []string{"scan", "t.jsonl", "u.jsonl"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: scan: wrong number of arguments\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestCliUnreadablePath(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"summary", "missing.jsonl"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: missing.jsonl: cannot read file\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

type errorReader struct {
	data []byte
	err  error
}

func (r errorReader) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	return n, r.err
}

type writeResult struct {
	n   int
	err error
}

type resultWriter struct {
	results []writeResult
	calls   int
}

func (w *resultWriter) Write(p []byte) (int, error) {
	if w.calls >= len(w.results) {
		return len(p), nil
	}
	result := w.results[w.calls]
	w.calls++
	if result.n > len(p) {
		result.n = len(p)
	}
	return result.n, result.err
}

func TestCliInputDependenciesAndReadFailures(t *testing.T) {
	tests := []struct {
		name string
		args []string
		env  cli.Env
		want cli.Result
	}{
		{name: "nil file reader", args: []string{"keys", "missing"}, want: cli.Result{Code: 2, Err: "trailspin: missing: cannot read file\n"}},
		{name: "nil stdin", args: []string{"keys", "-"}, want: cli.Result{Code: 2, Err: "trailspin: <stdin>: cannot read file\n"}},
		{name: "stdin returns bytes and error", args: []string{"keys", "-"}, env: cli.Env{Stdin: errorReader{data: []byte(`{"tool":"x"}`), err: errors.New("read failed")}}, want: cli.Result{Code: 2, Err: "trailspin: <stdin>: cannot read file\n"}},
		{name: "file reader error", args: []string{"keys", "broken"}, env: cli.Env{ReadFile: func(string) ([]byte, error) { return []byte(`{"tool":"x"}`), errors.New("read failed") }}, want: cli.Result{Code: 2, Err: "trailspin: broken: cannot read file\n"}},
	}
	for _, tc := range tests {
		if got := cli.Run(tc.args, tc.env); got != tc.want {
			t.Errorf("%s: Run(%v) = %#v, want %#v", tc.name, tc.args, got, tc.want)
		}
	}
}

func TestCliValidationPrecedenceDoesNotReadInput(t *testing.T) {
	reads := 0
	env := cli.Env{
		ReadFile: func(string) ([]byte, error) {
			reads++
			return []byte(`{"tool":"x"}`), nil
		},
		Stdin: errorReader{err: errors.New("must not read")},
	}
	tests := []struct {
		args []string
		want cli.Result
	}{
		{args: []string{"unknown", "--bad"}, want: cli.Result{Code: 2, Err: "trailspin: unknown command \"unknown\"\n"}},
		{args: []string{"scan", "--bad"}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"--bad\"\n"}},
		{args: []string{"scan", "one", "two"}, want: cli.Result{Code: 2, Err: "trailspin: scan: wrong number of arguments\n"}},
	}
	for _, tc := range tests {
		if got := cli.Run(tc.args, env); got != tc.want {
			t.Errorf("Run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
	if reads != 0 {
		t.Errorf("input callback called %d times, want 0", reads)
	}
}

func TestCliRawDiagnosticText(t *testing.T) {
	got := cli.Run([]string{"a\"b\\c"}, cli.Env{})
	want := cli.Result{Code: 2, Err: "trailspin: unknown command \"a\"b\\c\"\n"}
	if got != want {
		t.Errorf("Run with raw command text = %#v, want %#v", got, want)
	}
}

func TestCliMainWritesExactBytesAndReturnsStatus(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stdin  string
		code   int
		stdout string
		stderr string
	}{
		{name: "success", args: []string{"keys", "-"}, stdin: `{"tool":"x"}` + "\n", code: 0, stdout: "1 x ok 0 {}\n"},
		{name: "finding", args: []string{"scan", "-"}, stdin: "{\"tool\":\"a\"}\n{\"tool\":\"a\"}\n{\"tool\":\"a\"}\n{\"tool\":\"a\"}\n", code: 1, stdout: "4 repeat a {} x4 in last 4\n1 finding in 4 calls\n"},
		{name: "usage error", args: []string{"frob"}, code: 2, stderr: "trailspin: unknown command \"frob\"\n"},
		{name: "no arguments", args: []string{}, code: 2, stderr: cli.Usage},
	}
	for _, tc := range tests {
		var stdout, stderr bytes.Buffer
		if got := cli.Main(tc.args, bytes.NewBufferString(tc.stdin), &stdout, &stderr); got != tc.code || stdout.String() != tc.stdout || stderr.String() != tc.stderr {
			t.Errorf("%s: Main = (%d, %q, %q), want (%d, %q, %q)", tc.name, got, stdout.String(), stderr.String(), tc.code, tc.stdout, tc.stderr)
		}
	}
}

func TestCliMainWriterFailures(t *testing.T) {
	writeErr := errors.New("write failed")
	tests := []struct {
		name       string
		args       []string
		stdout     *resultWriter
		stderr     *resultWriter
		wantStatus int
	}{
		{name: "stdout error", args: []string{"help"}, stdout: &resultWriter{results: []writeResult{{err: writeErr}}}, stderr: &resultWriter{}, wantStatus: 2},
		{name: "stderr error", args: []string{"missing"}, stdout: &resultWriter{}, stderr: &resultWriter{results: []writeResult{{err: writeErr}}}, wantStatus: 2},
		{name: "stdout short write", args: []string{"help"}, stdout: &resultWriter{results: []writeResult{{n: 1, err: io.ErrShortWrite}}}, stderr: &resultWriter{}, wantStatus: 2},
		{name: "stderr short write", args: []string{"frob"}, stdout: &resultWriter{}, stderr: &resultWriter{results: []writeResult{{n: 1, err: io.ErrShortWrite}}}, wantStatus: 2},
		{name: "stdout short write without error", args: []string{"help"}, stdout: &resultWriter{results: []writeResult{{n: 1}}}, stderr: &resultWriter{}, wantStatus: 2},
		{name: "stderr short write without error", args: []string{"frob"}, stdout: &resultWriter{}, stderr: &resultWriter{results: []writeResult{{n: 1}}}, wantStatus: 2},
	}
	for _, tc := range tests {
		if got := cli.Main(tc.args, bytes.NewReader(nil), tc.stdout, tc.stderr); got != tc.wantStatus {
			t.Errorf("%s: Main status = %d, want %d", tc.name, got, tc.wantStatus)
		}
	}
}

var _ io.Writer = (*resultWriter)(nil)
