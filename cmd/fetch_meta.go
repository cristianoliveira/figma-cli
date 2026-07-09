package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
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
	Run: func(cmd *cobra.Command, args []string) {
		input, err := figma.ParseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		token, err := env.GetFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		apiURL, err := figma.BuildFileURL(input.FileID, input.NodeIDs, "", "1")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building API URL: %v\n", err)
			os.Exit(1)
		}

		client := figma.NewClient(token)
		result, err := client.FetchJSON(apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
			os.Exit(1)
		}

		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(fetchMetaCmd)
}
