package assets

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
