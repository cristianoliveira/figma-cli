/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/spf13/cobra"
)

// rootCmd is the production command tree. Tests build isolated roots with newRootCommand.
var rootCmd = newRootCommand()

func newRootCommand(children ...*cobra.Command) *cobra.Command {
	command := &cobra.Command{
		Use:   "figma",
		Short: "Explore and inspect Figma designs from the command line",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			authentication := "configured"
			if os.Getenv("FIGMA_ACCESS_TOKEN") == "" {
				authentication = "missing FIGMA_ACCESS_TOKEN"
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), `figma — agent-facing Figma inspection CLI
Authentication: %s
Next:
  figma me
  figma inspect "<figma-url>?node-id=<node-id>"
  figma --help
`, authentication)
			return err
		},
		Long: `Query Figma files using a file key or full Figma URL.
Results are structured for scripts and agents. Set FIGMA_ACCESS_TOKEN before
commands that access Figma. Run figma <command> --help for local options.`,
		Example: `  figma meta <url>
  figma inspect "<url>?node-id=42-1"
  figma export --id <node-id> --format png <url>`,
	}
	command.PersistentFlags().Bool("json", false,
		"emit every result as JSON (wraps text/file results in a JSON envelope)")
	command.SetFlagErrorFunc(cli.NewFlagUsageError)
	command.AddCommand(children...)
	return command
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Commands return errors via RunE; Execute prints them with an "error:" prefix
// to stderr (matching the prior cli.Die behaviour) and exits non-zero.
func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	cli.MarkUsageErrors(rootCmd)
	err := rootCmd.Execute()
	if err != nil {
		var exitErr *cli.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(cli.ExitCode(err))
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(cli.ExitCode(err))
	}
}
