package assetsedge

import (
	"fmt"
	"io"
	"net/http"
)

// DownloadFile downloads fileURL via httpClient and writes the bytes to
// outputPath. It is the standalone download helper used by the
// `figma export` command (which downloads a single asset rather than a
// manifest).
func DownloadFile(httpClient *http.Client, outputPath string, fileURL string) error {
	resp, err := httpClient.Get(fileURL)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return classifyDownloadStatus(resp.StatusCode)
	}

	file, err := createFile(outputPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	return nil
}
