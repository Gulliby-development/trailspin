package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMain lets the test binary act as the program itself: with TRAILSPIN_AS_MAIN=1 it runs main().
func TestMain(m *testing.M) {
	if os.Getenv("TRAILSPIN_AS_MAIN") == "1" {
		main()
		return
	}
	os.Exit(m.Run())
}

func program(t *testing.T, stdin []byte, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "TRAILSPIN_AS_MAIN=1")
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return code, stdout.String(), stderr.String()
}

func TestProgramReadsRealFilesAndWritesLfOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "t.jsonl")
	trace := "{\"tool\":\"a\"}\r\n{\"tool\":\"a\"}\r\n{\"tool\":\"a\"}\r\n{\"tool\":\"a\"}\r\n"
	if err := os.WriteFile(path, []byte(trace), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errText := program(t, nil, "scan", path)
	want := "4 repeat a {} x4 in last 4\n1 finding in 4 calls\n"
	if code != 1 || out != want || errText != "" {
		t.Errorf("scan = (%d, %q, %q), want (1, %q, \"\")", code, out, errText, want)
	}
	code, out, errText = program(t, nil, "scan", path+".missing")
	wantErr := "trailspin: " + path + ".missing: cannot read file\n"
	if code != 2 || out != "" || errText != wantErr {
		t.Errorf("missing file = (%d, %q, %q), want (2, \"\", %q)", code, out, errText, wantErr)
	}
}

func TestProgramReadsStandardInputAsBytes(t *testing.T) {
	code, out, errText := program(t, []byte{'{', 0xff, '}'}, "keys", "-")
	want := "trailspin: <stdin>: input is not valid UTF-8\n"
	if code != 2 || out != "" || errText != want {
		t.Errorf("bad bytes = (%d, %q, %q), want (2, \"\", %q)", code, out, errText, want)
	}
	code, out, errText = program(t, []byte("{\"tool\":\"x\",\"tokens\":7}\n"), "summary", "-")
	if code != 0 || out != "x 1 0 7\ntotal 1 0 7\n" || errText != "" {
		t.Errorf("summary = (%d, %q, %q)", code, out, errText)
	}
	code, _, errText = program(t, nil)
	if code != 2 || errText == "" {
		t.Errorf("no arguments = (%d, %q)", code, errText)
	}
}
