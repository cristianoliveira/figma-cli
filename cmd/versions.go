package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions [file-id-or-url]",
	Short: "Fetch version history for a Figma file",
	Long: `Fetch version history for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  figma versions grnVU2vAihHXwYgHryu2xE
  figma versions https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=4-1082&p=f&m=dev`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunSimpleFetch(cmd, args, func(fileID string) (string, error) {
			return figma.BuildVersionsURL(fileID), nil
		}, "versions")
	},
}

func init() {
	rootCmd.AddCommand(versionsCmd)
}
