package comments

import (
	"encoding/json"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

func mapAPIComments(comments []api.Comment) []extract.CommentOutput {
	outputs := make([]extract.CommentOutput, 0, len(comments))
	for _, comment := range comments {
		output := extract.CommentOutput{
			ID:        comment.Id,
			Message:   comment.Message,
			CreatedAt: comment.CreatedAt.String(),
			Resolved:  comment.ResolvedAt != nil,
			User:      comment.User.Handle,
		}
		if comment.ParentId != nil {
			output.ParentID = *comment.ParentId
		}
		output.NodeID = nodeIDFromClientMeta(comment.ClientMeta)
		outputs = append(outputs, output)
	}
	return outputs
}

func nodeIDFromClientMeta(clientMeta api.Comment_ClientMeta) string {
	data, err := clientMeta.MarshalJSON()
	if err != nil {
		return ""
	}
	var raw struct {
		NodeID string `json:"node_id"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}
	return raw.NodeID
}
