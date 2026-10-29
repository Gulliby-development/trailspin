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
	set  func(*settings, int64)
}

// settings holds what the options of a command set.
type settings struct {
	scan scan.Options
	top  int
}

var summaryOptionSpecs = []optionSpec{
	{name: "--top", min: 1, max: 1000, set: func(options *settings, value int64) { options.top = int(value) }},
}

var scanOptionSpecs = []optionSpec{
	{name: "--window", min: 2, max: 1000, set: func(options *settings, value int64) { options.scan.Window = int(value) }},
	{name: "--max-repeat", min: 1, max: 1000, set: func(options *settings, value int64) { options.scan.MaxRepeat = int(value) }},
	{name: "--max-errors", min: 1, max: 1000, set: func(options *settings, value int64) { options.scan.MaxErrors = int(value) }},
	{name: "--cycle-reps", min: 2, max: 100, set: func(options *settings, value int64) { options.scan.CycleReps = int(value) }},
	{name: "--max-tokens", min: 1, max: 1000000000, set: func(options *settings, value int64) { options.scan.MaxTokens = value }},
	{name: "--max-calls", min: 1, max: 100000, set: func(options *settings, value int64) { options.scan.MaxCalls = int(value) }},
}

// parseOptions removes command options from args and returns a diagnostic without
// the common trailspin prefix when parsing fails.
func parseOptions(command string, args []string) (settings, []string, string) {
	options := settings{}
	specs := summaryOptionSpecs
	if command == "scan" {
		options.scan = scan.DefaultOptions()
		specs = scanOptionSpecs
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
		if command == "keys" {
			return settings{}, nil, "unknown option \"" + argument + "\""
		}
		spec, found := findOption(specs, name)
		if !found {
			return settings{}, nil, "unknown option \"" + argument + "\""
		}
		if !hasEquals {
			if index+1 >= len(args) {
				return settings{}, nil, "option " + name + " needs a value"
			}
			index++
			value = args[index]
		}
		parsed, valid := parseOptionValue(value, spec.min, spec.max)
		if !valid {
			return settings{}, nil, "bad " + name + " value \"" + value + "\": expected an integer from " + strconv.FormatInt(spec.min, 10) + " to " + strconv.FormatInt(spec.max, 10)
		}
		spec.set(&options, parsed)
	}
	return options, positional, ""
}

func findOption(specs []optionSpec, name string) (optionSpec, bool) {
	for _, spec := range specs {
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
