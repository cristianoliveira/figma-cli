/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// extractFileID extracts a Figma file ID from either a file ID or a Figma URL.
// If the input looks like a URL, it parses it and extracts the file ID from the path.
// If the input is already a file ID, it returns it unchanged.
func extractFileID(input string) (string, error) {
	// Check if input looks like a URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		path := u.Path
		segments := strings.Split(strings.Trim(path, "/"), "/")
		// Look for "design" or "file" segment
		for i, seg := range segments {
			if seg == "design" || seg == "file" {
				if i+1 < len(segments) {
					return segments[i+1], nil
				} else {
					return "", fmt.Errorf("URL missing file ID after /%s/", seg)
				}
			}
		}
		return "", fmt.Errorf("URL does not contain /design/ or /file/ path")
	}
	// Not a URL, assume it's a file ID
	return input, nil
}

// fileInput contains parsed information from a file ID or Figma URL.
type fileInput struct {
	fileID  string
	nodeIDs []string // node IDs extracted from node-id query parameter(s)
}

// parseInput extracts file ID and node IDs from either a file ID or a Figma URL.
// Node IDs are extracted from the node-id query parameter (multiple node IDs can be comma-separated).
// The node-id format uses hyphens (e.g., "339-27545") which are converted to colons for the API.
func parseInput(input string) (*fileInput, error) {
	// Check if input looks like a URL
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		path := u.Path
		segments := strings.Split(strings.Trim(path, "/"), "/")
		var fileID string
		// Look for "design" or "file" segment
		for i, seg := range segments {
			if seg == "design" || seg == "file" {
				if i+1 < len(segments) {
					fileID = segments[i+1]
					break
				} else {
					return nil, fmt.Errorf("URL missing file ID after /%s/", seg)
				}
			}
		}
		if fileID == "" {
			return nil, fmt.Errorf("URL does not contain /design/ or /file/ path")
		}
		// Extract node IDs from query parameter "node-id"
		var nodeIDs []string
		for _, param := range u.Query()["node-id"] {
			// Split by commas in case multiple IDs in a single parameter
			parts := strings.Split(param, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				// Convert hyphens to colons
				nodeID := strings.ReplaceAll(part, "-", ":")
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
		return &fileInput{fileID: fileID, nodeIDs: nodeIDs}, nil
	}
	// Not a URL, assume it's a file ID
	return &fileInput{fileID: input}, nil
}

// fetchMetaCmd represents the fetch-meta command
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
		nodeID, _ := cmd.Flags().GetString("id")
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

		// Build API URL with query parameters
		apiURL := fmt.Sprintf("https://api.figma.com/v1/files/%s", input.fileID)
		u, err := url.Parse(apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing API URL: %v\n", err)
			os.Exit(1)
		}
		q := u.Query()
		q.Set("depth", "1")
		nodeIDs := resolveNodeIDs(input, nodeID)
		if len(nodeIDs) > 0 {
			q.Set("ids", strings.Join(nodeIDs, ","))
		}
		u.RawQuery = q.Encode()
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("X-Figma-Token", token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error making request: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "error: API returned status %d: %s\n", resp.StatusCode, body)
			os.Exit(1)
		}

		var result map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&result); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding JSON: %v\n", err)
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
	fetchMetaCmd.Flags().String("id", "", "node ID to fetch; accepts 20089:685897, 20089-685897, or comma-separated IDs")
	rootCmd.AddCommand(fetchMetaCmd)
}
