package main

import (
	"os"

	"github.com/runecode-systems/runecontext/internal/cli"
)

func main() {
	code := cli.Run(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}
