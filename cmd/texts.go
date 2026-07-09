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

var textsCmd = &cobra.Command{
	Use:   "texts [file-id-or-url]",
	Short: "Extract text from Figma layers",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		layerName, _ := cmd.Flags().GetString("layer")
		recursive, _ := cmd.Flags().GetBool("recursive")
		if layerName == "" {
			fmt.Fprintln(os.Stderr, "error: --layer is required")
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

		var apiURL string
		var doc any
		if len(input.NodeIDs) > 0 {
			apiURL, err = figma.BuildNodesURL(input.FileID, input.NodeIDs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error building API URL: %v\n", err)
				os.Exit(1)
			}
			var resp api.GetFileNodesResponse
			if err := client.Fetch(apiURL, &resp); err != nil {
				fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
				os.Exit(1)
			}
			// Extract document from the first (or only) node
			for _, node := range resp.Nodes {
				doc, err = figma.UnmarshalDocument(node.Document)
				break // Use first node's document
			}
		} else {
			apiURL, err = figma.BuildFileURL(input.FileID, nil, "", "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error building API URL: %v\n", err)
				os.Exit(1)
			}
			var resp api.GetFileResponse
			if err := client.Fetch(apiURL, &resp); err != nil {
				fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
				os.Exit(1)
			}
			doc, err = figma.UnmarshalDocument(resp.Document)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing document: %v\n", err)
			os.Exit(1)
		}

		matches := figma.FindTextByLayerName(doc, layerName, recursive)
		output, err := json.MarshalIndent(map[string]any{"layer": layerName, "matches": matches}, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func init() {
	textsCmd.Flags().String("layer", "", "layer name to extract text from")
	textsCmd.Flags().Bool("recursive", false, "include text from all descendant nodes")
	rootCmd.AddCommand(textsCmd)
}
