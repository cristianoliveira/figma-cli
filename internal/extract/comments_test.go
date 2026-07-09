package extract

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractNodeIDFromClientMeta_FrameOffset(t *testing.T) {
	var cm api.Comment_ClientMeta
	require.NoError(t, cm.FromFrameOffset(api.FrameOffset{NodeId: "1:2", NodeOffset: api.Vector{}}))

	assert.Equal(t, "1:2", ExtractNodeIDFromClientMeta(cm))
}

func TestCommentNodeIDs_RecursiveControlsDescendants(t *testing.T) {
	document := map[string]any{
		"id": "1:1",
		"children": []any{
			map[string]any{"id": "1:2"},
		},
	}

	assert.Equal(t, map[string]struct{}{"1:1": {}}, CommentNodeIDs([]any{document}, false))
	assert.Equal(t, map[string]struct{}{"1:1": {}, "1:2": {}}, CommentNodeIDs([]any{document}, true))
}

func TestFilterCommentsByNodeIDs_IncludesThreadReplies(t *testing.T) {
	comments := []CommentOutput{
		{ID: "reply", ParentID: "root"},
		{ID: "other", NodeID: "9:9"},
		{ID: "root", NodeID: "1:2"},
		{ID: "nested-reply", ParentID: "reply"},
	}

	filtered := FilterCommentsByNodeIDs(comments, map[string]struct{}{"1:2": {}})

	assert.Equal(t, []string{"reply", "root", "nested-reply"}, []string{filtered[0].ID, filtered[1].ID, filtered[2].ID})
}

func TestFilterCommentsByNodeIDs_NoMatches(t *testing.T) {
	filtered := FilterCommentsByNodeIDs([]CommentOutput{{ID: "other", NodeID: "9:9"}}, map[string]struct{}{"1:2": {}})

	assert.Empty(t, filtered)
}

func TestExtractNodeIDFromClientMeta_VectorEmpty(t *testing.T) {
	// A plain Vector carries no node_id.
	var cm api.Comment_ClientMeta
	require.NoError(t, cm.FromVector(api.Vector{}))

	assert.Empty(t, ExtractNodeIDFromClientMeta(cm))
}
