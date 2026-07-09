package figma

import (
	"encoding/json"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

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

// UnmarshalDocument converts a typed document node (from the generated API types)
// into a generic tree for use with the extractor functions in internal/extract.
func UnmarshalDocument(node any) (any, error) {
	data, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}
