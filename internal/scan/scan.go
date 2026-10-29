// Package scan detects repeated calls, cycles, failure streaks, and exceeded budgets.
package scan

import (
	"fmt"

	"github.com/Gulliby-development/trailspin/internal/trace"
)

// Options configures the ordered scan detectors. A zero budget disables that budget.
type Options struct {
	Window    int
	MaxRepeat int
	MaxErrors int
	CycleReps int
	MaxTokens int64
	MaxCalls  int
}

// DefaultOptions returns the detector defaults used by the scan command.
func DefaultOptions() Options {
	return Options{Window: 10, MaxRepeat: 3, MaxErrors: 3, CycleReps: 3}
}

// Analyze returns finding lines in call order without trailing line feeds.
// Calls and options must already have been validated by the caller.
func Analyze(calls []trace.Call, options Options) []string {
	findings := make([]string, 0)
	failed := 0
	var tokens int64
	tokensExceeded := false
	previousCycle := false
	keys := make([]string, len(calls))
	for i, call := range calls {
		key := call.Key()
		keys[i] = key

		start := i + 1 - options.Window
		if start < 0 {
			start = 0
		}
		count := 0
		for j := start; j <= i; j++ {
			if keys[j] == key {
				count++
			}
		}
		if count == options.MaxRepeat+1 {
			findings = append(findings, fmt.Sprintf("%d repeat %s x%d in last %d", call.Step, key, count, i+1-start))
		}

		cycle, period := hasCycle(keys, i+1, options.CycleReps)
		if cycle && !previousCycle {
			blockStart := i + 1 - period
			block := keys[blockStart : blockStart+period]
			findings = append(findings, fmt.Sprintf("%d cycle period %d x%d: %s", call.Step, period, options.CycleReps, joinKeys(block)))
		}
		previousCycle = cycle

		if !call.OK {
			failed++
		} else {
			failed = 0
		}
		if failed == options.MaxErrors {
			findings = append(findings, fmt.Sprintf("%d errors %d failed calls in a row", call.Step, failed))
		}

		if options.MaxTokens > 0 && !tokensExceeded {
			tokens += call.Tokens
			if tokens > options.MaxTokens {
				findings = append(findings, fmt.Sprintf("%d tokens %d over limit %d", call.Step, tokens, options.MaxTokens))
				tokensExceeded = true
			}
		}
		if options.MaxCalls > 0 && i+1 == options.MaxCalls+1 {
			findings = append(findings, fmt.Sprintf("%d calls over limit %d", call.Step, options.MaxCalls))
		}
	}
	return findings
}

// hasCycle reports the smallest qualifying period ending at the current step.
func hasCycle(keys []string, length, repetitions int) (bool, int) {
	for period := 2; period <= 6; period++ {
		needed := period * repetitions
		if length < needed {
			continue
		}
		start := length - needed
		matches := true
		for i := start + period; i < length; i++ {
			if keys[i] != keys[start+(i-start)%period] {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		constant := true
		for i := 1; i < period; i++ {
			if keys[start+i] != keys[start] {
				constant = false
				break
			}
		}
		if !constant {
			return true, period
		}
	}
	return false, 0
}

func joinKeys(keys []string) string {
	joined := ""
	for i, key := range keys {
		if i > 0 {
			joined += " | "
		}
		joined += key
	}
	return joined
}
