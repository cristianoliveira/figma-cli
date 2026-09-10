package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchProjectFiles fetches all files visible to the configured client and
// returns them as a stable ProjectFiles DTO. The generated API response is
// mapped at the adapter boundary so callers do not depend on
// internal/figma/api.
func FetchProjectFiles(client *Client, apiURL string) (ProjectFiles, error) {
	var raw api.GetProjectFilesResponse
	if err := client.Fetch(apiURL, &raw); err != nil {
		return ProjectFiles{}, fmt.Errorf("fetching project files: %w", err)
	}
	return MapProjectFiles(raw), nil
}
