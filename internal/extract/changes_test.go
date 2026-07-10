package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffDocumentsReportsStructuralFrontendChanges(t *testing.T) {
	from := map[string]any{"id": "0:0", "name": "Page", "type": "DOCUMENT", "children": []any{
		map[string]any{
			"id": "1:1", "name": "Old card", "type": "INSTANCE", "componentId": "C:old",
			"layoutMode": "HORIZONTAL", "itemSpacing": 8.0, "layoutWrap": "NO_WRAP",
			"paddingLeft": 8.0, "styles": map[string]any{"fill": "S:old"},
			"fills":               []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0}}},
			"absoluteBoundingBox": map[string]any{"x": 0.0, "y": 0.0, "width": 100.0, "height": 40.0},
		},
		map[string]any{"id": "1:2", "name": "Removed", "type": "FRAME"},
	}}
	to := map[string]any{"id": "0:0", "name": "Page", "type": "DOCUMENT", "children": []any{
		map[string]any{
			"id": "1:1", "name": "Card", "type": "INSTANCE", "componentId": "C:new",
			"layoutMode": "VERTICAL", "itemSpacing": 16.0, "layoutWrap": "WRAP",
			"paddingLeft": 12.0, "styles": map[string]any{"fill": "S:new"},
			"fills":               []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 0.0, "g": 0.0, "b": 1.0}}},
			"absoluteBoundingBox": map[string]any{"x": 4.0, "y": 0.0, "width": 120.0, "height": 40.0},
		},
		map[string]any{"id": "1:3", "name": "Added", "type": "FRAME"},
	}}

	changes := DiffDocuments(from, to)

	require.Len(t, changes, 3)
	assert.Equal(t, StructuralChange{ID: "1:3", Path: "Page/Added", Type: "added", NodeType: "FRAME"}, changes[0])
	assert.Equal(t, "1:1", changes[1].ID)
	assert.Equal(t, "Page/Card", changes[1].Path)
	assert.Equal(t, "modified", changes[1].Type)
	assert.Equal(t, []PropertyChange{
		{Property: "name", From: "Old card", To: "Card"},
		{Property: "componentId", From: "C:old", To: "C:new"},
		{Property: "layout.mode", From: "HORIZONTAL", To: "VERTICAL"},
		{Property: "layout.gap", From: float64(8), To: float64(16)},
		{Property: "layout.wrap", From: "NO_WRAP", To: "WRAP"},
		{Property: "layout.paddingLeft", From: float64(8), To: float64(12)},
		{Property: "fills", From: []string{"#FF0000"}, To: []string{"#0000FF"}},
		{Property: "styles.fill", From: "S:old", To: "S:new"},
		{Property: "bounds.x", From: float64(0), To: float64(4)},
		{Property: "bounds.width", From: float64(100), To: float64(120)},
	}, changes[1].Changes)
	assert.Equal(t, StructuralChange{ID: "1:2", Path: "Page/Removed", Type: "removed", NodeType: "FRAME"}, changes[2])
}

func TestDiffDocumentsReportsChildReordering(t *testing.T) {
	from := map[string]any{"id": "1", "name": "Frame", "type": "FRAME", "children": []any{
		map[string]any{"id": "2", "name": "A", "type": "RECTANGLE"},
		map[string]any{"id": "3", "name": "B", "type": "RECTANGLE"},
	}}
	to := map[string]any{"id": "1", "name": "Frame", "type": "FRAME", "children": []any{
		map[string]any{"id": "3", "name": "B", "type": "RECTANGLE"},
		map[string]any{"id": "2", "name": "A", "type": "RECTANGLE"},
	}}

	changes := DiffDocuments(from, to)

	require.Len(t, changes, 1)
	assert.Equal(t, PropertyChange{Property: "childrenOrder", From: []string{"2", "3"}, To: []string{"3", "2"}}, changes[0].Changes[0])
}

func TestDiffDocumentsReturnsEmptyForEquivalentTrees(t *testing.T) {
	document := map[string]any{"id": "1", "name": "Frame", "type": "FRAME"}

	assert.Empty(t, DiffDocuments(document, document))
}
