package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindTextByLayerNameNonRecursive(t *testing.T) {
	document := map[string]any{
		"id":   "1:1",
		"name": "Target",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Immediate", "type": "TEXT", "characters": "Hello"},
			map[string]any{
				"id":   "1:3",
				"name": "Nested frame",
				"type": "FRAME",
				"children": []any{
					map[string]any{"id": "1:4", "name": "Nested", "type": "TEXT", "characters": "Hidden"},
				},
			},
		},
	}

	got := FindTextByLayerName(document, "Target", false)

	require.Len(t, got, 1)
	assert.Equal(t, "1:1", got[0].ID)
	require.Len(t, got[0].Texts, 1)
	assert.Equal(t, "Hello", got[0].Texts[0].Text)
}

func TestFindTextByLayerNameRecursive(t *testing.T) {
	document := map[string]any{
		"id":   "1:1",
		"name": "Target",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Immediate", "type": "TEXT", "characters": "Hello"},
			map[string]any{
				"id":   "1:3",
				"name": "Nested frame",
				"type": "FRAME",
				"children": []any{
					map[string]any{"id": "1:4", "name": "Nested", "type": "TEXT", "characters": "Hidden"},
				},
			},
		},
	}

	got := FindTextByLayerName(document, "Target", true)

	require.Len(t, got, 1)
	require.Len(t, got[0].Texts, 2)
	assert.Equal(t, "Hello", got[0].Texts[0].Text)
	assert.Equal(t, "Hidden", got[0].Texts[1].Text)
}

func TestFindTextByLayerNameReturnsAllMatches(t *testing.T) {
	document := map[string]any{
		"id":   "root",
		"name": "Root",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Target", "type": "TEXT", "characters": "First"},
			map[string]any{"id": "1:2", "name": "Target", "type": "TEXT", "characters": "Second"},
		},
	}

	got := FindTextByLayerName(document, "Target", false)

	assert.Len(t, got, 2)
}
