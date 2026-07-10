package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchVersions fetches the version history of a file via the Figma API.
func FetchVersions(client *Client, apiURL string) (api.GetFileVersionsResponse, error) {
	var result api.GetFileVersionsResponse
	if err := client.Fetch(apiURL, &result); err != nil {
		return api.GetFileVersionsResponse{}, fmt.Errorf("fetching versions: %w", err)
	}
	return result, nil
}
