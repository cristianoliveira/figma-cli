package assets

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// DefaultExportOutputPath builds the default file path for an exported asset.
func DefaultExportOutputPath(fileID string, nodeID string, format string) string {
	return fmt.Sprintf("%s_%s.%s", fileID, strings.ReplaceAll(nodeID, ":", "-"), format)
}

// DownloadFile downloads fileURL via httpClient and writes the bytes to outputPath.
func DownloadFile(httpClient *http.Client, outputPath string, fileURL string) error {
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
