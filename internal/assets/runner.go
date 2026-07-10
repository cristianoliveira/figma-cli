package assets

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

const (
	assetFormatAuto = "auto"
	assetKindAll    = "all"
	assetKindIcon   = "icon"
	assetKindVector = "vector"
)

// AssetExportRequest describes an asset discovery and download operation.
type AssetExportRequest struct {
	Client          *figma.Client
	FileID          string
	NodeIDs         []string
	OutputDirectory string
	Kind            string
	Format          string
	NameFilter      string
	DownloadClient  *http.Client
	Filename        func(extract.Asset) string
}

// ExportAssets fetches selected node trees, filters their exportable assets, and
// downloads every matching asset into OutputDirectory.
func ExportAssets(request AssetExportRequest) (AssetExportManifest, error) {
	if request.Client == nil {
		return AssetExportManifest{}, fmt.Errorf("figma client is required")
	}

	documents, err := figma.FetchNodeDocuments(request.Client, request.FileID, request.NodeIDs)
	if err != nil {
		return AssetExportManifest{}, err
	}
	if err := os.MkdirAll(request.OutputDirectory, 0o755); err != nil {
		return AssetExportManifest{}, err
	}

	downloadClient := request.DownloadClient
	if downloadClient == nil {
		downloadClient = request.Client.HTTP
	}
	exporter := AssetExporter{
		HTTPClient: downloadClient,
		Filename:   request.Filename,
		FetchURL: func(nodeID, format string) (string, error) {
			apiURL, err := figma.BuildExportURL(request.FileID, []string{nodeID}, format)
			if err != nil {
				return "", err
			}
			return figma.FetchExportURL(request.Client, apiURL, nodeID)
		},
	}
	return exporter.Export(request.OutputDirectory, filterAssets(extract.ExtractAssets(documents), request.Kind, request.Format, request.NameFilter)), nil
}

func filterAssets(assets []extract.Asset, kind, format, nameFilter string) []extract.Asset {
	filtered := make([]extract.Asset, 0, len(assets))
	for _, asset := range assets {
		if kind == assetKindIcon && asset.Kind != "instance" && asset.Kind != assetKindVector {
			continue
		}
		if kind != assetKindAll && kind != assetKindIcon && asset.Kind != kind {
			continue
		}
		if nameFilter != "" && !strings.Contains(strings.ToLower(asset.Name), strings.ToLower(nameFilter)) {
			continue
		}
		if format != assetFormatAuto {
			asset.Format = format
		}
		filtered = append(filtered, asset)
	}
	return filtered
}
