<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# trailspin

trailspin analyzes JSON Lines (JSONL) traces of tool calls made by an AI agent. It helps agent developers and CI users spot repeated calls, loops, consecutive failures, and token or call budgets exceeded in a completed trace. It reads the full input before analyzing it; it does not tail a live session.

The analyzer is offline and deterministic: it uses no network, model, API key, or clock. The Go implementation uses only the standard library. It is intended for Windows, macOS, and Linux.

[![CI](https://github.com/Gulliby-development/trailspin/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/trailspin/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install and build

Install Go 1.27.0 or newer, then clone the repository and install the command:

```sh
git clone https://github.com/Gulliby-development/trailspin.git
cd trailspin
go install .
```

This installs the `trailspin` command using the current module source. To build a binary in the current directory instead:

```sh
go build -o trailspin .
```

On Windows, use an executable name such as `trailspin.exe` with `go build -o trailspin.exe .`. The examples below use `trailspin`; substitute the built executable path as needed. The module path is `github.com/Gulliby-development/trailspin`.

## Input format and identities

A trace is UTF-8 text with one JSON object per line and one tool call per object. The file is read as a whole. A single leading byte-order mark (U+FEFF) is removed. LF and CRLF line endings are accepted; the final line need not end in a newline. Empty lines, lines containing only spaces or tabs, and lines beginning with `#` in column 1 are ignored. Diagnostics use physical line numbers, while call steps count only valid calls.

Each call requires `tool`, a 1–64 character name containing only ASCII letters, digits, underscore, period, colon, or hyphen. `args` defaults to `{}` and otherwise must be an object; `ok` defaults to `true` and otherwise must be a boolean; `tokens` defaults to zero and otherwise must be an integer from 0 through 1,000,000,000. Unknown fields are ignored only after their values have been validated.

Input bytes must be valid UTF-8; invalid UTF-8 is rejected before parsing. Within each JSON line, syntax errors, non-integer numbers (including in ignored fields), duplicate keys, strings containing U+FFFD, excessive nesting, and trailing non-whitespace are checked in encounter order. A duplicate key is reported as soon as its second key is read, before reading that key's value. Only after JSON validation succeeds are schema checks applied in this fixed order: the value must be an object; `tool` must be present; `tool` must be a valid name; `args` must be an object; `ok` must be a boolean; and `tokens` must be an integer from 0 through 1,000,000,000. Nesting beyond 20 arrays and objects is rejected.

The canonical form of `args` has no insignificant whitespace, sorts object keys by UTF-8 byte order, preserves array order, and writes integers without changing their digits except that `-0` becomes `0`. Strings are surrounded by double quotes and escape only the double quote (`\"`), backslash (`\\`), newline (`\n`), carriage return (`\r`), tab (`\t`), and other characters below U+0020 as `\u00xx` with lowercase hexadecimal digits. Every other character is written literally, including `<`, `>`, `&`, U+007F, U+2028, and characters outside the BMP. Booleans and null are written as `true`, `false`, and `null`. A call's identity key is the tool name, one space, and canonical args—for example, `read_file {"path":"a.txt"}`. Calls with equal keys have the same identity even if their input object keys were ordered differently.

## Commands

```text
trailspin COMMAND [ARGS]
```

Use `-` as TRACE to read standard input. The process reads stdin through EOF before producing a report; it does not stream or tail input. Options are available only for `scan`. Each may be written as `--name N` or `--name=N`, anywhere among the arguments after the command, including before or after TRACE. The last occurrence wins, but every occurrence is validated. Arguments beginning with `-` (except the path `-`) are parsed as options. `keys` and `summary` accept no options.

### `keys TRACE`

Prints one row per call:

```text
STEP TOOL STATUS TOKENS CANONICAL_ARGS
```

STATUS is `ok` or `fail`. Save this complete input as `keys.jsonl`:

```jsonl
{"tool":"read_file","args":{"path":"a.txt"},"tokens":120}
{"tool":"grep","args":{"q":"retry"},"tokens":80}
{"tool":"edit","args":{"path":"b.txt"},"tokens":300}
{"tool":"run_tests","args":{"cmd":"go test","dir":"."},"tokens":40}
```

Then `trailspin keys keys.jsonl` prints:

```text
1 read_file ok 120 {"path":"a.txt"}
2 grep ok 80 {"q":"retry"}
3 edit ok 300 {"path":"b.txt"}
4 run_tests ok 40 {"cmd":"go test","dir":"."}
```

### `summary TRACE`

Prints one row per tool, with columns `TOOL CALLS FAILS TOKENS`, ordered by descending call count and then tool name. A final `total CALLS FAILS TOKENS` row includes all calls; an empty trace prints `total 0 0 0`.

Save this complete input as `summary.jsonl`:

```jsonl
{"tool":"read_file","args":{"path":"a.txt"},"tokens":40000}
{"tool":"grep","args":{"q":"retry"},"tokens":30000}
{"tool":"edit","args":{"path":"b.txt"},"tokens":30001}
{"tool":"run_tests","args":{"cmd":"go test","dir":"."},"tokens":5}
```

Then `trailspin summary summary.jsonl` prints:

```text
edit 1 0 30001
grep 1 0 30000
read_file 1 0 40000
run_tests 1 0 5
total 4 0 100006
```

### `scan TRACE [OPTIONS]`

Scans calls in order. Defaults and permitted ranges are:

| Option | Default | Range |
| --- | ---: | ---: |
| `--window` | 10 | 2–1000 |
| `--max-repeat` | 3 | 1–1000 |
| `--max-errors` | 3 | 1–1000 |
| `--cycle-reps` | 3 | 2–100 |
| `--max-tokens` | disabled | 1–1,000,000,000 |
| `--max-calls` | disabled | 1–100,000 |

Option values must be 1–10 ASCII digits with no leading zero; values must also be in range. Each finding is emitted in this order at a step: repeat, cycle, errors, tokens, calls.

- `STEP repeat KEY xC in last W`: the key occurs exactly `--max-repeat + 1` times in the current window, including this call. A saturated window can produce repeated findings as calls advance; a count that drops and later rises can also be reported again.
- `STEP cycle period P xR: KEY | KEY ...`: the smallest period from 2 through 6 whose last `P × R` keys are repetitions of a non-constant block. A finding starts only when a cycle holds and none held at the preceding step.
- `STEP errors N failed calls in a row`: the consecutive failure count equals `--max-errors`.
- `STEP tokens T over limit M`: the running token total first exceeds `--max-tokens` (equality is not an exceedance).
- `STEP calls over limit M`: the call count reaches `--max-calls + 1`.

With findings, scan prints the findings followed by a count summary and exits 1. With no findings it prints `ok, M calls` (or `ok, 1 call`) and exits 0. Save this complete input as `scan.jsonl`:

```jsonl
{"tool":"read_file","args":{"path":"a.txt"},"tokens":100}
{"tool":"read_file","args":{"path":"a.txt"},"tokens":100}
{"tool":"read_file","args":{"path":"a.txt"},"tokens":100}
{"tool":"read_file","args":{"path":"a.txt"},"tokens":100}
{"tool":"read_file","args":{"path":"a.txt"},"tokens":100}
{"tool":"edit","args":{"path":"b.txt"},"tokens":50}
```

Then `trailspin scan scan.jsonl` prints:

```text
4 repeat read_file {"path":"a.txt"} x4 in last 4
1 finding in 6 calls
```

### Standard input

Use a dash in place of TRACE; stdin is consumed until EOF. For example:

```sh
trailspin keys -
```

Pipe or redirect a completed trace to the command. The resulting output is the same as when that trace is read from a file.

## Exit status and errors

- `0`: successful keys/summary report, or scan found no findings.
- `1`: scan found one or more findings (findings are on stdout).
- `2`: usage or input error.

Usage and input errors are atomic: stdout is empty and the diagnostic is written to stderr with a `trailspin: ` prefix. Input diagnostics identify the path (or `<stdin>`) and, for malformed trace lines, the physical line number. The command is checked first, then options, argument count, and finally the input file. Use `trailspin help`, `trailspin -h`, or `trailspin --help` for usage on stdout. With no arguments, usage is written to stderr and status 2 is returned.

## Development

Run the checks from the module root:

```sh
go test ./...
go vet ./...
gofmt -l .
```

`gofmt -l .` should print no file names. To format the Go source, run:

```sh
go fmt ./...
```

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED
