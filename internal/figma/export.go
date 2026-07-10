package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// ValidateExportFormat checks whether format is supported by Figma's image export API.
func ValidateExportFormat(format string) error {
	switch format {
	case "png", "jpg", "svg", "pdf":
		return nil
	default:
		return fmt.Errorf("invalid format %q: expected png, jpg, svg, or pdf", format)
	}
}

// FetchExportURL requests an export asset URL for nodeID from the Figma API.
func FetchExportURL(client *Client, apiURL string, nodeID string) (string, error) {
	var result api.GetImagesResponse
	if err := client.Fetch(apiURL, &result); err != nil {
		return "", fmt.Errorf("fetching export URL: %w", err)
	}
	assetURL, ok := result.Images[nodeID]
	if !ok || assetURL == nil {
		return "", fmt.Errorf("no export URL returned for node %s", nodeID)
	}
	return *assetURL, nil
}
