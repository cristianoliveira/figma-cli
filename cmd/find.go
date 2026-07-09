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

type layerMatchOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

var findCmd = &cobra.Command{
	Use:   "find [figma-url-or-file-id]",
	Short: "Find Figma layers by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		layerName, _ := cmd.Flags().GetString("name")
		nodeID, _ := cmd.Flags().GetString("id")
		if layerName == "" {
			fmt.Fprintln(os.Stderr, "error: --name is required")
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

		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		apiURL, err := figma.BuildFileURL(input.FileID, nodeIDs, "", "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building find URL: %v\n", err)
			os.Exit(1)
		}

		client := figma.NewClient(token)
		var resp api.GetFileResponse
		if err := client.Fetch(apiURL, &resp); err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
			os.Exit(1)
		}
		doc, err := figma.UnmarshalDocument(resp.Document)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing document: %v\n", err)
			os.Exit(1)
		}

		matches := findLayersByName(doc, layerName)
		output, err := json.MarshalIndent(map[string]any{"name": layerName, "matches": matches}, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func findLayersByName(value any, layerName string) []layerMatchOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	var matches []layerMatchOutput
	if object["name"] == layerName {
		matches = append(matches, layerMatchOutput{
			ID:   figma.StringValue(object["id"]),
			Name: figma.StringValue(object["name"]),
			Type: figma.StringValue(object["type"]),
		})
	}

	children, ok := object["children"].([]any)
	if !ok {
		return matches
	}
	for _, child := range children {
		matches = append(matches, findLayersByName(child, layerName)...)
	}
	return matches
}

func init() {
	findCmd.Flags().String("name", "", "exact layer name to find")
	findCmd.Flags().String("id", "", "node ID to search within; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(findCmd)
}
