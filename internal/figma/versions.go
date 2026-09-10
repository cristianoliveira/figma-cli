package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchVersions fetches the version history of a file via the Figma API and
// returns the result as a stable FileVersions DTO. The generated API
// response is mapped at the adapter boundary so callers do not depend on
// internal/figma/api.
func FetchVersions(client *Client, apiURL string) (FileVersions, error) {
	var raw api.GetFileVersionsResponse
	if err := client.Fetch(apiURL, &raw); err != nil {
		return FileVersions{}, fmt.Errorf("fetching versions: %w", err)
	}
	return MapFileVersions(raw), nil
}
