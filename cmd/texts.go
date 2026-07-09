package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
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
		if len(input.NodeIDs) > 0 {
			apiURL, err = figma.BuildNodesURL(input.FileID, input.NodeIDs)
		} else {
			apiURL, err = figma.BuildFileURL(input.FileID, nil, "", "")
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building API URL: %v\n", err)
			os.Exit(1)
		}

		figmaJSON, err := client.FetchJSON(apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
			os.Exit(1)
		}

		matches := figma.FindLayerTexts(figmaJSON, layerName, recursive)
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
