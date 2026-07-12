package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractCoordinateAnnotationsUsesScopeRelativePixelCoverage(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Card", "type": "FRAME",
		"absoluteBoundingBox": map[string]any{"x": 100.25, "y": 200.5, "width": 50.2, "height": 40.1},
		"children": []any{map[string]any{
			"id": "1:2", "name": "Label", "type": "TEXT",
			"absoluteBoundingBox": map[string]any{"x": 112.5, "y": 205.25, "width": 20.1, "height": 10.2},
		}},
	}

	result, err := ExtractCoordinateAnnotations(document, "1:1", -1, false)

	require.NoError(t, err)
	assert.Equal(t, 51, result.CoordinateSpace.Width)
	assert.Equal(t, 41, result.CoordinateSpace.Height)
	require.Len(t, result.Annotations, 2)
	assert.Equal(t, "1:1", result.Annotations[0].ID)
	assert.Equal(t, 0, result.Annotations[0].Bounds.X)
	assert.Equal(t, 51, result.Annotations[0].Bounds.Width)
	assert.Equal(t, "1:2", result.Annotations[1].ID)
	assert.Equal(t, 12, result.Annotations[1].Bounds.X)
	assert.Equal(t, 33, result.Annotations[1].Bounds.X+result.Annotations[1].Bounds.Width)
	assert.Equal(t, "figma", result.Annotations[1].Metadata["source"])
	assert.Equal(t, "TEXT", result.Annotations[1].Metadata["nodeType"])
}

func TestExtractCoordinateAnnotationsOmitsHiddenAndUnboundedNodes(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "absoluteBoundingBox": map[string]any{"width": 10.0, "height": 10.0},
		"children": []any{
			map[string]any{"id": "hidden", "visible": false, "absoluteBoundingBox": map[string]any{"width": 2.0, "height": 2.0}},
			map[string]any{"id": "unbounded"},
		},
	}

	result, err := ExtractCoordinateAnnotations(document, "1:1", 2, false)

	require.NoError(t, err)
	assert.Len(t, result.Annotations, 1)
}
