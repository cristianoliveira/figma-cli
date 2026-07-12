package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeToInspectOutputPreservesVectorGeometry(t *testing.T) {
	node := map[string]any{
		"id": "1:2", "name": "Wave", "type": "VECTOR",
		"fillGeometry": []any{
			map[string]any{"path": "M0 0 C 1 2 3 4 5 6 Z", "windingRule": "NONZERO", "overrideID": 7.0},
			map[string]any{"path": "M1 1 L2 2", "windingRule": "EVENODD"},
		},
		"strokeGeometry":    []any{map[string]any{"path": "M0 0 L5 6", "windingRule": "NONZERO"}},
		"relativeTransform": []any{[]any{1.0, 0.0, 12.0}, []any{0.0, 1.0, 8.0}},
		"size":              map[string]any{"x": 5.0, "y": 6.0},
		"fillOverrideTable": map[string]any{"7": map[string]any{"fills": []any{}}},
	}

	result := NodeToInspectOutput(node)

	require.Len(t, result.FillGeometry, 2)
	assert.Equal(t, "M0 0 C 1 2 3 4 5 6 Z", result.FillGeometry[0].Path)
	assert.Equal(t, "NONZERO", result.FillGeometry[0].WindingRule)
	assert.Equal(t, 7.0, *result.FillGeometry[0].OverrideID)
	require.Len(t, result.StrokeGeometry, 1)
	assert.Equal(t, "M0 0 L5 6", result.StrokeGeometry[0].Path)
	assert.NotNil(t, result.RelativeTransform)
	assert.NotNil(t, result.VectorSize)
	assert.Contains(t, result.FillOverrideTable, "7")
}
