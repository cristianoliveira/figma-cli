/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
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
		"emit compatibility JSON instead of default TOON (wraps text/file results)")
	command.SetFlagErrorFunc(cli.NewFlagUsageError)
	command.AddCommand(children...)
	return command
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Commands return errors via RunE; Execute owns structured process-level error
// rendering and stable exit codes.
func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	cli.MarkUsageErrors(rootCmd)
	err := rootCmd.Execute()
	if err == nil {
		return
	}
	if renderErr := cli.RenderError(rootCmd, err); renderErr != nil {
		fmt.Fprintln(os.Stderr, "failed to render structured error")
	}
	os.Exit(cli.ExitCode(err))
}
