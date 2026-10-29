package report_test

import (
	"testing"

	"github.com/Gulliby-development/trailspin/internal/report"
	"github.com/Gulliby-development/trailspin/internal/trace"
)

func TestSummaryTop(t *testing.T) {
	calls := []trace.Call{
		{Step: 1, Tool: "b", OK: true, Tokens: 5},
		{Step: 2, Tool: "a", OK: false, Tokens: 7},
		{Step: 3, Tool: "a", OK: true, Tokens: 11},
		{Step: 4, Tool: "c", OK: false, Tokens: 13},
		{Step: 5, Tool: "c", OK: false, Tokens: 17},
		{Step: 6, Tool: "d", OK: true, Tokens: 19},
	}
	tests := []struct {
		name string
		top  int
		want string
	}{
		{name: "all tools", top: 0, want: "a 2 1 18\nc 2 2 30\nb 1 0 5\nd 1 0 19\ntotal 6 3 72\n"},
		{name: "top one", top: 1, want: "a 2 1 18\n(others) 4 2 54\ntotal 6 3 72\n"},
		{name: "top three", top: 3, want: "a 2 1 18\nc 2 2 30\nb 1 0 5\n(others) 1 0 19\ntotal 6 3 72\n"},
		{name: "top as many as tools", top: 4, want: "a 2 1 18\nc 2 2 30\nb 1 0 5\nd 1 0 19\ntotal 6 3 72\n"},
		{name: "top above the tool count", top: 9, want: "a 2 1 18\nc 2 2 30\nb 1 0 5\nd 1 0 19\ntotal 6 3 72\n"},
	}
	for _, tc := range tests {
		if got := report.Summary(calls, tc.top); got != tc.want {
			t.Errorf("%s: Summary(top %d) = %q, want %q", tc.name, tc.top, got, tc.want)
		}
	}
}

func TestSummaryTopOfAnEmptyTrace(t *testing.T) {
	if got := report.Summary(nil, 3); got != "total 0 0 0\n" {
		t.Errorf("Summary(nil, 3) = %q", got)
	}
}
