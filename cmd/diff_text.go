package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	figmadiff "github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/spf13/cobra"
)

var diffTextCmd = &cobra.Command{
	Use:   "text [file-id-or-url] --from version-id --to version-id",
	Short: "Diff text nodes between two Figma file versions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fromVersion, _ := cmd.Flags().GetString("from")
		toVersion, _ := cmd.Flags().GetString("to")
		nodeID, _ := cmd.Flags().GetString("id")
		if fromVersion == "" || toVersion == "" {
			fmt.Fprintln(os.Stderr, "error: --from and --to are required")
			os.Exit(1)
		}

		input, err := parseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		token, err := getFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		nodeIDs := resolveNodeIDs(input, nodeID)
		fromURL, err := figmadiff.BuildFileVersionURL(input.fileID, nodeIDs, fromVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building from URL: %v\n", err)
			os.Exit(1)
		}
		toURL, err := figmadiff.BuildFileVersionURL(input.fileID, nodeIDs, toVersion)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building to URL: %v\n", err)
			os.Exit(1)
		}

		fromJSON, err := figmadiff.FetchFigmaJSON(fromURL, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching from version: %v\n", err)
			os.Exit(1)
		}
		toJSON, err := figmadiff.FetchFigmaJSON(toURL, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching to version: %v\n", err)
			os.Exit(1)
		}

		textDiff := figmadiff.Text(
			figmadiff.ExtractTextNodes(fromJSON["document"]),
			figmadiff.ExtractTextNodes(toJSON["document"]),
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
	diffTextCmd.Flags().String("id", "", "node ID to diff; accepts 20089:685897, 20089-685897, or comma-separated IDs")
	diffCmd.AddCommand(diffTextCmd)
	rootCmd.AddCommand(diffCmd)
}
