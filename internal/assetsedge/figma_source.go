// Package assetsedge holds the infrastructure adapters that implement
// the asset application ports defined in internal/assets. The
// application package owns the workflow and policy; this package owns
// Figma transport, HTTP downloads, and filesystem writes.
package assetsedge

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// FigmaAssetSource implements assets.AssetSource against the production
// Figma client.
type FigmaAssetSource struct {
	Client  *figma.Client
	FileID  string
	NodeIDs []string
}

var _ assets.AssetSource = (*FigmaAssetSource)(nil)

// NewFigmaAssetSource builds a Figma-backed AssetSource.
func NewFigmaAssetSource(client *figma.Client, fileID string, nodeIDs []string) *FigmaAssetSource {
	return &FigmaAssetSource{Client: client, FileID: fileID, NodeIDs: nodeIDs}
}

// Candidates fetches the requested node documents and runs the existing
// pure extraction. The application is responsible for filtering and
// naming policy; this adapter only supplies the raw candidate stream.
func (s *FigmaAssetSource) Candidates(ctx context.Context) ([]extract.Asset, error) {
	documents, err := figma.FetchNodeDocuments(s.Client.WithContext(ctx), s.FileID, s.NodeIDs)
	if err != nil {
		return nil, err
	}
	return extract.ExtractAssets(documents), nil
}
