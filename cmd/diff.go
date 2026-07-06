package cmd

import "github.com/spf13/cobra"

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff Figma file versions",
}
