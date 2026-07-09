package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindLayersByName(t *testing.T) {
	doc := map[string]any{
		"id": "0:0", "name": "root", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Button", "type": "FRAME"},
			map[string]any{"id": "1:2", "name": "Other", "type": "FRAME", "children": []any{
				map[string]any{"id": "1:3", "name": "Button", "type": "INSTANCE"},
			}},
		},
	}

	matches := FindLayersByName(doc, "Button")
	assert.Len(t, matches, 2)
	assert.Equal(t, "1:1", matches[0].ID)
	assert.Equal(t, "1:3", matches[1].ID)

	// Matching is exact, not substring.
	assert.Empty(t, FindLayersByName(doc, "But"))
}

func TestFindLayersByNameNil(t *testing.T) {
	assert.Nil(t, FindLayersByName(nil, "x"))
}
