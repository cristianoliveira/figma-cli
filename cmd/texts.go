package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	figmadiff "github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/spf13/cobra"
)

var textsCmd = &cobra.Command{
	Use:   "texts [file-id-or-url]",
	Short: "Extract text from Figma layers",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		layerName, _ := cmd.Flags().GetString("layer")
		recursive, _ := cmd.Flags().GetBool("recursive")
		nodeID, _ := cmd.Flags().GetString("id")
		if layerName == "" {
			fmt.Fprintln(os.Stderr, "error: --layer is required")
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

		apiURL, err := buildTextsAPIURL(input.fileID, resolveNodeIDs(input, nodeID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building texts API URL: %v\n", err)
			os.Exit(1)
		}
		figmaJSON, err := figmadiff.FetchFigmaJSON(apiURL, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
			os.Exit(1)
		}

		matches := findLayerTextsInFigmaJSON(figmaJSON, layerName, recursive)
		output, err := json.MarshalIndent(map[string]any{"layer": layerName, "matches": matches}, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func buildTextsAPIURL(fileID string, nodeIDs []string) (string, error) {
	apiURL := fmt.Sprintf("https://api.figma.com/v1/files/%s", fileID)
	if len(nodeIDs) == 0 {
		return apiURL, nil
	}

	u, err := url.Parse(apiURL + "/nodes")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("ids", strings.Join(nodeIDs, ","))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func findLayerTextsInFigmaJSON(figmaJSON map[string]any, layerName string, recursive bool) []figmadiff.LayerTextOutput {
	if document, ok := figmaJSON["document"]; ok {
		return figmadiff.FindTextByLayerName(document, layerName, recursive)
	}

	var matches []figmadiff.LayerTextOutput
	nodes, ok := figmaJSON["nodes"].(map[string]any)
	if !ok {
		return matches
	}
	for _, node := range nodes {
		nodeObject, ok := node.(map[string]any)
		if !ok {
			continue
		}
		document, ok := nodeObject["document"]
		if !ok {
			continue
		}
		matches = append(matches, figmadiff.FindTextByLayerName(document, layerName, recursive)...)
	}
	return matches
}

func init() {
	textsCmd.Flags().String("layer", "", "layer name to extract text from")
	textsCmd.Flags().String("id", "", "node ID to search within; accepts 20089:685897 or 20089-685897")
	textsCmd.Flags().Bool("recursive", false, "include text from all descendant nodes")
	rootCmd.AddCommand(textsCmd)
}
