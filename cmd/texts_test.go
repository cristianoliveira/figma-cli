package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextMatchesUsesSelectedURLNodeWithoutLayer(t *testing.T) {
	document := map[string]any{
		"id":   "4707:15504",
		"name": "Selected frame",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "4707:15505", "name": "Title", "type": "TEXT", "characters": "Adminless groups"},
		},
	}
	input, err := figma.ParseInput("https://www.figma.com/design/QAhpkgySSOJ6gwJUTB0glb/Adminless-groups?node-id=4707-15504&m=dev")
	require.NoError(t, err)

	result, err := textResult(document, input, "", false)

	require.NoError(t, err)
	texts, ok := result.([]extract.OrderedTextOutput)
	require.True(t, ok)
	require.Len(t, texts, 1)
	assert.Equal(t, "4707:15505", texts[0].ID)
	assert.Equal(t, "Adminless groups", texts[0].Text)
	assert.Equal(t, 1, texts[0].Depth)
	assert.Equal(t, "Selected frame", texts[0].ParentName)
}

func TestTextMatchesRequiresLayerWhenInputHasNoNodeID(t *testing.T) {
	input, err := figma.ParseInput("QAhpkgySSOJ6gwJUTB0glb")
	require.NoError(t, err)

	_, err = textResult(map[string]any{}, input, "", false)

	assert.EqualError(t, err, "--layer or a node ID is required")
}
