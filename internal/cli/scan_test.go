package cli_test

import (
	"github.com/Gulliby-development/trailspin/internal/cli"
	"testing"
)

func TestScanCleanTrace(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "ok, 4 calls\n"}},
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": ""}, want: cli.Result{Out: "ok, 0 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanRepeatedCalls(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Code: 1, Out: "4 repeat read_file {\"path\":\"a.txt\"} x4 in last 4\n1 finding in 6 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat", "1"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Code: 1, Out: "2 repeat read_file {\"path\":\"a.txt\"} x2 in last 2\n1 finding in 6 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat=4"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Code: 1, Out: "5 repeat read_file {\"path\":\"a.txt\"} x5 in last 5\n1 finding in 6 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat", "5"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Out: "ok, 6 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "3", "--max-repeat", "2"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Code: 1, Out: "3 repeat read_file {\"path\":\"a.txt\"} x3 in last 3\n4 repeat read_file {\"path\":\"a.txt\"} x3 in last 3\n5 repeat read_file {\"path\":\"a.txt\"} x3 in last 3\n3 findings in 6 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--window", "2", "--max-repeat", "2"}, files: map[string]string{"t.jsonl": traceRepeat}, want: cli.Result{Out: "ok, 6 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanCycles(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": traceCycle}, want: cli.Result{Code: 1, Out: "6 cycle period 2 x3: read_file {\"path\":\"a.txt\"} | edit {\"path\":\"b.txt\"}\n7 repeat read_file {\"path\":\"a.txt\"} x4 in last 7\n8 repeat edit {\"path\":\"b.txt\"} x4 in last 8\n3 findings in 9 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--cycle-reps", "2"}, files: map[string]string{"t.jsonl": traceCycle}, want: cli.Result{Code: 1, Out: "4 cycle period 2 x2: read_file {\"path\":\"a.txt\"} | edit {\"path\":\"b.txt\"}\n7 repeat read_file {\"path\":\"a.txt\"} x4 in last 7\n8 repeat edit {\"path\":\"b.txt\"} x4 in last 8\n3 findings in 9 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--cycle-reps", "4"}, files: map[string]string{"t.jsonl": traceCycle}, want: cli.Result{Code: 1, Out: "7 repeat read_file {\"path\":\"a.txt\"} x4 in last 7\n8 repeat edit {\"path\":\"b.txt\"} x4 in last 8\n8 cycle period 2 x4: read_file {\"path\":\"a.txt\"} | edit {\"path\":\"b.txt\"}\n3 findings in 9 calls\n"}},
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": traceCycle3}, want: cli.Result{Code: 1, Out: "9 cycle period 3 x3: ls {} | read_file {\"path\":\"a.txt\"} | grep {\"q\":\"retry\"}\n10 repeat ls {} x4 in last 10\n2 findings in 10 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--cycle-reps", "4"}, files: map[string]string{"t.jsonl": traceCycle3}, want: cli.Result{Code: 1, Out: "10 repeat ls {} x4 in last 10\n1 finding in 10 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanErrorStreaks(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": traceErrors}, want: cli.Result{Code: 1, Out: "5 repeat run_tests {\"cmd\":\"go test\",\"dir\":\".\"} x4 in last 5\n6 errors 3 failed calls in a row\n2 findings in 7 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-errors", "2"}, files: map[string]string{"t.jsonl": traceErrors}, want: cli.Result{Code: 1, Out: "2 errors 2 failed calls in a row\n5 repeat run_tests {\"cmd\":\"go test\",\"dir\":\".\"} x4 in last 5\n5 errors 2 failed calls in a row\n3 findings in 7 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-errors", "1"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 1, Out: "2 errors 1 failed calls in a row\n4 errors 1 failed calls in a row\n6 cycle period 2 x3: read_file {\"path\":\"a.txt\"} | grep {\"q\":\"retry\"}\n6 errors 1 failed calls in a row\n7 repeat read_file {\"path\":\"a.txt\"} x4 in last 7\n8 errors 1 failed calls in a row\n6 findings in 8 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanBudgets(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl", "--max-tokens", "100000"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Code: 1, Out: "3 tokens 100001 over limit 100000\n1 finding in 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-tokens", "100001"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Code: 1, Out: "4 tokens 100006 over limit 100001\n1 finding in 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-tokens", "39999"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Code: 1, Out: "1 tokens 40000 over limit 39999\n1 finding in 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-calls", "3"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Code: 1, Out: "4 calls over limit 3\n1 finding in 4 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-calls", "4"}, files: map[string]string{"t.jsonl": traceBudget}, want: cli.Result{Out: "ok, 4 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanSeveralKindsInOneStep(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl", "--max-repeat", "1", "--max-errors", "2", "--max-tokens", "500"}, files: map[string]string{"t.jsonl": traceErrors}, want: cli.Result{Code: 1, Out: "1 tokens 900 over limit 500\n2 repeat run_tests {\"cmd\":\"go test\",\"dir\":\".\"} x2 in last 2\n2 errors 2 failed calls in a row\n5 errors 2 failed calls in a row\n4 findings in 7 calls\n"}},
		{args: []string{"scan", "t.jsonl", "--max-repeat", "1", "--window", "2", "--max-calls", "5", "--cycle-reps", "2"}, files: map[string]string{"t.jsonl": traceMixed}, want: cli.Result{Code: 1, Out: "4 cycle period 2 x2: read_file {\"path\":\"a.txt\"} | grep {\"q\":\"retry\"}\n6 calls over limit 5\n2 findings in 8 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanStandardInput(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "-"}, stdin: traceRepeat, want: cli.Result{Code: 1, Out: "4 repeat read_file {\"path\":\"a.txt\"} x4 in last 4\n1 finding in 6 calls\n"}},
		{args: []string{"scan", "--max-repeat", "1", "-"}, stdin: traceClean, want: cli.Result{Out: "ok, 4 calls\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
func TestScanSingularOutput(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t"}, files: map[string]string{"t": "{\"tool\":\"x\"}\n"}, want: cli.Result{Out: "ok, 1 call\n"}},
		{args: []string{"scan", "-", "--max-errors", "1"}, stdin: "{\"tool\":\"x\",\"ok\":false}\n", want: cli.Result{Code: 1, Out: "1 errors 1 failed calls in a row\n1 finding in 1 call\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run() = %#v, want %#v", got, tc.want)
		}
	}
}
