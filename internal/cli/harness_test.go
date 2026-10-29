package cli_test

import (
	"io/fs"
	"strings"

	"github.com/Gulliby-development/trailspin/internal/cli"
)

// run executes one command line against in-memory files and standard input.
func run(args []string, files map[string]string, stdin string) cli.Result {
	env := cli.Env{
		ReadFile: func(path string) ([]byte, error) {
			text, ok := files[path]
			if !ok {
				return nil, fs.ErrNotExist
			}
			return []byte(text), nil
		},
		Stdin: strings.NewReader(stdin),
	}
	return cli.Run(args, env)
}
