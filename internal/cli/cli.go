package cli

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ppp16bit/subnetting/internal/network"
	"github.com/ppp16bit/subnetting/internal/ui"
)

type options struct {
	colorMode ui.ColorMode
	help      bool
	network   string
	expanded  bool
	split     string
	offset    *big.Int
	limit     int
	paging    bool
}

func Run(args []string) int { return run(args, os.Stdout, os.Stderr, ui.Run) }

func run(args []string, stdout, stderr io.Writer, startUI func(ui.ColorMode) error) int {
	options, err := parseArgs(args)
	if err != nil {
		return argumentError(stderr, err)
	}
	if options.help {
		if _, err := fmt.Fprint(stdout, helpText); err != nil {
			fmt.Fprintf(stderr, "subnetting: %v\n", err)
			return 1
		}
		return 0
	}
	if options.network != "" {
		output, err := calculateOutput(options)
		if err != nil {
			return argumentError(stderr, err)
		}
		if _, err := io.WriteString(stdout, output); err != nil {
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

func argumentError(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "subnetting: %v\n", err)
	fmt.Fprintln(stderr, "Try 'subnetting --help' for more information.")
	return 2
}

func parseArgs(args []string) (options, error) {
	o := options{colorMode: ui.ColorAuto, offset: new(big.Int), limit: 20}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help":
			o.help = true
			continue
		case "--expanded":
			o.expanded = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			name, value, hasValue := strings.Cut(arg, "=")
			switch name {
			case "--color", "--split", "--offset", "--limit":
			default:
				return o, fmt.Errorf("unexpected argument %q", arg)
			}
			if !hasValue {
				if i+1 >= len(args) {
					return o, fmt.Errorf("%s requires a value", name)
				}
				i++
				value = args[i]
			}
			switch name {
			case "--color":
				mode, err := ui.ParseColorMode(value)
				if err != nil {
					return o, err
				}
				o.colorMode = mode
			case "--split":
				if value == "" {
					return o, errors.New("--split requires a child prefix")
				}
				o.split = value
			case "--offset":
				n, ok := new(big.Int).SetString(value, 10)
				if !ok || n.Sign() < 0 {
					return o, errors.New("--offset must be a nonnegative integer")
				}
				o.offset, o.paging = n, true
			case "--limit":
				n, err := strconv.Atoi(value)
				if err != nil || n < 1 || n > network.MaxPageSize {
					return o, fmt.Errorf("--limit must be between 1 and %d", network.MaxPageSize)
				}
				o.limit, o.paging = n, true
			}
		} else if strings.HasPrefix(arg, "-") || o.network != "" {
			return o, fmt.Errorf("unexpected argument %q", arg)
		} else {
			o.network = arg
		}
	}
	if !o.help {
		if o.network == "" && (o.expanded || o.split != "" || o.paging) {
			return o, errors.New("calculation options require an IP/CIDR argument")
		}
		if o.paging && o.split == "" {
			return o, errors.New("--offset and --limit require --split")
		}
	}
	return o, nil
}

const helpText = `Usage: subnetting [options] [IP/CIDR]

IPv4 and IPv6 subnet calculator and learning tool.
Without IP/CIDR, open the interactive TUI. Otherwise print network details.

Options:
  --color=auto|always|never  TUI color mode (default auto)
  --expanded                Expand IPv6 addresses in CLI output
  --split /PREFIX           Split into equal-size child prefixes
  --offset INTEGER          First child index, starting at 0 (default 0)
  --limit INTEGER           Children to display, 1-1000 (default 20)
  -h, --help                Show this help

Examples:
  subnetting 2001:db8::/32
  subnetting 2001:db8:1234::/48 --split /64 --limit 20
  subnetting 192.168.1.0/24 --split /26

Environment:
  NO_COLOR                  Disable colors in auto mode when non-empty
`
