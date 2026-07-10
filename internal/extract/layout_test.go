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

func TestExtractLayoutMeasuresAdjacentVerticalSpacingWhenRequested(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Content", "type": "FRAME", "layoutMode": "VERTICAL", "itemSpacing": float64(16),
		"children": []any{
			map[string]any{"id": "1:2", "name": "Title", "type": "TEXT", "characters": "Title", "absoluteBoundingBox": map[string]any{"x": float64(0), "y": float64(10), "width": float64(100), "height": float64(20)}},
			map[string]any{"id": "1:3", "name": "Body", "type": "TEXT", "characters": "Body", "absoluteBoundingBox": map[string]any{"x": float64(0), "y": float64(46), "width": float64(100), "height": float64(20)}},
		},
	}

	withoutMeasurement := ExtractLayout(document)
	withMeasurement := ExtractLayout(document, LayoutOptions{MeasureSpacing: true})

	assert.Nil(t, withoutMeasurement.Children[1].SpacingFromPrevious)
	assert.Equal(t, &LayoutSpacing{
		PreviousID: "1:2", Axis: "vertical", Measured: 16, Declared: numberPointer(16), MatchesDeclared: true,
	}, withMeasurement.Children[1].SpacingFromPrevious)
}

func TestExtractLayoutOmitsSpacingForAbsoluteOrNonAdjacentMeaningfulChildren(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Content", "type": "FRAME", "layoutMode": "VERTICAL",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Title", "type": "TEXT", "characters": "Title", "absoluteBoundingBox": map[string]any{"x": float64(0), "y": float64(0), "width": float64(10), "height": float64(10)}},
			map[string]any{"id": "1:9", "name": "Decoration", "type": "RECTANGLE"},
			map[string]any{"id": "1:3", "name": "Body", "type": "TEXT", "characters": "Body", "layoutPositioning": "ABSOLUTE", "absoluteBoundingBox": map[string]any{"x": float64(0), "y": float64(20), "width": float64(10), "height": float64(10)}},
		},
	}

	got := ExtractLayout(document, LayoutOptions{MeasureSpacing: true})

	assert.Nil(t, got.Children[1].SpacingFromPrevious)
}

func TestExtractLayoutReturnsEmptyForInvalidDocument(t *testing.T) {
	assert.Equal(t, LayoutNode{}, ExtractLayout(nil))
}

func numberPointer(value float64) *float64 {
	return &value
}
