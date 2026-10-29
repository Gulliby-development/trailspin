// Package report formats deterministic reports for validated tool-call traces.
package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Gulliby-development/trailspin/internal/trace"
)

// Keys returns one canonical identity row per call, terminated by LF.
func Keys(calls []trace.Call) string {
	var output strings.Builder
	for _, call := range calls {
		status := "fail"
		if call.OK {
			status = "ok"
		}
		fmt.Fprintf(&output, "%d %s %s %d %s\n", call.Step, call.Tool, status, call.Tokens, call.Args)
	}
	return output.String()
}

// Summary returns per-tool totals ordered by descending call count and then tool name.
// With top above zero, only the first top tools get a row and one "(others)" row
// adds up the rest; the total row always covers every call.
func Summary(calls []trace.Call, top int) string {
	type totals struct {
		calls  int
		fails  int
		tokens int64
	}

	byTool := make(map[string]totals)
	var all totals
	for _, call := range calls {
		tool := byTool[call.Tool]
		tool.calls++
		if !call.OK {
			tool.fails++
		}
		tool.tokens += call.Tokens
		byTool[call.Tool] = tool

		all.calls++
		if !call.OK {
			all.fails++
		}
		all.tokens += call.Tokens
	}

	tools := make([]string, 0, len(byTool))
	for tool := range byTool {
		tools = append(tools, tool)
	}
	sort.Slice(tools, func(i, j int) bool {
		left, right := byTool[tools[i]], byTool[tools[j]]
		if left.calls != right.calls {
			return left.calls > right.calls
		}
		return tools[i] < tools[j]
	})

	var output strings.Builder
	var others totals
	for index, name := range tools {
		tool := byTool[name]
		if top > 0 && index >= top {
			others.calls += tool.calls
			others.fails += tool.fails
			others.tokens += tool.tokens
			continue
		}
		fmt.Fprintf(&output, "%s %d %d %d\n", name, tool.calls, tool.fails, tool.tokens)
	}
	if others.calls > 0 {
		fmt.Fprintf(&output, "(others) %d %d %d\n", others.calls, others.fails, others.tokens)
	}
	fmt.Fprintf(&output, "total %d %d %d\n", all.calls, all.fails, all.tokens)
	return output.String()
}
