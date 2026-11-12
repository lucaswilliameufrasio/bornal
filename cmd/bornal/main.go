package main

import (
	"os"

	"github.com/lucaswilliameufrasio/bornal/internal/cli"
)

func main() {
	interactive := false
	info, err := os.Stdin.Stat()
	if err == nil && info.Mode()&os.ModeCharDevice != 0 {
		interactive = true
	}
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, os.Getenv, interactive))
}
