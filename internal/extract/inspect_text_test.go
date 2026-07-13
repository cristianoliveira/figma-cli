package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatInspectTextSelectsFieldsAndPreservesDepth(t *testing.T) {
	nodes := []InspectOutput{
		{
			Depth:          0,
			Name:           "Conversation List",
			Type:           "FRAME",
			RelativeBounds: &relativeBoundsOutput{X: 0, Y: 0, Width: 320, Height: 800, RelativeTo: "16:6407"},
			Layout:         layoutOutput{Mode: "VERTICAL", Gap: 8, PaddingLeft: 16},
			Fills:          []string{"#FAFAFA"},
		},
		{
			Depth:          1,
			Name:           "Search",
			Type:           "TEXT",
			RelativeBounds: &relativeBoundsOutput{X: 16, Y: 12, Width: 84, Height: 0, RelativeTo: "16:6407"},
			Fills:          []string{"#000000"},
		},
	}

	text, err := FormatInspectText(nodes, []string{"name", "type", "relativeBounds", "layout.mode", "layout.gap", "layout.paddingLeft", "fills"})

	require.NoError(t, err)
	assert.Equal(t, "Conversation List (FRAME) x:0 y:0 w:320 h:800 layout:VERTICAL gap:8 layout.paddingLeft:16 fills:#FAFAFA\n  Search (TEXT) x:16 y:12 w:84 h:0 fills:#000000\n", text)
}

func TestFormatInspectTextRejectsUnknownFields(t *testing.T) {
	_, err := FormatInspectText(nil, []string{"name", "layout.padding"})

	assert.EqualError(t, err, `unknown inspect field "layout.padding"`)
}

func TestInspectTreeRecordsHierarchyDepthWithoutChangingJSON(t *testing.T) {
	nodes := InspectTree(map[string]any{
		"id": "1:1", "name": "Root", "type": "FRAME",
		"children": []any{map[string]any{"id": "1:2", "name": "Child", "type": "TEXT"}},
	})

	require.Len(t, nodes, 2)
	assert.Equal(t, 0, nodes[0].Depth)
	assert.Equal(t, 1, nodes[1].Depth)
}
