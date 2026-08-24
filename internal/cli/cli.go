package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ppp16bit/subnetting/internal/ui"
)

type options struct {
	colorMode ui.ColorMode
	help      bool
}

func Run(args []string) int {
	return run(args, os.Stdout, os.Stderr, ui.Run)
}

func run(args []string, stdout, stderr io.Writer, startUI func(ui.ColorMode) error) int {
	options, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(stderr, "subnetting: %v\n", err)
		fmt.Fprintln(stderr, "Try 'subnetting --help' for more information.")
		return 2
	}

	if options.help {
		if _, err := fmt.Fprint(stdout, helpText); err != nil {
			fmt.Fprintf(stderr, "subnetting: %v\n", err)
			return 1
		}
		return 0
	}

	if err := startUI(options.colorMode); err != nil {
		fmt.Fprintf(stderr, "Fatal error: %v\n", err)
		return 1
	}
	return 0
}

func parseArgs(args []string) (options, error) {
	options := options{colorMode: ui.ColorAuto}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			options.help = true
		case arg == "--color":
			if i+1 >= len(args) {
				return options, errors.New("--color requires auto, always, or never")
			}
			i++
			mode, err := ui.ParseColorMode(args[i])
			if err != nil {
				return options, err
			}
			options.colorMode = mode
		case strings.HasPrefix(arg, "--color="):
			mode, err := ui.ParseColorMode(strings.TrimPrefix(arg, "--color="))
			if err != nil {
				return options, err
			}
			options.colorMode = mode
		default:
			return options, fmt.Errorf("unexpected argument %q", arg)
		}
	}
	return options, nil
}

const helpText = `Usage: subnetting [options]

Interactive IPv4 subnet calculator and learning tool.

Options:
  --color=auto|always|never  TUI color mode (default auto)
  -h, --help                 Show this help

Environment:
  NO_COLOR                   Disable colors in auto mode when non-empty
`
