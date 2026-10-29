package cli_test

import (
	"github.com/Gulliby-development/trailspin/internal/cli"
	"testing"
)

func TestOptionValueRanges(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl", "--window", "1"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"1\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "2"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "1000"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "1001"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"1001\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "0"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"0\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "01"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"01\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "x"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"x\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "+5"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"+5\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "5.0"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"5.0\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "-3"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"-3\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "12345678901"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --window value \"12345678901\": expected an integer from 2 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat", "0"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-repeat value \"0\": expected an integer from 1 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat", "1000"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat", "1001"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-repeat value \"1001\": expected an integer from 1 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--cycle-reps", "1"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --cycle-reps value \"1\": expected an integer from 2 to 100\n"}},
		{args: []string{"scan", "t.jsonl", "--cycle-reps", "100"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--cycle-reps", "101"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --cycle-reps value \"101\": expected an integer from 2 to 100\n"}},
		{args: []string{"scan", "t.jsonl", "--max-tokens", "0"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-tokens value \"0\": expected an integer from 1 to 1000000000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-tokens", "1000000000"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-tokens", "1000000001"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-tokens value \"1000000001\": expected an integer from 1 to 1000000000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-calls", "0"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-calls value \"0\": expected an integer from 1 to 100000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-calls", "100000"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-calls", "100001"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-calls value \"100001\": expected an integer from 1 to 100000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-errors", "0"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-errors value \"0\": expected an integer from 1 to 1000\n"}},
		{args: []string{"scan", "t.jsonl", "--max-errors", "1001"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-errors value \"1001\": expected an integer from 1 to 1000\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestOptionMissingValues(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl", "--window"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: option --window needs a value\n"}},
		{args: []string{"scan", "--max-calls=", "t.jsonl"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: bad --max-calls value \"\": expected an integer from 1 to 100000\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestOptionUnknownOptions(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl", "--bogus"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"--bogus\"\n"}},
		{args: []string{"keys", "t.jsonl", "--window", "3"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"--window\"\n"}},
		{args: []string{"summary", "-x", "t.jsonl"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Code: 2, Err: "trailspin: unknown option \"-x\"\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestOptionLastValueWins(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl", "--max-repeat", "9", "--max-repeat", "1"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Code: 1, Out: "2 repeat read_file {\"path\":\"a.txt\"} x2 in last 2\n1 finding in 6 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestOptionOptionsMayStandAnywhere(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "--max-repeat=1", "t.jsonl"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Code: 1, Out: "2 repeat read_file {\"path\":\"a.txt\"} x2 in last 2\n1 finding in 6 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
