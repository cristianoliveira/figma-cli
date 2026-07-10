package comments

import (
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAPICommentsMapsNodeScopedReply(t *testing.T) {
	parentID := "root"
	var clientMeta api.Comment_ClientMeta
	require.NoError(t, clientMeta.FromFrameOffset(api.FrameOffset{NodeId: "1:2", NodeOffset: api.Vector{}}))

	outputs := mapAPIComments([]api.Comment{{
		Id:         "reply",
		Message:    "Fixed",
		CreatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		ParentId:   &parentID,
		ClientMeta: clientMeta,
		User:       api.User{Handle: "Ada"},
	}})

	assert.Equal(t, "reply", outputs[0].ID)
	assert.Equal(t, "root", outputs[0].ParentID)
	assert.Equal(t, "1:2", outputs[0].NodeID)
	assert.Equal(t, "Ada", outputs[0].User)
}

func TestMapAPICommentsLeavesNodeIDEmptyForUnanchoredComment(t *testing.T) {
	outputs := mapAPIComments([]api.Comment{{Id: "comment", ClientMeta: api.Comment_ClientMeta{}}})

	require.Len(t, outputs, 1)
	assert.Empty(t, outputs[0].NodeID)
}
