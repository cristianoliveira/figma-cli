package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchTeamProjects fetches all projects visible to the configured client
// and returns them as a stable TeamProjects DTO. The generated API response
// is mapped at the adapter boundary so callers do not depend on
// internal/figma/api.
func FetchTeamProjects(client *Client, apiURL string) (TeamProjects, error) {
	var raw api.GetTeamProjectsResponse
	if err := client.Fetch(apiURL, &raw); err != nil {
		return TeamProjects{}, fmt.Errorf("fetching team projects: %w", err)
	}
	return MapTeamProjects(raw), nil
}
