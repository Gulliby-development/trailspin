package cli_test

import (
	"testing"

	"github.com/Gulliby-development/trailspin/internal/cli"
)

func TestSummaryTop(t *testing.T) {
	ties := "{\"tool\":\"b\",\"args\":{}}\n{\"tool\":\"a\",\"args\":{}}\n{\"tool\":\"c\",\"args\":{}}\n{\"tool\":\"c\",\"args\":{}}\n{\"tool\":\"a\",\"args\":{}}\n{\"tool\":\"d\",\"args\":{}}\n"

	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"summary", "t.jsonl", "--top", "1"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Out: "read_file 4 0 10\n(others) 4 4 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "2"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Out: "read_file 4 0 10\ngrep 3 3 0\n(others) 1 1 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "3"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Out: "read_file 4 0 10\ngrep 3 3 0\nedit 1 1 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "1000"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Out: "read_file 4 0 10\ngrep 3 3 0\nedit 1 1 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl", "--top=1"}, files: map[string]string{"t.jsonl": traceErrors}, want: cli.Result{Out: "run_tests 6 6 5400\n(others) 1 0 20\ntotal 7 6 5420\n"}},
		{args: []string{"summary", "--top", "2", "t.jsonl"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Out: "edit 1 0 30001\ngrep 1 0 30000\n(others) 2 0 40005\ntotal 4 0 100006\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "3"}, files: map[string]string{"t.jsonl": ties}, want: cli.Result{Out: "a 2 0 0\nc 2 0 0\nb 1 0 0\n(others) 1 0 0\ntotal 6 0 0\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "1", "--top", "3"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Out: "read_file 4 0 10\ngrep 3 3 0\nedit 1 1 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "1"}, files: map[string]string{"t.jsonl": "# nothing\n"}, want: cli.Result{Out: "total 0 0 0\n"}},
		{args: []string{"summary", "-", "--top", "1"}, stdin: traceMixed, want: cli.Result{Out: "read_file 4 0 10\n(others) 4 4 0\ntotal 8 4 10\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "0"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: bad --top value \"0\": expected an integer from 1 to 1000\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "1001"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: bad --top value \"1001\": expected an integer from 1 to 1000\n"}},
		{args: []string{"summary", "t.jsonl", "--top", "02"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: bad --top value \"02\": expected an integer from 1 to 1000\n"}},
		{args: []string{"summary", "t.jsonl", "--top"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: option --top needs a value\n"}},
		{args: []string{"summary", "t.jsonl", "--window", "5"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"--window\"\n"}},
		{args: []string{"keys", "t.jsonl", "--top", "1"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"--top\"\n"}},
		{args: []string{"scan", "t.jsonl", "--top", "1"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"--top\"\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}
