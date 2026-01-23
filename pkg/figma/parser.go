package figma

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
)

// ParsedURL represents the extracted components from a Figma URL.
type ParsedURL struct {
	FileKey   string
	NodeID    string
	FileName  string
	Version   string
	Timestamp string
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
	re := regexp.MustCompile(`^/(?:design|file)/([A-Za-z0-9]+)/([^/?]+)`)
	matches := re.FindStringSubmatch(u.Path)
	if matches == nil {
		return nil, errors.New("URL path must be /design/{key}/{name} or /file/{key}/{name}")
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
