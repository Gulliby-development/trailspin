package cli_test

import (
	"testing"

	"github.com/Gulliby-development/trailspin/internal/cli"
	"github.com/Gulliby-development/trailspin/internal/report"
	"github.com/Gulliby-development/trailspin/internal/trace"
)

func TestKeysCanonicalArguments(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": traceArgs}, want: cli.Result{Out: "1 read_file ok 7 {\"lines\":[1,2],\"opts\":{\"a\":true,\"z\":null},\"path\":\"a.txt\"}\n2 read_file ok 9 {\"lines\":[1,2],\"opts\":{\"a\":true,\"z\":null},\"path\":\"a.txt\"}\n3 web.search fail 0 {\"big\":12345678901234567890,\"n\":0,\"q\":\"caf\u00e9 <b> & \\\"x\\\"\\n\"}\n4 a:b-c_d ok 0 {\"c\":\"\\u001f\\t\u007f\",\"e\":\"\U0001f600\"}\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": traceClean}, want: cli.Result{Out: "1 read_file ok 120 {\"path\":\"a.txt\"}\n2 grep ok 80 {\"q\":\"retry\"}\n3 edit ok 300 {\"path\":\"b.txt\"}\n4 run_tests ok 40 {\"cmd\":\"go test\",\"dir\":\".\"}\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestKeysDefaults(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"x\"}\n{\"tool\":\"y\",\"args\":{},\"ok\":true,\"tokens\":0}\n{\"tool\":\"z\",\"ok\":false}\n"}, want: cli.Result{Out: "1 x ok 0 {}\n2 y ok 0 {}\n3 z fail 0 {}\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestKeysTextRules(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "\ufeff# header\n\n  \t\n{\"tool\":\"x\"}\r\n# note\n \t{ \"tool\" : \"y\" , \"args\" : { \"b\" : [ ] , \"a\" : { } } } \t\n"}, want: cli.Result{Out: "1 x ok 0 {}\n2 y ok 0 {\"a\":{},\"b\":[]}\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"x\"}"}, want: cli.Result{Out: "1 x ok 0 {}\n"}},
		{args: []string{"keys", "-"}, stdin: "{\"tool\":\"x\"}\n", want: cli.Result{Out: "1 x ok 0 {}\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestKeysUnknownKeysAreIgnored(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"x\",\"note\":{\"a\":1.5,\"a\":2},\"step\":7}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:1: number must be an integer\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestKeysEmptyReport(t *testing.T) {
	if got := run([]string{"keys", "t.jsonl"}, map[string]string{"t.jsonl": ""}, ""); got != (cli.Result{}) {
		t.Errorf("empty keys = %#v, want empty result", got)
	}
	if got := report.Keys([]trace.Call{}); got != "" {
		t.Errorf("report.Keys([]trace.Call{}) = %q, want empty string", got)
	}
}

func TestKeysStandardInputDefault(t *testing.T) {
	want := cli.Result{Code: 0, Out: "1 x ok 0 {}\n"}
	got := run([]string{"keys", "-"}, nil, "{\"tool\":\"x\"}\n")
	if got != want {
		t.Errorf("stdin keys = %#v, want %#v", got, want)
	}
}
