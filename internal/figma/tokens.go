package figma

import (
	"encoding/json"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// FetchVariables returns the `meta` object from GET /files/{key}/variables/local.
// This is Figma's design-token system; it requires Enterprise org access, so
// callers should surface a clear error when the API rejects the request.
func FetchVariables(client *Client, fileID string) (map[string]any, error) {
	var resp api.GetLocalVariablesResponse
	if err := client.Fetch(BuildVariablesURL(fileID), &resp); err != nil {
		return nil, err
	}
	return toMap(resp.Meta)
}

// FetchStyles returns the published style metadata list from
// GET /files/{key}/styles. Each entry carries a node_id used to resolve values.
func FetchStyles(client *Client, fileID string) ([]map[string]any, error) {
	var resp api.GetFileStylesResponse
	if err := client.Fetch(BuildStylesURL(fileID), &resp); err != nil {
		return nil, err
	}
	styles := make([]map[string]any, 0, len(resp.Meta.Styles))
	for _, s := range resp.Meta.Styles {
		m, err := toMap(s)
		if err != nil {
			return nil, err
		}
		styles = append(styles, m)
	}
	return styles, nil
}

// FetchNodes returns node documents keyed by node id from
// GET /files/{key}/nodes?ids=... Each value is the raw node entry (which
// contains a "document" key holding the node tree).
func FetchNodes(client *Client, fileID string, nodeIDs []string) (map[string]any, error) {
	u, err := BuildNodesURL(fileID, nodeIDs)
	if err != nil {
		return nil, err
	}
	var resp api.GetFileNodesResponse
	if err := client.Fetch(u, &resp); err != nil {
		return nil, err
	}
	return toMap(resp.Nodes)
}

// toMap converts any value into a generic map via JSON round-trip, mirroring how
// FetchDocument turns typed API responses into trees for the extract package.
func toMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
