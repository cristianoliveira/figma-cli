package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchComments fetches the comments of a file via the Figma API and returns
// them as stable Comment DTOs. The generated API response is mapped at the
// adapter boundary so callers do not depend on internal/figma/api.
func FetchComments(client *Client, apiURL string) ([]Comment, error) {
	var raw api.GetCommentsResponse
	if err := client.Fetch(apiURL, &raw); err != nil {
		return nil, fmt.Errorf("fetching comments: %w", err)
	}
	return MapComments(raw.Comments), nil
}
