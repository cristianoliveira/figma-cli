/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "figma",
	Short: "Explore and inspect Figma designs from the command line",
	Long: `figma is a CLI for querying Figma designs. It accepts file URLs directly
and produces structured JSON — designed for humans and AI agents alike.

Common workflows:

  Explore a file's structure:
    figma fetch-meta <file-url>
    figma components --id <node-id> <file-url>

  Find and inspect layers:
    figma find --name "Button" --id <node-id> <file-url>
    figma inspect --id <node-id> <file-url>

  Extract design tokens:
    figma texts --layer "Header" <file-url>
    figma colors --id <node-id> <file-url>

  Export assets:
    figma export --id <node-id> --format png <file-url>

  Track changes:
    figma versions <file-url>
    figma diff text --from <v1> --to <v2> <file-url>

All commands accept either a Figma file key or a full Figma URL.
Requires FIGMA_ACCESS_TOKEN environment variable.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
