package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

var diffTextCmd = &cobra.Command{
	Use:   "text [file-id-or-url] --from version-id --to version-id",
	Short: "Diff text nodes between two Figma file versions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fromVersion, _ := cmd.Flags().GetString("from")
		toVersion, _ := cmd.Flags().GetString("to")
		if fromVersion == "" || toVersion == "" {
			fmt.Fprintln(os.Stderr, "error: --from and --to are required")
			os.Exit(1)
		}

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

		client := figma.NewClient(token)

		fromURL, err := figma.BuildFileURL(input.FileID, input.NodeIDs, fromVersion, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building from URL: %v\n", err)
			os.Exit(1)
		}
		toURL, err := figma.BuildFileURL(input.FileID, input.NodeIDs, toVersion, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building to URL: %v\n", err)
			os.Exit(1)
		}

		var fromResp, toResp api.GetFileResponse
		if err := client.Fetch(fromURL, &fromResp); err != nil {
			fmt.Fprintf(os.Stderr, "error fetching from version: %v\n", err)
			os.Exit(1)
		}
		if err := client.Fetch(toURL, &toResp); err != nil {
			fmt.Fprintf(os.Stderr, "error fetching to version: %v\n", err)
			os.Exit(1)
		}

		fromDoc, err := figma.UnmarshalDocument(fromResp.Document)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing from document: %v\n", err)
			os.Exit(1)
		}
		toDoc, err := figma.UnmarshalDocument(toResp.Document)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing to document: %v\n", err)
			os.Exit(1)
		}

		textDiff := figma.DiffText(
			figma.ExtractTextNodes(fromDoc),
			figma.ExtractTextNodes(toDoc),
		)
		output, err := json.MarshalIndent(textDiff, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func init() {
	diffTextCmd.Flags().String("from", "", "source Figma version ID")
	diffTextCmd.Flags().String("to", "", "target Figma version ID")
	diffCmd.AddCommand(diffTextCmd)
	rootCmd.AddCommand(diffCmd)
}
