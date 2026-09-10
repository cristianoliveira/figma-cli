package comments

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAPICommentsMapsNodeScopedReply(t *testing.T) {
	parentID := "root"
	meta := json.RawMessage(`{"node_id":"1:2"}`)

	outputs := mapAPIComments([]figma.Comment{{
		ID:         "reply",
		Message:    "Fixed",
		CreatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		ParentID:   &parentID,
		ClientMeta: meta,
		User:       figma.User{Handle: "Ada", ID: "1"},
	}})

	assert.Equal(t, "reply", outputs[0].ID)
	assert.Equal(t, "root", outputs[0].ParentID)
	assert.Equal(t, "1:2", outputs[0].NodeID)
	assert.Equal(t, "Ada", outputs[0].User)
}

func TestMapAPICommentsLeavesNodeIDEmptyForUnanchoredComment(t *testing.T) {
	outputs := mapAPIComments([]figma.Comment{{ID: "comment"}})

	require.Len(t, outputs, 1)
	assert.Empty(t, outputs[0].NodeID)
}

func TestMapAPICommentsFallsBackWhenClientMetaIsMalformed(t *testing.T) {
	outputs := mapAPIComments([]figma.Comment{{ID: "comment", ClientMeta: json.RawMessage(`not json`)}})

	require.Len(t, outputs, 1)
	assert.Empty(t, outputs[0].NodeID, "malformed client meta yields empty node id")
}
