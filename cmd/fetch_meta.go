package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var fetchMetaCmd = &cobra.Command{
	Use:   "fetch-meta [file-id-or-url]",
	Short: "Fetch metadata for a Figma file",
	Long: `Fetch metadata for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  export FIGMA_ACCESS_TOKEN=your_token
  # Using file ID:
  figma-cli fetch-meta ABCDEFGHIJKLMNOPQRSTUVWXYZ
  # Using Figma URL:
  figma-cli fetch-meta https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		client = client.WithContext(cmd.Context())

		apiURL, err := figma.BuildFileURL(input.FileID, input.NodeIDs, "", "1")
		if err != nil {
			return err
		}

		result, err := client.FetchJSON(apiURL)
		if err != nil {
			return err
		}

		if err := cli.NewPrinter(cmd).JSON(result); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fetchMetaCmd)
}
