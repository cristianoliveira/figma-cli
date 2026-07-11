package main

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/cmd"
)

func main() {
	command := cmd.NewPixelPerfectCommand()
	command.SilenceErrors = true
	command.SilenceUsage = true
	if err := command.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
