package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLayoutPreservesTreeOrderAndCopy(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Dialog", "type": "FRAME",
		"layoutMode": "VERTICAL", "itemSpacing": float64(16),
		"paddingTop": float64(24), "paddingRight": float64(20), "paddingBottom": float64(24), "paddingLeft": float64(20),
		"children": []any{
			map[string]any{"id": "1:9", "name": "Decoration", "type": "RECTANGLE"},
			map[string]any{"id": "1:8", "name": "Icon", "type": "INSTANCE", "children": []any{
				map[string]any{"id": "1:7", "name": "Path", "type": "VECTOR"},
			}},
			map[string]any{"id": "1:2", "name": "Old title", "type": "TEXT", "characters": "Actual title"},
			map[string]any{
				"id": "1:3", "name": "Actions", "type": "FRAME", "layoutMode": "HORIZONTAL", "itemSpacing": float64(8),
				"children": []any{
					map[string]any{"id": "1:4", "name": "Label", "type": "TEXT", "characters": "Continue"},
				},
			},
		},
	}

	got := ExtractLayout(document)

	assert.Equal(t, LayoutNode{
		ID: "1:1", Name: "Dialog", Type: "FRAME", LayoutMode: "VERTICAL", Gap: numberPointer(16),
		Padding: &LayoutPadding{Top: 24, Right: 20, Bottom: 24, Left: 20},
		Children: []LayoutNode{
			{ID: "1:2", Name: "Old title", Type: "TEXT", Text: "Actual title"},
			{ID: "1:3", Name: "Actions", Type: "FRAME", LayoutMode: "HORIZONTAL", Gap: numberPointer(8), Children: []LayoutNode{
				{ID: "1:4", Name: "Label", Type: "TEXT", Text: "Continue"},
			}},
		},
	}, got)
}

func TestExtractLayoutReturnsEmptyForInvalidDocument(t *testing.T) {
	assert.Equal(t, LayoutNode{}, ExtractLayout(nil))
}

func numberPointer(value float64) *float64 {
	return &value
}
