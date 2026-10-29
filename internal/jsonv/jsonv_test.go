package jsonv

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseValues(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{input: `{}`, want: map[string]any{}},
		{input: `[]`, want: []any{}},
		{input: `null`, want: nil},
		{input: `[true,false,"x",-0,12345678901234567890]`, want: []any{true, false, "x", json.Number("0"), json.Number("12345678901234567890")}},
		{input: " \t\r\n{\"a\":1} \n", want: map[string]any{"a": json.Number("1")}},
	}
	for _, tc := range tests {
		got, err := Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", tc.input, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Parse(%q) = %#v, want %#v", tc.input, got, tc.want)
		}
	}
}

func TestParseProblems(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `{"a":01}`, want: "invalid JSON"},
		{input: `{"a":NaN}`, want: "invalid JSON"},
		{input: `{"a":"\q"}`, want: "invalid JSON"},
		{input: `{"a":1,}`, want: "invalid JSON"},
		{input: `{} {}`, want: "invalid JSON"},
		{input: `{} x`, want: "invalid JSON"},
		{input: `1.5`, want: "number must be an integer"},
		{input: `1.0`, want: "number must be an integer"},
		{input: `1e3`, want: "number must be an integer"},
		{input: `{"a":1,"\u0061":2}`, want: `duplicate key "a"`},
		{input: `"\ud800"`, want: "invalid string"},
		{input: `"\ufffd"`, want: "invalid string"},
		{input: "\"�\"", want: "invalid string"},
		{input: `{"\ud800":1}`, want: "invalid string"},
		{input: `{"a":"\ud800","b":1.5}`, want: "invalid string"},
	}
	for _, tc := range tests {
		_, err := Parse(tc.input)
		if err == nil || err.Error() != tc.want {
			t.Errorf("Parse(%q) error = %v, want %q", tc.input, err, tc.want)
		}
	}
}

func TestParseEncounterOrder(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `{"a":1.5,"a":2}`, want: "number must be an integer"},
		{input: `{"a":1,"a":1.5}`, want: `duplicate key "a"`},
		{input: `{"a":1,"a":`, want: `duplicate key "a"`},
		{input: `{"a":"\ud800","b":1.5}`, want: "invalid string"},
		{input: `{"a":1.5,`, want: "number must be an integer"},
		{input: `[[[[[[[[[[[[[[[[[[[[0]]]]]]]]]]]]]]]]]]]]`, want: ""},
		{input: `[[[[[[[[[[[[[[[[[[[[[0]]]]]]]]]]]]]]]]]]]]]`, want: "nesting deeper than 20"},
		{input: `[[[[[[[[[[[[[[[[[[[[[`, want: "nesting deeper than 20"},
		{input: `{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[0]}]}]}]}]}]}]}]}]}]}`, want: ""},
		{input: `{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[{"a":[[0]]}]}]}]}]}]}]}]}]}]}`, want: "nesting deeper than 20"},
		{input: `{"a":[{"b":[0]}]}`, want: ""},
		{input: `{"a":{"b":1,"b":2}}`, want: `duplicate key "b"`},
	}
	for _, tc := range tests {
		_, err := Parse(tc.input)
		if tc.want == "" {
			if err != nil {
				t.Errorf("Parse(%q) error = %v, want success", tc.input, err)
			}
		} else if err == nil || err.Error() != tc.want {
			t.Errorf("Parse(%q) error = %v, want %q", tc.input, err, tc.want)
		}
	}
}

func TestCanonical(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `{"z":null,"b":[],"a":{},"n":-0,"big":12345678901234567890}`, want: `{"a":{},"b":[],"big":12345678901234567890,"n":0,"z":null}`},
		{input: `[3,0,-2,true,false,null]`, want: `[3,0,-2,true,false,null]`},
		{input: `{"é":1,"z":2,"A":3}`, want: `{"A":3,"z":2,"é":1}`},
		{input: `"\u0000\b\f\u001f\t\r\n\"\\"`, want: `"\u0000\u0008\u000c\u001f\t\r\n\"\\"`},
		{input: `"<>&\u007f\u2028\u2029\ud83d\ude00"`, want: "\"<>&\x7f\u2028\u2029😀\""},
	}
	for _, tc := range tests {
		value, err := Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", tc.input, err)
			continue
		}
		got := Canonical(value)
		if got != tc.want {
			t.Errorf("Canonical(%q) = %q, want %q", tc.input, got, tc.want)
		}
		if Canonical(value) != got {
			t.Errorf("Canonical(%q) is not stable", tc.input)
		}
		roundTrip, err := Parse(got)
		if err != nil || !reflect.DeepEqual(roundTrip, value) {
			t.Errorf("round trip %q = %#v, %v; want %#v", got, roundTrip, err, value)
		}
	}
}
