package cli_test

import (
	"testing"

	"github.com/Gulliby-development/trailspin/internal/cli"
)

func TestSummaryPerTool(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"summary", "t.jsonl"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Out: "read_file 4 0 10\ngrep 3 3 0\nedit 1 1 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl"}, files: map[string]string{"t.jsonl": traceErrors}, want: cli.Result{Out: "run_tests 6 6 5400\nedit 1 0 20\ntotal 7 6 5420\n"}},
		{args: []string{"summary", "t.jsonl"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Out: "edit 1 0 30001\ngrep 1 0 30000\nread_file 1 0 40000\nrun_tests 1 0 5\ntotal 4 0 100006\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestSummaryEmptyTrace(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"summary", "t.jsonl"}, files: map[string]string{"t.jsonl": "# nothing\n"}, want: cli.Result{Out: "total 0 0 0\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestSummaryOrderingAndTies(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"summary", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"b\",\"args\":{}}\n{\"tool\":\"a\",\"args\":{}}\n{\"tool\":\"c\",\"args\":{}}\n{\"tool\":\"c\",\"args\":{}}\n{\"tool\":\"a\",\"args\":{}}\n{\"tool\":\"d\",\"args\":{}}\n"}, want: cli.Result{Out: "a 2 0 0\nc 2 0 0\nb 1 0 0\nd 1 0 0\ntotal 6 0 0\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}
