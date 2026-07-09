package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchMe calls /v1/me for the currently authenticated user. apiURL is passed
// in (rather than built here) so tests can point it at an httptest server,
// matching the FetchExportURL seam.
func FetchMe(client *Client, apiURL string) (api.GetMeResponse, error) {
	var result api.GetMeResponse
	if err := client.Fetch(apiURL, &result); err != nil {
		return api.GetMeResponse{}, fmt.Errorf("fetching current user: %w", err)
	}
	return result, nil
}
