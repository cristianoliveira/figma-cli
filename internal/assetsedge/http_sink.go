package assetsedge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cristianoliveira/figma-cli/internal/artifact"
)

// HTTPAssetSink implements assets.AssetSink against an *http.Client and
// the filesystem. It is the edge adapter; the asset application knows
// nothing about HTTP or filesystem mechanics.
type HTTPAssetSink struct {
	Client *http.Client
}

// NewHTTPAssetSink builds an HTTP/filesystem AssetSink.
func NewHTTPAssetSink(client *http.Client) *HTTPAssetSink {
	return &HTTPAssetSink{Client: client}
}

// Write performs a full download + persistence for one asset. The path
// argument is the final destination path (already including any
// collision suffix and format extension).
func (s *HTTPAssetSink) Write(ctx context.Context, path, url string) error {
	if s.Client == nil {
		return errors.New("HTTPAssetSink: HTTP client is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return classifyDownloadStatus(resp.StatusCode)
	}

	file, err := createFile(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	return nil
}

// createFile creates the destination file, creating parent directories
// as needed. Filesystem failures are translated to a neutral
// artifact-access error.
func createFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, artifact.ClassifyFileError(err)
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, artifact.ClassifyFileError(err)
	}
	return file, nil
}
