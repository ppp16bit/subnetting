# Subnetting

A fast, interactive IPv4 subnet calculator and learning tool for the terminal.
Enter an address in CIDR notation and see the network details update as you
type. Optional panels explain the calculation without taking focus from the
calculator.

![Subnetting calculator waiting for IPv4 input](screenshots/tui.png)

## What it calculates

- Subnet mask
- Network address
- Broadcast address
- First and last usable addresses
- Number of usable hosts
- The CIDR-to-mask calculation and containing subnet block
- The bitwise AND that produces the network address

## Run

Requires Go 1.26.3 or later.

Run the application directly during development:

```sh
make run
```

The equivalent Go command is:

```sh
go run ./cmd/subnetting
```

Color follows the terminal's ANSI palette, so the interface adapts to light,
dark, high-contrast, and customized terminal themes without detecting a theme
by name. The default mode enables color only when the output supports it and
respects a non-empty `NO_COLOR` environment variable:

```sh
go run ./cmd/subnetting --color=auto
go run ./cmd/subnetting --color=always
go run ./cmd/subnetting --color=never
NO_COLOR=1 go run ./cmd/subnetting
```

Type an IPv4 address followed by its CIDR prefix:

```text
192.168.1.0/24
```

The calculation appears immediately:

![Subnetting calculator displaying results for 192.168.1.0/24](screenshots/input_ex.png)

## Controls

| Key | Action |
| --- | --- |
| Type | Enter or edit an IPv4/CIDR address |
| `L` | Toggle the live step-by-step learning panel |
| `B` | Toggle the binary visualization panel |
| `?` | Toggle keyboard help |
| `Q` | Quit |
| `Esc` | Quit |
| `Ctrl+C` | Quit |

Panel state is kept for the duration of the session. On wide terminals the
optional panels appear beside the calculator; on narrow terminals they stack
below it.

## Build

Create and run a standalone executable in `bin/subnetting`:

```sh
make build
./bin/subnetting
```

The equivalent Go commands are:

```sh
mkdir -p bin
go build -o bin/subnetting ./cmd/subnetting
```

## Development commands

| Command | Purpose |
| --- | --- |
| `make run` | Run the application during development |
| `make build` | Build `bin/subnetting` |
| `make test` | Run all Go tests |
| `make vet` | Analyze all packages with `go vet` |
| `make fmt` | Format all Go packages |
| `make install` | Install the command in `~/.local/bin` by default |
| `make uninstall` | Remove the installed `subnetting` binary |
| `make clean` | Remove the generated `bin/` directory |

## Install

Install `subnetting` for the current user:

```sh
make install
```

The default destination is `~/.local/bin/subnetting`. Override the directory
when needed without changing `GOBIN` or `GOPATH`:

```sh
make install INSTALL_DIR="$HOME/bin"
sudo make install INSTALL_DIR=/usr/local/bin
```

The Makefile does not invoke `sudo` or modify shell configuration. To run
`subnetting` from any directory, ensure `~/.local/bin` is in `PATH`. For
example:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

Remove only the installed executable with:

```sh
make uninstall
```

Only IPv4 addresses are supported.
