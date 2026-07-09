package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchTeamProjects fetches all projects visible to the configured client.
func FetchTeamProjects(client *Client, apiURL string) (api.GetTeamProjectsResponse, error) {
	var result api.GetTeamProjectsResponse
	if err := client.Fetch(apiURL, &result); err != nil {
		return api.GetTeamProjectsResponse{}, fmt.Errorf("fetching team projects: %w", err)
	}
	return result, nil
}
