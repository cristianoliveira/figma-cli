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

func TestExtractNodeIDFromClientMeta_VectorEmpty(t *testing.T) {
	// A plain Vector carries no node_id.
	var cm api.Comment_ClientMeta
	require.NoError(t, cm.FromVector(api.Vector{}))

	assert.Empty(t, ExtractNodeIDFromClientMeta(cm))
}
