package figma

import (
	"github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// TextHistory fetches Figma version history and node text for the diff use case.
type TextHistory struct {
	client  *Client
	fileID  string
	nodeIDs []string
}

// NewTextHistory creates a Figma-backed text history for the requested nodes.
func NewTextHistory(client *Client, fileID string, nodeIDs []string) *TextHistory {
	return &TextHistory{client: client, fileID: fileID, nodeIDs: nodeIDs}
}

func (h *TextHistory) Versions() ([]diff.Version, error) {
	versions, err := FetchAllVersions(h.client, h.fileID)
	if err != nil {
		return nil, err
	}

	history := make([]diff.Version, len(versions))
	for index, version := range versions {
		history[index] = diff.Version{
			ID:        version.Id,
			CreatedAt: version.CreatedAt,
			User:      version.User.Handle,
		}
	}
	return history, nil
}

func (h *TextHistory) TextAt(versionID string) ([]extract.TextNode, error) {
	doc, err := FetchDocument(h.client, h.fileID, h.nodeIDs, versionID, "")
	if err != nil {
		return nil, err
	}
	return extract.ExtractTextNodes(doc), nil
}

// FetchAllVersions pages through a file's version history (newest-first) until
// the end. It uses the largest page size to minimise requests.
func FetchAllVersions(client *Client, fileID string) ([]api.Version, error) {
	var all []api.Version
	after := ""
	for {
		u, err := BuildVersionsURL(fileID, VersionsQuery{PageSize: 50, After: after})
		if err != nil {
			return nil, err
		}
		resp, err := FetchVersions(client, u)
		if err != nil {
			return nil, err
		}
		if len(resp.Versions) == 0 {
			break
		}
		all = append(all, resp.Versions...)
		if resp.Pagination.NextPage == nil {
			break
		}
		after = resp.Versions[len(resp.Versions)-1].Id
	}
	return all, nil
}
