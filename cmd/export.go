package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

type exportResponse struct {
	Images map[string]string `json:"images"`
}

var exportCmd = &cobra.Command{
	Use:   "export [figma-url-with-node-id]",
	Short: "Export a Figma node asset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		outputPath, _ := cmd.Flags().GetString("output")
		nodeID, _ := cmd.Flags().GetString("id")

		input, err := figma.ParseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			fmt.Fprintln(os.Stderr, "error: export requires --id or a Figma URL with node-id")
			os.Exit(1)
		}
		if outputPath == "" {
			outputPath = defaultExportOutputPath(input.FileID, nodeIDs[0], format)
		}

		token, err := env.GetFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		apiURL, err := figma.BuildExportURL(input.FileID, []string{nodeIDs[0]}, format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building export URL: %v\n", err)
			os.Exit(1)
		}
		client := figma.NewClient(token)
		assetURL, err := fetchExportURL(client, apiURL, nodeIDs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching export URL: %v\n", err)
			os.Exit(1)
		}
		if err := downloadFile(http.DefaultClient, outputPath, assetURL); err != nil {
			fmt.Fprintf(os.Stderr, "error downloading export: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(outputPath)
	},
}

func defaultExportOutputPath(fileID string, nodeID string, format string) string {
	return fmt.Sprintf("%s_%s.%s", fileID, strings.ReplaceAll(nodeID, ":", "-"), format)
}

func fetchExportURL(client *figma.Client, apiURL string, nodeID string) (string, error) {
	var result exportResponse
	if err := client.Fetch(apiURL, &result); err != nil {
		return "", fmt.Errorf("fetching export URL: %w", err)
	}
	assetURL := result.Images[nodeID]
	if assetURL == "" {
		return "", fmt.Errorf("no export URL returned for node %s", nodeID)
	}
	return assetURL, nil
}

func downloadFile(httpClient *http.Client, outputPath string, fileURL string) error {
	resp, err := httpClient.Get(fileURL)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download returned status %d: %s", resp.StatusCode, body)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	return nil
}

func init() {
	exportCmd.Flags().String("format", "png", "export format: png, jpg, svg, or pdf")
	exportCmd.Flags().String("id", "", "node ID to export; accepts 20089:685897 or 20089-685897")
	exportCmd.Flags().StringP("output", "o", "", "output file path; defaults to <file-key>_<node-id>.<format>")
	rootCmd.AddCommand(exportCmd)
}
