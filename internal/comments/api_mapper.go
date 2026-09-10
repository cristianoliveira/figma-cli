package comments

import (
	"encoding/json"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// mapAPIComments converts the stable figma.Comment DTOs returned by the
// Figma adapter into the command-layer extract.CommentOutput. This is the
// single place where generated-API shape meets command output shape; the
// raw API types are kept inside the figma package.
func mapAPIComments(comments []figma.Comment) []extract.CommentOutput {
	outputs := make([]extract.CommentOutput, 0, len(comments))
	for _, comment := range comments {
		output := extract.CommentOutput{
			ID:        comment.ID,
			Message:   comment.Message,
			CreatedAt: comment.CreatedAt.String(),
			Resolved:  comment.Resolved,
			User:      comment.User.Handle,
		}
		if comment.ParentID != nil {
			output.ParentID = *comment.ParentID
		}
		output.NodeID = nodeIDFromClientMeta(comment.ClientMeta)
		outputs = append(outputs, output)
	}
	return outputs
}

// nodeIDFromClientMeta extracts the node_id field from the raw client_meta
// JSON payload. Returns empty when the payload is missing or malformed.
func nodeIDFromClientMeta(clientMeta json.RawMessage) string {
	if len(clientMeta) == 0 {
		return ""
	}
	var raw struct {
		NodeID string `json:"node_id"`
	}
	if err := json.Unmarshal(clientMeta, &raw); err != nil {
		return ""
	}
	return raw.NodeID
}
