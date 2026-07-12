/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/spf13/cobra"
)

// rootCmd is the production command tree. Tests build isolated roots with newRootCommand.
var rootCmd = newRootCommand()

func newRootCommand(children ...*cobra.Command) *cobra.Command {
	command := &cobra.Command{
		Use:   "figma",
		Short: "Explore and inspect Figma designs from the command line",
		Long: `figma is a CLI for querying Figma designs. It accepts file URLs directly
and produces structured JSON — designed for humans and AI agents alike.

  What is this file about?
    figma meta <url>

  What's inside this screen?
    figma components --id <node-id> <url>

  Where is the "Button" layer?
    figma find --name "Button" --id <node-id> <url>

  Show me the details of that button.
    figma inspect "<url>?node-id=42-1"

  How is this frame laid out?
    figma layout "<url>?node-id=42-1"

  What text is on this screen?
    figma texts "<url>?node-id=42-1"
    figma texts --id 42:1 <url>
    figma texts --layer "Screen Name" <url>

  What colors does this frame use?
    figma colors --id <node-id> <url>

  How has the copy changed?
    figma versions <url>
    figma diff text --from <v1> --to <v2> <url>

  Export this node as PNG.
    figma export --id <node-id> --format png <url>

All commands accept either a Figma file key or a full Figma URL.
Requires FIGMA_ACCESS_TOKEN environment variable.`,
	}
	command.PersistentFlags().Bool("json", false,
		"emit every result as JSON (wraps text/file results in a JSON envelope)")
	command.SetFlagErrorFunc(flagErrorWithAvailableOptions)
	command.AddCommand(children...)
	return command
}

func flagErrorWithAvailableOptions(cmd *cobra.Command, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w\n\n%s", err, availableFlagHelp(cmd))
}

func availableFlagHelp(cmd *cobra.Command) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Available flags for %q:\n", cmd.CommandPath())
	localFlags := strings.TrimRight(cmd.LocalFlags().FlagUsages(), "\n")
	if localFlags == "" {
		builder.WriteString("  (none)")
	} else {
		builder.WriteString(localFlags)
	}
	inheritedFlags := strings.TrimRight(cmd.InheritedFlags().FlagUsages(), "\n")
	if inheritedFlags != "" {
		builder.WriteString("\n\nGlobal flags:\n")
		builder.WriteString(inheritedFlags)
	}
	fmt.Fprintf(&builder, "\n\nRun `%s --help` for details.", cmd.CommandPath())
	return builder.String()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
// Commands return errors via RunE; Execute prints them with an "error:" prefix
// to stderr (matching the prior cli.Die behaviour) and exits non-zero.
func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	err := rootCmd.Execute()
	if err != nil {
		var exitErr *cli.ExitCodeError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
