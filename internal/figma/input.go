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
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		return &FileInput{FileID: input}, nil
	}
	u, err := url.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	fileID, err := extractFileIDFromURL(u)
	if err != nil {
		return nil, err
	}
	return &FileInput{FileID: fileID, NodeIDs: extractNodeIDsFromQuery(u)}, nil
}

// extractFileIDFromURL extracts the file key from a Figma URL path.
func extractFileIDFromURL(u *url.URL) (string, error) {
	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, seg := range segments {
		if seg == "design" || seg == "file" {
			if i+1 < len(segments) {
				return segments[i+1], nil
			}
			return "", fmt.Errorf("URL missing file ID after /%s/", seg)
		}
	}
	return "", fmt.Errorf("URL does not contain /design/ or /file/ path")
}

// extractNodeIDsFromQuery extracts node IDs from the URL query (node-id params).
func extractNodeIDsFromQuery(u *url.URL) []string {
	var nodeIDs []string
	for _, param := range u.Query()["node-id"] {
		for _, part := range strings.Split(param, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			nodeIDs = append(nodeIDs, strings.ReplaceAll(part, "-", ":"))
		}
	}
	return nodeIDs
}
