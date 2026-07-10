package cli

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// AssetExportItem records the outcome for one requested asset.
type AssetExportItem struct {
	NodeID string `json:"node_id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Format string `json:"format"`
	Path   string `json:"path,omitempty"`
	Error  string `json:"error,omitempty"`
}

// AssetExportManifest records every requested asset in request order.
type AssetExportManifest struct {
	Items     []AssetExportItem `json:"items"`
	Succeeded int               `json:"succeeded"`
	Failed    int               `json:"failed"`
}

// AssetExporter downloads assets using configured dependencies.
type AssetExporter struct {
	HTTPClient *http.Client
	FetchURL   func(nodeID, format string) (string, error)
	Filename   func(asset extract.Asset) string
}

// Export attempts every asset and returns a complete deterministic manifest.
func (e AssetExporter) Export(outputDirectory string, assets []extract.Asset) AssetExportManifest {
	manifest := AssetExportManifest{Items: make([]AssetExportItem, 0, len(assets))}
	usedPaths := make(map[string]int)
	client := e.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	filename := e.Filename
	if filename == nil {
		filename = defaultAssetFilename
	}

	for _, asset := range assets {
		item := AssetExportItem{NodeID: asset.ID, Name: asset.Name, Kind: asset.Kind, Format: asset.Format}
		assetURL, err := e.FetchURL(asset.ID, asset.Format)
		if err != nil {
			item.Error = err.Error()
			manifest.Failed++
			manifest.Items = append(manifest.Items, item)
			continue
		}

		basePath := filepath.Join(outputDirectory, filename(asset))
		usedPaths[basePath]++
		path := basePath
		if usedPaths[basePath] > 1 {
			path += "-" + strconv.Itoa(usedPaths[basePath])
		}
		path += "." + asset.Format
		item.Path = path
		if err := DownloadFile(client, path, assetURL); err != nil {
			item.Path = ""
			item.Error = err.Error()
			manifest.Failed++
		} else {
			manifest.Succeeded++
		}
		manifest.Items = append(manifest.Items, item)
	}
	return manifest
}

func defaultAssetFilename(asset extract.Asset) string {
	return asset.ID
}
