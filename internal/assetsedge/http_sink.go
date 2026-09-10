package assetsedge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/output"
)

// HTTPAssetSink implements assets.AssetSink against an *http.Client and
// the filesystem (via output.CreateFile, which also creates parent
// directories). It is the edge adapter; the asset application knows
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download returned status %d: %s", resp.StatusCode, body)
	}

	file, err := output.CreateFile(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}
	return nil
}
