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
