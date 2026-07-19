package cmd

import "github.com/spf13/cobra"

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Diff Figma file versions",
	Example: `  figma diff text --from <version-id> --to <version-id> <url>
  figma diff blame --to <version-id> --id 42:1 <url>`,
}
