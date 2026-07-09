package figma

import (
	"fmt"
	"net/url"
	"strings"
)

// FileInput contains parsed information from a file ID or Figma URL.
type FileInput struct {
	FileID  string
	NodeIDs []string // node IDs extracted from node-id query parameter(s)
}

// ParseInput extracts file ID and node IDs from either a file ID or a Figma URL.
// Node IDs are extracted from the node-id query parameter (multiple node IDs can be comma-separated).
// The node-id format uses hyphens (e.g., "339-27545") which are converted to colons for the API.
func ParseInput(input string) (*FileInput, error) {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		path := u.Path
		segments := strings.Split(strings.Trim(path, "/"), "/")
		var fileID string
		for i, seg := range segments {
			if seg == "design" || seg == "file" {
				if i+1 < len(segments) {
					fileID = segments[i+1]
					break
				}
				return nil, fmt.Errorf("URL missing file ID after /%s/", seg)
			}
		}
		if fileID == "" {
			return nil, fmt.Errorf("URL does not contain /design/ or /file/ path")
		}
		var nodeIDs []string
		for _, param := range u.Query()["node-id"] {
			parts := strings.Split(param, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}
				nodeID := strings.ReplaceAll(part, "-", ":")
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
		return &FileInput{FileID: fileID, NodeIDs: nodeIDs}, nil
	}
	return &FileInput{FileID: input}, nil
}
