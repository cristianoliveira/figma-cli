package figma

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// FileInput contains parsed information from a file ID or Figma URL.
type FileInput struct {
	FileID    string
	NodeIDs   []string // node IDs extracted from node-id query parameter(s)
	CommentID string   // numeric comment ID extracted from URL fragment
}

// ParseInput extracts file ID and node IDs from either a file ID or a Figma URL.
// Node IDs are extracted from the node-id query parameter (multiple node IDs can be comma-separated).
// The node-id format uses hyphens (e.g., "339-27545") which are converted to colons for the API.
var numericID = regexp.MustCompile(`^[0-9]+$`)

// ParseTeamInput extracts a numeric team ID from a bare ID or Figma team URL.
func ParseTeamInput(input string) (string, error) {
	return parseDiscoveryInput(input, "team")
}

// ParseProjectInput extracts a numeric project ID from a bare ID or Figma project URL.
func ParseProjectInput(input string) (string, error) {
	return parseDiscoveryInput(input, "project")
}

func parseDiscoveryInput(input, kind string) (string, error) {
	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		if numericID.MatchString(input) {
			return input, nil
		}
		return "", fmt.Errorf("invalid %s ID %q: expected a numeric ID or Figma %s URL", kind, input, kind)
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid %s URL: %w", kind, err)
	}
	if u.Hostname() != "figma.com" && u.Hostname() != "www.figma.com" {
		return "", fmt.Errorf("invalid %s URL: expected figma.com host", kind)
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, segment := range segments {
		if segment != kind {
			continue
		}
		if i+1 < len(segments) && numericID.MatchString(segments[i+1]) {
			return segments[i+1], nil
		}
		return "", fmt.Errorf("invalid %s URL: missing numeric ID after /%s/", kind, kind)
	}
	return "", fmt.Errorf("invalid %s URL: path does not contain /%s/", kind, kind)
}

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
	return &FileInput{FileID: fileID, NodeIDs: extractNodeIDsFromQuery(u), CommentID: extractCommentID(u)}, nil
}

func extractCommentID(u *url.URL) string {
	if numericID.MatchString(u.Fragment) {
		return u.Fragment
	}
	return ""
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
