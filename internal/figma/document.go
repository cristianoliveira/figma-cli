package figma

import (
	"encoding/json"
	"fmt"
	"strings"

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

// FetchNodeDocuments fetches only the requested node subtrees, preserving the
// caller's node ID order. Use this when traversal must not include siblings.
func FetchNodeDocuments(client *Client, fileID string, nodeIDs []string) ([]any, error) {
	if len(nodeIDs) == 0 {
		return nil, fmt.Errorf("at least one node ID is required")
	}
	u, err := BuildNodesURL(fileID, nodeIDs)
	if err != nil {
		return nil, err
	}
	var resp api.GetFileNodesResponse
	if err := client.Fetch(u, &resp); err != nil {
		return nil, err
	}

	documents := make([]any, 0, len(nodeIDs))
	var fileDocument any
	for _, nodeID := range nodeIDs {
		if node, ok := resp.Nodes[nodeID]; ok {
			document, decodeErr := UnmarshalDocument(node.Document)
			if decodeErr == nil && documentNodeID(document) != "" {
				documents = append(documents, document)
				continue
			}
		}

		if fileDocument == nil {
			fileDocument, err = FetchDocument(client, fileID, nil, "", "")
			if err != nil {
				return nil, fmt.Errorf("fetching file to resolve node %s: %w", nodeID, err)
			}
		}
		resolved := findDocumentNode(fileDocument, nodeID)
		if resolved == nil {
			return nil, fmt.Errorf("node %s was not returned by Figma", nodeID)
		}
		documents = append(documents, resolved)
	}
	return documents, nil
}

func findDocumentNode(document any, nodeID string) any {
	if node := findNodeMatching(document, func(id string) bool { return id == nodeID }); node != nil {
		return node
	}
	return findNodeMatching(document, func(id string) bool {
		return strings.HasSuffix(id, ";"+nodeID)
	})
}

func findNodeMatching(value any, matches func(string) bool) any {
	node, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	if matches(documentNodeID(node)) {
		return node
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		if match := findNodeMatching(child, matches); match != nil {
			return match
		}
	}
	return nil
}

func documentNodeID(value any) string {
	node, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	id, _ := node["id"].(string)
	return id
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
