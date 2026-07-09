package figma

import (
	"fmt"
	"net/url"
	"strings"
)

var baseURL = "https://api.figma.com/v1"

// BuildFileURL builds the API URL for fetching a Figma file.
func BuildFileURL(fileID string, nodeIDs []string, versionID string, depth string) (string, error) {
	raw := fmt.Sprintf("%s/files/%s", baseURL, fileID)
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q := u.Query()
	if versionID != "" {
		q.Set("version", versionID)
	}
	if depth != "" {
		q.Set("depth", depth)
	}
	if len(nodeIDs) > 0 {
		q.Set("ids", strings.Join(nodeIDs, ","))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// BuildNodesURL builds the API URL for fetching specific nodes from a Figma file.
func BuildNodesURL(fileID string, nodeIDs []string) (string, error) {
	raw := fmt.Sprintf("%s/files/%s/nodes", baseURL, fileID)
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q := u.Query()
	if len(nodeIDs) > 0 {
		q.Set("ids", strings.Join(nodeIDs, ","))
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// BuildVersionsURL builds the API URL for fetching version history of a Figma file.
func BuildVersionsURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/versions", baseURL, fileID)
}

// BuildStylesURL builds the API URL for listing the published styles of a file.
func BuildStylesURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/styles", baseURL, fileID)
}

// BuildVariablesURL builds the API URL for the local variables of a file
// (Figma's design-token system; requires Enterprise org access).
func BuildVariablesURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/variables/local", baseURL, fileID)
}

// BuildCommentsURL builds the API URL for fetching comments of a Figma file.
// If nodeID is non-empty, filters comments to that node.
func BuildCommentsURL(fileID string, nodeID string) string {
	raw := fmt.Sprintf("%s/files/%s/comments", baseURL, fileID)
	if nodeID == "" {
		return raw
	}
	return fmt.Sprintf("%s?node_id=%s", raw, nodeID)
}

// BuildExportURL builds the API URL for exporting a node image.
func BuildExportURL(fileID string, nodeIDs []string, format string) (string, error) {
	if len(nodeIDs) == 0 {
		return "", fmt.Errorf("node ID is required")
	}
	if format == "" {
		return "", fmt.Errorf("format is required")
	}
	raw := fmt.Sprintf("%s/images/%s", baseURL, fileID)
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("format", format)
	q.Set("ids", strings.Join(nodeIDs, ","))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
