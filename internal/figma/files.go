package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchProjectFiles fetches all files visible to the configured client.
func FetchProjectFiles(client *Client, apiURL string) (api.GetProjectFilesResponse, error) {
	var result api.GetProjectFilesResponse
	if err := client.Fetch(apiURL, &result); err != nil {
		return api.GetProjectFilesResponse{}, fmt.Errorf("fetching project files: %w", err)
	}
	return result, nil
}
