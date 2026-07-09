package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

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

		input, err := parseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		nodeIDs := resolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			fmt.Fprintln(os.Stderr, "error: export requires --id or a Figma URL with node-id")
			os.Exit(1)
		}
		if outputPath == "" {
			outputPath = defaultExportOutputPath(input.fileID, nodeIDs[0], format)
		}

		token, err := getFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		apiURL, err := buildExportAPIURL(input.fileID, []string{nodeIDs[0]}, format)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building export URL: %v\n", err)
			os.Exit(1)
		}
		assetURL, err := fetchExportURL(apiURL, nodeIDs[0], token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching export URL: %v\n", err)
			os.Exit(1)
		}
		if err := downloadFile(outputPath, assetURL); err != nil {
			fmt.Fprintf(os.Stderr, "error downloading export: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(outputPath)
	},
}

func buildExportAPIURL(fileID string, nodeIDs []string, format string) (string, error) {
	if len(nodeIDs) == 0 {
		return "", errors.New("node ID is required")
	}
	if format == "" {
		return "", errors.New("format is required")
	}

	u, err := url.Parse(fmt.Sprintf("https://api.figma.com/v1/images/%s", fileID))
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("format", format)
	q.Set("ids", strings.Join(nodeIDs, ","))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func defaultExportOutputPath(fileID string, nodeID string, format string) string {
	return fmt.Sprintf("%s_%s.%s", fileID, strings.ReplaceAll(nodeID, ":", "-"), format)
}

func fetchExportURL(apiURL string, nodeID string, token string) (string, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Figma-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	var result exportResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding JSON: %w", err)
	}
	assetURL := result.Images[nodeID]
	if assetURL == "" {
		return "", fmt.Errorf("no export URL returned for node %s", nodeID)
	}
	return assetURL, nil
}

func downloadFile(outputPath string, fileURL string) error {
	resp, err := http.Get(fileURL)
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
