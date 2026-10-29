package cli_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Gulliby-development/trailspin/internal/cli"
)

func TestInputBadFiles(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"scan", "t.jsonl"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl: cannot read file\n"}},
		{args: []string{"scan", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"x\"}\n\xff\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl: input is not valid UTF-8\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestInputInvalidJson(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\"\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\"} x\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{'tool':'x'}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\" \"x\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\" \"ok\":true}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"ok\":tru}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n\u00a0{}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{}{}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":01}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":\"\\q\"}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":NaN}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid JSON\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestInputNotAnObject(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n[1]\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: not a JSON object\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n\"s\"\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: not a JSON object\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n5\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: not a JSON object\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\nnull\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: not a JSON object\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n[]\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: not a JSON object\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestInputFieldProblems(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: missing \"tool\"\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":1}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tool\" must be 1 to 64 characters of A-Z a-z 0-9 _ . : -\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tool\" must be 1 to 64 characters of A-Z a-z 0-9 _ . : -\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"a b\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tool\" must be 1 to 64 characters of A-Z a-z 0-9 _ . : -\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tool\" must be 1 to 64 characters of A-Z a-z 0-9 _ . : -\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":[]}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"args\" must be an object\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":null}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"args\" must be an object\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"ok\":null}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"ok\" must be true or false\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"ok\":1}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"ok\" must be true or false\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"ok\":\"true\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"ok\" must be true or false\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"tokens\":-1}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tokens\" must be an integer from 0 to 1000000000\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"tokens\":1000000001}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tokens\" must be an integer from 0 to 1000000000\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"tokens\":\"5\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: \"tokens\" must be an integer from 0 to 1000000000\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"tokens\":1.0}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: number must be an integer\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestInputWalkProblems(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"tool\":\"y\"}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: duplicate key \"tool\"\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":1,\"a\":2}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: duplicate key \"a\"\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":{\"b\":1,\"b\":1}}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: duplicate key \"b\"\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":1.5}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: number must be an integer\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":1e3}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: number must be an integer\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":\"\\ud800\"}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid string\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":\"\\ufffd\"}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid string\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":\"\ufffd\"}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid string\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"\\ud800\":1}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: invalid string\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"tokens\":1.5,\"args\":{\"a\":1,\"a\":2}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: number must be an integer\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestInputDepth(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":[[[[[[[[[[[[[[[[[[1]]]]]]]]]]]]]]]]]]}}\n"}, want: cli.Result{Out: "1 ok ok 0 {}\n2 x ok 0 {\"a\":[[[[[[[[[[[[[[[[[[1]]]]]]]]]]]]]]]]]]}\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":[[[[[[[[[[[[[[[[[[[1]]]]]]]]]]]]]]]]]]]}}\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: nesting deeper than 20\n"}},
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "{\"tool\":\"ok\",\"args\":{}}\n{\"tool\":\"x\",\"args\":{\"a\":[[[[[[[[[[[[[[[[[[[[[[[[[\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:2: nesting deeper than 20\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestInputLineNumbersCountEveryLine(t *testing.T) {
	tests := []struct {
		args  []string
		files map[string]string
		stdin string
		want  cli.Result
	}{
		{args: []string{"keys", "t.jsonl"}, files: map[string]string{"t.jsonl": "# c\n\n{\"tool\":\"ok\",\"args\":{}}\n   \n{\n"}, want: cli.Result{Code: 2, Err: "trailspin: t.jsonl:5: invalid JSON\n"}},
	}
	for _, tc := range tests {
		if got := run(tc.args, tc.files, tc.stdin); got != tc.want {
			t.Errorf("run(%v) = %#v, want %#v", tc.args, got, tc.want)
		}
	}
}

func TestMain(m *testing.M) {
	if os.Getenv("TRAILSPIN_INPUT_AS_MAIN") == "1" {
		os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

func TestInputRealFileBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	if err := os.WriteFile(path, []byte{'{', '"', 't', 'o', 'o', 'l', '"', ':', '"', 'x', '"', '}', '\n', 0xff}, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "keys", path)
	cmd.Env = append(os.Environ(), "TRAILSPIN_INPUT_AS_MAIN=1")
	var stdout, stderr bytesBuffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("program timed out: %v", ctx.Err())
	}
	exitCode := 0
	if exit, ok := err.(*exec.ExitError); ok {
		exitCode = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	wantErr := "trailspin: " + path + ": input is not valid UTF-8\n"
	if exitCode != 2 || stdout.String() != "" || stderr.String() != wantErr {
		t.Errorf("real invalid file = (%d, %q, %q), want (2, \"\", %q)", exitCode, stdout.String(), stderr.String(), wantErr)
	}
}

func TestInputRealFileTextRules(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trace.jsonl")
	text := []byte("\xef\xbb\xbf# heading\r\n\r\n{\"tool\":\"x\"}\r\n")
	if err := os.WriteFile(path, text, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "keys", path)
	cmd.Env = append(os.Environ(), "TRAILSPIN_INPUT_AS_MAIN=1")
	var stdout, stderr bytesBuffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("program failed: %v: %s", err, stderr.String())
	}
	if ctx.Err() != nil {
		t.Fatalf("program timed out: %v", ctx.Err())
	}
	if stdout.String() != "1 x ok 0 {}\n" || stderr.String() != "" {
		t.Errorf("real CRLF file = (%q, %q)", stdout.String(), stderr.String())
	}
}

type bytesBuffer struct{ data []byte }

func (b *bytesBuffer) Write(p []byte) (int, error) { b.data = append(b.data, p...); return len(p), nil }
func (b *bytesBuffer) String() string              { return string(b.data) }
