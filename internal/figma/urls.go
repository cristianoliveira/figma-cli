package figma

import (
	"fmt"
	"net/url"
	"strconv"
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

// BuildFileComponentsURL builds the API URL for published components in a Figma file.
func BuildFileComponentsURL(fileID string) string {
	return fmt.Sprintf("%s/files/%s/components", baseURL, fileID)
}

// BuildMeURL builds the API URL for the currently authenticated user (/v1/me).
func BuildMeURL() string {
	return baseURL + "/me"
}

// BuildTeamProjectsURL builds the API URL for listing a team's projects.
func BuildTeamProjectsURL(teamID string) string {
	return fmt.Sprintf("%s/teams/%s/projects", baseURL, teamID)
}

// BuildProjectFilesURL builds the API URL for listing a project's files.
func BuildProjectFilesURL(projectID string, branchData bool) string {
	raw := fmt.Sprintf("%s/projects/%s/files", baseURL, projectID)
	if !branchData {
		return raw
	}
	return raw + "?branch_data=true"
}

// VersionsQuery carries the Figma versions endpoint pagination params.
// See GET /v1/files/{file_key}/versions in openapi/openapi.yaml.
type VersionsQuery struct {
	// PageSize caps the number of versions returned (1..50). Zero means
	// unset, leaving the API default of 30.
	PageSize int
	// Before is a version ID: fetch versions newer than this one.
	Before string
	// After is a version ID: fetch versions older than this one.
	After string
}

// BuildVersionsURL builds the API URL for fetching version history of a Figma
// file, applying optional pagination params. Before and After are mutually
// exclusive since the endpoint paginates one direction at a time.
func BuildVersionsURL(fileID string, query VersionsQuery) (string, error) {
	if fileID == "" {
		return "", fmt.Errorf("file ID is required")
	}
	u, err := url.Parse(fmt.Sprintf("%s/files/%s/versions", baseURL, fileID))
	if err != nil {
		return "", err
	}
	if query.Before != "" && query.After != "" {
		return "", fmt.Errorf("--before and --after are mutually exclusive")
	}
	q := u.Query()
	if query.PageSize != 0 {
		if query.PageSize < 1 || query.PageSize > 50 {
			return "", fmt.Errorf("page-size must be between 1 and 50, got %d", query.PageSize)
		}
		q.Set("page_size", strconv.Itoa(query.PageSize))
	}
	if query.Before != "" {
		q.Set("before", query.Before)
	}
	if query.After != "" {
		q.Set("after", query.After)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
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

// BuildCommentWebURL builds a direct Figma URL for a comment.
func BuildCommentWebURL(fileID, nodeID, commentID string) string {
	u := &url.URL{Scheme: "https", Host: "www.figma.com", Path: "/design/" + fileID, Fragment: commentID}
	query := u.Query()
	query.Set("m", "dev")
	if nodeID != "" {
		query.Set("node-id", strings.ReplaceAll(nodeID, ":", "-"))
	}
	u.RawQuery = query.Encode()
	return u.String()
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
