package cmd

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/spf13/cobra"
)

// ExecuteRoot runs a fully constructed root command. It owns the process
// entry contract: structured error rendering and stable exit codes.
//
// Composition lives in cmd/figma/main.go, which builds the dependencies and
// the root via newRootCommand before calling ExecuteRoot. Tests can also
// call ExecuteRoot with isolated roots and verify state isolation without
// touching process exit codes.
func ExecuteRoot(root *cobra.Command) {
	root.SilenceErrors = true
	root.SilenceUsage = true
	cli.MarkUsageErrors(root)
	err := root.Execute()
	if err == nil {
		return
	}
	if renderErr := cli.RenderError(root, err); renderErr != nil {
		fmt.Fprintln(os.Stderr, "failed to render structured error")
	}
	os.Exit(cli.ExitCode(err))
}
