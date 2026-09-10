package assetsedge

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// FigmaExportURLSource implements assets.ExportURLSource against the
// Figma client. It preserves the existing file/node/format/scale
// semantics.
type FigmaExportURLSource struct {
	Client *figma.Client
	FileID string
}

// NewFigmaExportURLSource builds a Figma-backed ExportURLSource.
func NewFigmaExportURLSource(client *figma.Client, fileID string) *FigmaExportURLSource {
	return &FigmaExportURLSource{Client: client, FileID: fileID}
}

// ExportURL resolves the export URL for one node/format pair using the
// default scale of 1. Scale is not part of the application contract
// because the existing command never changes it.
func (s *FigmaExportURLSource) ExportURL(ctx context.Context, nodeID, format string) (string, error) {
	apiURL, err := figma.BuildExportURL(s.FileID, []string{nodeID}, format, 1)
	if err != nil {
		return "", err
	}
	return figma.FetchExportURL(s.Client.WithContext(ctx), apiURL, nodeID)
}
