package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var commentsCmd = &cobra.Command{
	Use:   "comments [file-id-or-url]",
	Short: "Fetch comments for a Figma file",
	Long: `Fetch comments for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  figma comments grnVU2vAihHXwYgHryu2xE
  figma comments https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=4-1082&p=f&m=dev`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSimpleFetch(args, func(fileID string) (string, error) {
			return figma.BuildCommentsURL(fileID), nil
		}, "comments")
	},
}

func init() {
	rootCmd.AddCommand(commentsCmd)
}
