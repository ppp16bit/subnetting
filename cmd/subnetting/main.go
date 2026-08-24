package main

import (
	"os"

	"github.com/ppp16bit/subnetting/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
