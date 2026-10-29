package scan

import (
	"reflect"
	"testing"

	"github.com/Gulliby-development/trailspin/internal/trace"
)

func findingsEqual(got, want []string) bool {
	return reflect.DeepEqual(got, want) || len(got) == 0 && len(want) == 0
}

func TestDefaultOptionsAndEmpty(t *testing.T) {
	if got, want := DefaultOptions(), (Options{Window: 10, MaxRepeat: 3, MaxErrors: 3, CycleReps: 3}); got != want {
		t.Fatalf("DefaultOptions() = %#v, want %#v", got, want)
	}
	for _, calls := range [][]trace.Call{nil, {}} {
		if got := Analyze(calls, DefaultOptions()); len(got) != 0 {
			t.Errorf("Analyze(%v) = %v, want no findings", calls, got)
		}
	}
}

func TestRepeatWindowAndEquality(t *testing.T) {
	tests := []struct {
		name    string
		calls   []trace.Call
		options Options
		want    []string
	}{
		{"below threshold", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}}, Options{Window: 4, MaxRepeat: 2}, nil},
		{"startup threshold", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}}, Options{Window: 4, MaxRepeat: 1}, []string{"2 repeat a {} x2 in last 2"}},
		{"above threshold only exact equality", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}}, Options{Window: 4, MaxRepeat: 1}, []string{"2 repeat a {} x2 in last 2"}},
		{"window eviction and renewed equality", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}, {Step: 3, Tool: "b", Args: "{}"}, {Step: 4, Tool: "b", Args: "{}"}, {Step: 5, Tool: "b", Args: "{}"}}, Options{Window: 2, MaxRepeat: 1}, []string{"2 repeat a {} x2 in last 2", "4 repeat b {} x2 in last 2", "5 repeat b {} x2 in last 2"}},
		{"saturated repeated reports", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "a", Args: "{}"}}, Options{Window: 3, MaxRepeat: 2}, []string{"3 repeat a {} x3 in last 3", "4 repeat a {} x3 in last 3"}},
		{"repeat threshold greater than window", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}}, Options{Window: 2, MaxRepeat: 2}, nil},
		{"different args on same tool", []trace.Call{{Step: 1, Tool: "a", Args: "{\"x\":1}"}, {Step: 2, Tool: "a", Args: "{\"x\":2}"}}, Options{Window: 4, MaxRepeat: 1}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			options := tc.options
			options.MaxErrors = 100
			options.CycleReps = 100
			if got := Analyze(tc.calls, options); !findingsEqual(got, tc.want) {
				t.Errorf("Analyze() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestCyclesPeriodsAndBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		calls []trace.Call
		reps  int
		want  []string
	}{
		{"period 2", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "b", Args: "{}"}}, 2, []string{"4 cycle period 2 x2: a {} | b {}"}},
		{"period 3", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "c", Args: "{}"}, {Step: 4, Tool: "a", Args: "{}"}, {Step: 5, Tool: "b", Args: "{}"}, {Step: 6, Tool: "c", Args: "{}"}}, 2, []string{"6 cycle period 3 x2: a {} | b {} | c {}"}},
		{"period 4", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "c", Args: "{}"}, {Step: 4, Tool: "d", Args: "{}"}, {Step: 5, Tool: "a", Args: "{}"}, {Step: 6, Tool: "b", Args: "{}"}, {Step: 7, Tool: "c", Args: "{}"}, {Step: 8, Tool: "d", Args: "{}"}}, 2, []string{"8 cycle period 4 x2: a {} | b {} | c {} | d {}"}},
		{"period 5", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "c", Args: "{}"}, {Step: 4, Tool: "d", Args: "{}"}, {Step: 5, Tool: "e", Args: "{}"}, {Step: 6, Tool: "a", Args: "{}"}, {Step: 7, Tool: "b", Args: "{}"}, {Step: 8, Tool: "c", Args: "{}"}, {Step: 9, Tool: "d", Args: "{}"}, {Step: 10, Tool: "e", Args: "{}"}}, 2, []string{"10 cycle period 5 x2: a {} | b {} | c {} | d {} | e {}"}},
		{"period 6", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "c", Args: "{}"}, {Step: 4, Tool: "d", Args: "{}"}, {Step: 5, Tool: "e", Args: "{}"}, {Step: 6, Tool: "f", Args: "{}"}, {Step: 7, Tool: "a", Args: "{}"}, {Step: 8, Tool: "b", Args: "{}"}, {Step: 9, Tool: "c", Args: "{}"}, {Step: 10, Tool: "d", Args: "{}"}, {Step: 11, Tool: "e", Args: "{}"}, {Step: 12, Tool: "f", Args: "{}"}}, 2, []string{"12 cycle period 6 x2: a {} | b {} | c {} | d {} | e {} | f {}"}},
		{"insufficient repetitions", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}}, 2, nil},
		{"constant block excluded", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "a", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "a", Args: "{}"}}, 2, nil},
		{"smallest period", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "b", Args: "{}"}, {Step: 5, Tool: "a", Args: "{}"}, {Step: 6, Tool: "b", Args: "{}"}, {Step: 7, Tool: "a", Args: "{}"}, {Step: 8, Tool: "b", Args: "{}"}}, 2, []string{"4 cycle period 2 x2: a {} | b {}"}},
		{"cycle reappears after gap", []trace.Call{{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"}, {Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "b", Args: "{}"}, {Step: 5, Tool: "x", Args: "{}"}, {Step: 6, Tool: "a", Args: "{}"}, {Step: 7, Tool: "b", Args: "{}"}, {Step: 8, Tool: "a", Args: "{}"}, {Step: 9, Tool: "b", Args: "{}"}}, 2, []string{"4 cycle period 2 x2: a {} | b {}", "9 cycle period 2 x2: a {} | b {}"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := Options{Window: 100, MaxRepeat: 100, CycleReps: tc.reps, MaxErrors: 100}
			if got := Analyze(tc.calls, opts); !findingsEqual(got, tc.want) {
				t.Errorf("Analyze() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestCycleSuppressionAcrossPeriods(t *testing.T) {
	calls := []trace.Call{
		{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"},
		{Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "b", Args: "{}"},
		{Step: 5, Tool: "x", Args: "{}"}, {Step: 6, Tool: "c", Args: "{}"},
		{Step: 7, Tool: "d", Args: "{}"}, {Step: 8, Tool: "e", Args: "{}"},
		{Step: 9, Tool: "c", Args: "{}"}, {Step: 10, Tool: "d", Args: "{}"},
		{Step: 11, Tool: "e", Args: "{}"},
	}
	opts := Options{Window: 100, MaxRepeat: 100, CycleReps: 2, MaxErrors: 100}
	want := []string{"4 cycle period 2 x2: a {} | b {}", "11 cycle period 3 x2: c {} | d {} | e {}"}
	if got := Analyze(calls, opts); !reflect.DeepEqual(got, want) {
		t.Errorf("Analyze() = %#v, want %#v", got, want)
	}

	keys := []string{"a", "b", "a", "b", "a", "b", "a", "b"}
	if ok, period := hasCycle(keys, len(keys), 2); !ok || period != 2 {
		t.Errorf("hasCycle() = (%t, %d), want (true, 2) when periods 2 and 4 qualify", ok, period)
	}
	calls = []trace.Call{
		{Step: 1, Tool: "a", Args: "{}"}, {Step: 2, Tool: "b", Args: "{}"},
		{Step: 3, Tool: "a", Args: "{}"}, {Step: 4, Tool: "b", Args: "{}"},
		{Step: 5, Tool: "a", Args: "{}"}, {Step: 6, Tool: "b", Args: "{}"},
		{Step: 7, Tool: "a", Args: "{}"}, {Step: 8, Tool: "b", Args: "{}"},
	}
	if got := Analyze(calls, opts); !reflect.DeepEqual(got, []string{"4 cycle period 2 x2: a {} | b {}"}) {
		t.Errorf("simultaneous periods findings = %#v", got)
	}
}

func TestErrorStreaks(t *testing.T) {
	tests := []struct {
		calls     []trace.Call
		threshold int
		want      []string
	}{
		{[]trace.Call{{Step: 1, OK: false}, {Step: 2, OK: false}, {Step: 3, OK: true}, {Step: 4, OK: false}, {Step: 5, OK: false}}, 2, []string{"2 errors 2 failed calls in a row", "5 errors 2 failed calls in a row"}},
		{[]trace.Call{{Step: 1, OK: false}, {Step: 2, OK: false}, {Step: 3, OK: false}}, 1, []string{"1 errors 1 failed calls in a row"}},
		{[]trace.Call{{Step: 1, OK: false}, {Step: 2, OK: true}, {Step: 3, OK: false}}, 2, nil},
	}
	for _, tc := range tests {
		opts := Options{Window: 100, MaxRepeat: 100, MaxErrors: tc.threshold, CycleReps: 100}
		if got := Analyze(tc.calls, opts); !findingsEqual(got, tc.want) {
			t.Errorf("Analyze() = %#v, want %#v", got, tc.want)
		}
	}
}

func TestBudgets(t *testing.T) {
	calls := []trace.Call{
		{Step: 1, Tool: "a", Args: "{}", OK: true, Tokens: 1000000000},
		{Step: 2, Tool: "b", Args: "{}", OK: true, Tokens: 1},
		{Step: 3, Tool: "c", Args: "{}", OK: true, Tokens: 5},
	}
	base := Options{Window: 100, MaxRepeat: 100, MaxErrors: 100, CycleReps: 100}
	if got := Analyze(calls, base); len(got) != 0 {
		t.Errorf("disabled budgets: %v", got)
	}
	opts := base
	opts.MaxTokens = 1000000000
	if got := Analyze(calls[:1], opts); len(got) != 0 {
		t.Errorf("exact token budget: %v", got)
	}
	want := []string{"2 tokens 1000000001 over limit 1000000000"}
	if got := Analyze(calls, opts); !reflect.DeepEqual(got, want) {
		t.Errorf("first token exceedance = %#v, want %#v", got, want)
	}
	opts.MaxTokens = 999999999
	want = []string{"1 tokens 1000000000 over limit 999999999"}
	if got := Analyze(calls, opts); !reflect.DeepEqual(got, want) {
		t.Errorf("initial token exceedance = %#v, want %#v", got, want)
	}
	opts = base
	opts.MaxCalls = 2
	want = []string{"3 calls over limit 2"}
	if got := Analyze(calls, opts); !reflect.DeepEqual(got, want) {
		t.Errorf("call budget = %#v, want %#v", got, want)
	}
	opts.MaxCalls = 3
	if got := Analyze(calls, opts); len(got) != 0 {
		t.Errorf("call limit M: %v", got)
	}
}

func TestSimultaneousFindingsAndInvocationState(t *testing.T) {
	calls := []trace.Call{
		{Step: 1, Tool: "a", Args: "{}", OK: false, Tokens: 2},
		{Step: 2, Tool: "b", Args: "{}", OK: false, Tokens: 2},
		{Step: 3, Tool: "a", Args: "{}", OK: false, Tokens: 2},
		{Step: 4, Tool: "b", Args: "{}", OK: false, Tokens: 2},
	}
	original := append([]trace.Call(nil), calls...)
	opts := Options{Window: 4, MaxRepeat: 1, MaxErrors: 4, CycleReps: 2, MaxTokens: 7, MaxCalls: 3}
	want := []string{"3 repeat a {} x2 in last 3", "4 repeat b {} x2 in last 4", "4 cycle period 2 x2: a {} | b {}", "4 errors 4 failed calls in a row", "4 tokens 8 over limit 7", "4 calls over limit 3"}
	for i := 0; i < 2; i++ {
		if got := Analyze(calls, opts); !reflect.DeepEqual(got, want) {
			t.Errorf("Analyze() = %#v, want %#v", got, want)
		}
	}
	if !reflect.DeepEqual(calls, original) {
		t.Errorf("Analyze mutated calls: %#v", calls)
	}
}
