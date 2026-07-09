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
		apiURL, err := buildFindAPIURL(input.fileID, resolveNodeIDs(input, nodeID))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building find URL: %v\n", err)
			os.Exit(1)
		}
		figmaJSON, err := figmadiff.FetchFigmaJSON(apiURL, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching Figma file: %v\n", err)
			os.Exit(1)
		}

		matches := findLayersInFigmaJSON(figmaJSON, layerName)
		output, err := json.MarshalIndent(map[string]any{"name": layerName, "matches": matches}, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func buildFindAPIURL(fileID string, nodeIDs []string) (string, error) {
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

func findLayersInFigmaJSON(figmaJSON map[string]any, layerName string) []layerMatchOutput {
	if document, ok := figmaJSON["document"]; ok {
		return findLayersByName(document, layerName)
	}

	var matches []layerMatchOutput
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
		matches = append(matches, findLayersByName(document, layerName)...)
	}
	return matches
}

func layerStringValue(value any) string {
	text, _ := value.(string)
	return text
}

func findLayersByName(value any, layerName string) []layerMatchOutput {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	var matches []layerMatchOutput
	if object["name"] == layerName {
		matches = append(matches, layerMatchOutput{
			ID:   layerStringValue(object["id"]),
			Name: layerStringValue(object["name"]),
			Type: layerStringValue(object["type"]),
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
