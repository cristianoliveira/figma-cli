package figma

import "github.com/cristianoliveira/figma-cli/internal/figma/api"

// FetchDocument fetches a Figma file (optionally scoped to nodeIDs, version, depth)
// and returns the parsed document tree. It is the single way commands obtain a
// document, so the fetch + unmarshal sequence lives in one place.
func FetchDocument(client *Client, fileID string, nodeIDs []string, version, depth string) (any, error) {
	u, err := BuildFileURL(fileID, nodeIDs, version, depth)
	if err != nil {
		return nil, err
	}
	var resp api.GetFileResponse
	if err := client.Fetch(u, &resp); err != nil {
		return nil, err
	}
	return UnmarshalDocument(resp.Document)
}
