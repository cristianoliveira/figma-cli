package figma

import (
	"fmt"
	"net/url"
	"regexp"
)

var urlPathRegex = regexp.MustCompile(`^/(?:design|file)/([A-Za-z0-9]+)/([^/?]+)`)

// ParsedURL represents the extracted components from a Figma URL.
type ParsedURL struct {
	FileKey   string // Figma file key (alphanumeric)
	NodeID    string // Node identifier from node-id query parameter
	FileName  string // Name of the file (URL decoded)
	Version   string // Version identifier from version-id query parameter
	Timestamp string // Timestamp from t query parameter
}

// ParseURL parses a Figma URL and extracts file key, node ID, and other components.
// Supports both design and file URLs:
// - https://www.figma.com/design/{key}/{name}?node-id={node-id}
// - https://www.figma.com/file/{key}/{name}?node-id={node-id}
func ParseURL(rawURL string) (*ParsedURL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Extract file key from path
	matches := urlPathRegex.FindStringSubmatch(u.Path)
	if matches == nil {
		return nil, fmt.Errorf("URL path must be /design/{key}/{name} or /file/{key}/{name}, got %q", u.Path)
	}
	fileKey := matches[1]
	fileName := matches[2]

	// Extract node-id from query
	nodeID := u.Query().Get("node-id")
	// Extract version from query (if present)
	version := u.Query().Get("version-id")
	// Extract timestamp (t parameter)
	timestamp := u.Query().Get("t")

	return &ParsedURL{
		FileKey:   fileKey,
		NodeID:    nodeID,
		FileName:  fileName,
		Version:   version,
		Timestamp: timestamp,
	}, nil
}
