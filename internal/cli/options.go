package cli

import (
	"strconv"
	"strings"

	"github.com/Gulliby-development/trailspin/internal/scan"
)

type optionSpec struct {
	name string
	min  int64
	max  int64
	set  func(*scan.Options, int64)
}

var scanOptionSpecs = []optionSpec{
	{name: "--window", min: 2, max: 1000, set: func(options *scan.Options, value int64) { options.Window = int(value) }},
	{name: "--max-repeat", min: 1, max: 1000, set: func(options *scan.Options, value int64) { options.MaxRepeat = int(value) }},
	{name: "--max-errors", min: 1, max: 1000, set: func(options *scan.Options, value int64) { options.MaxErrors = int(value) }},
	{name: "--cycle-reps", min: 2, max: 100, set: func(options *scan.Options, value int64) { options.CycleReps = int(value) }},
	{name: "--max-tokens", min: 1, max: 1000000000, set: func(options *scan.Options, value int64) { options.MaxTokens = value }},
	{name: "--max-calls", min: 1, max: 100000, set: func(options *scan.Options, value int64) { options.MaxCalls = int(value) }},
}

// parseOptions removes command options from args and returns a diagnostic without
// the common trailspin prefix when parsing fails.
func parseOptions(command string, args []string) (scan.Options, []string, string) {
	options := scan.Options{}
	if command == "scan" {
		options = scan.DefaultOptions()
	}

	positional := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "-" || !strings.HasPrefix(argument, "-") {
			positional = append(positional, argument)
			continue
		}

		name, value, hasEquals := argument, "", false
		if equals := strings.IndexByte(argument, '='); equals >= 0 {
			name, value, hasEquals = argument[:equals], argument[equals+1:], true
		}
		if command != "scan" {
			return scan.Options{}, nil, "unknown option \"" + argument + "\""
		}
		spec, found := findOption(name)
		if !found {
			return scan.Options{}, nil, "unknown option \"" + argument + "\""
		}
		if !hasEquals {
			if index+1 >= len(args) {
				return scan.Options{}, nil, "option " + name + " needs a value"
			}
			index++
			value = args[index]
		}
		parsed, valid := parseOptionValue(value, spec.min, spec.max)
		if !valid {
			return scan.Options{}, nil, "bad " + name + " value \"" + value + "\": expected an integer from " + strconv.FormatInt(spec.min, 10) + " to " + strconv.FormatInt(spec.max, 10)
		}
		spec.set(&options, parsed)
	}
	return options, positional, ""
}

func findOption(name string) (optionSpec, bool) {
	for _, spec := range scanOptionSpecs {
		if spec.name == name {
			return spec, true
		}
	}
	return optionSpec{}, false
}

func parseOptionValue(value string, min, max int64) (int64, bool) {
	if len(value) == 0 || len(value) > 10 || value[0] == '0' {
		return 0, false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed < min || parsed > max {
		return 0, false
	}
	return parsed, true
}
