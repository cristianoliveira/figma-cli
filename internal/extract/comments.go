package extract

import (
	"encoding/json"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

// CommentOutput is one comment, used by `figma comments`.
type CommentOutput struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	Resolved  bool   `json:"resolved"`
	NodeID    string `json:"node_id,omitempty"`
	User      string `json:"user"`
	ParentID  string `json:"parent_id,omitempty"`
	URL       string `json:"url,omitempty"`
}

// CommentNodeIDs returns IDs eligible for node-scoped comments.
func CommentNodeIDs(documents []any, recursive bool) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, document := range documents {
		collectCommentNodeIDs(document, recursive, ids)
	}
	return ids
}

func collectCommentNodeIDs(value any, recursive bool, ids map[string]struct{}) {
	node, ok := value.(map[string]any)
	if !ok {
		return
	}
	if id, ok := node["id"].(string); ok && id != "" {
		ids[id] = struct{}{}
	}
	if !recursive {
		return
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		collectCommentNodeIDs(child, true, ids)
	}
}

// FilterCommentsByNodeIDs keeps comments anchored to selected nodes and every
// reply in those comment threads while preserving API order.
func FilterCommentsByNodeIDs(comments []CommentOutput, nodeIDs map[string]struct{}) []CommentOutput {
	included := make(map[string]struct{})
	for _, comment := range comments {
		if _, ok := nodeIDs[comment.NodeID]; ok {
			included[comment.ID] = struct{}{}
		}
	}
	for changed := true; changed; {
		changed = false
		for _, comment := range comments {
			if _, ok := included[comment.ParentID]; !ok {
				continue
			}
			if _, ok := included[comment.ID]; ok {
				continue
			}
			included[comment.ID] = struct{}{}
			changed = true
		}
	}

	filtered := make([]CommentOutput, 0)
	for _, comment := range comments {
		if _, ok := included[comment.ID]; ok {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

// FilterCommentsByID returns the exact comment matching id.
func FilterCommentsByID(comments []CommentOutput, id string) []CommentOutput {
	for _, comment := range comments {
		if comment.ID == id {
			return []CommentOutput{comment}
		}
	}
	return nil
}

// FilterUnresolvedComments removes resolved comments while preserving order.
func FilterUnresolvedComments(comments []CommentOutput) []CommentOutput {
	filtered := make([]CommentOutput, 0, len(comments))
	for _, comment := range comments {
		if !comment.Resolved {
			filtered = append(filtered, comment)
		}
	}
	return filtered
}

// AncestorNodeIDs returns target and its ancestor path from a full document.
func AncestorNodeIDs(document any, targetID string) map[string]struct{} {
	var path []string
	if !findNodePath(document, targetID, &path) {
		return nil
	}
	ids := make(map[string]struct{}, len(path))
	for _, id := range path {
		ids[id] = struct{}{}
	}
	return ids
}

func findNodePath(value any, targetID string, path *[]string) bool {
	node, ok := value.(map[string]any)
	if !ok {
		return false
	}
	id := StringValue(node["id"])
	if id != "" {
		*path = append(*path, id)
	}
	if id == targetID {
		return true
	}
	children, _ := node["children"].([]any)
	for _, child := range children {
		if findNodePath(child, targetID, path) {
			return true
		}
	}
	if id != "" {
		*path = (*path)[:len(*path)-1]
	}
	return false
}

// ExtractNodeIDFromClientMeta extracts the node_id from a ClientMeta union.
// ClientMeta is a discriminated union of Vector | FrameOffset | FrameOffsetRegion.
// Vectors have no node_id; FrameOffsets and FrameOffsetRegions do.
func ExtractNodeIDFromClientMeta(cm api.Comment_ClientMeta) string {
	b, err := cm.MarshalJSON()
	if err != nil {
		return ""
	}
	var raw struct {
		NodeID string `json:"node_id"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return ""
	}
	return raw.NodeID
}
