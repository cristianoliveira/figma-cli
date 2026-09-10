package assets

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// AssetSource produces the ordered list of asset candidates the
// application will attempt. The implementation is responsible for
// fetching the document tree; the application does not know or care
// whether that is an HTTP call, a local cache, or a static fixture.
type AssetSource interface {
	// Candidates returns every node that may yield an exportable asset
	// (image fills, instances, vectors). The slice must be in the order
	// the caller will iterate; the application preserves that order.
	Candidates(ctx context.Context) ([]extract.Asset, error)
}

// ExportURLSource resolves an export URL for a single asset/format pair.
// Returning an error must be treated by the application as an
// item-level failure; later assets continue.
type ExportURLSource interface {
	ExportURL(ctx context.Context, nodeID, format string) (string, error)
}

// AssetSink delivers the bytes behind an export URL to a path on the
// chosen storage backend. Returning an error must be treated as an
// item-level failure; later assets continue.
type AssetSink interface {
	// Write performs a full download + persistence for one asset. The
	// path argument is the final destination path the application
	// computed (including any collision suffix and format extension).
	Write(ctx context.Context, path, url string) error
}
