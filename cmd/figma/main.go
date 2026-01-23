package main

import (
	"os"

	"github.com/cristianoliveira/figma-cli/cmd/figma/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
