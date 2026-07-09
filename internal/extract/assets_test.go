package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAssets_FindsInstancesRasterAndStandaloneVectors(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "type": "FRAME",
		"children": []any{
			map[string]any{
				"id": "1:2", "name": "Logo", "type": "INSTANCE",
				"children": []any{map[string]any{"id": "I1:2;1:3", "name": "Path", "type": "VECTOR"}},
			},
			map[string]any{"id": "1:4", "name": "Photo", "type": "RECTANGLE", "fills": []any{map[string]any{"type": "IMAGE"}}},
			map[string]any{"id": "1:5", "name": "Decoration", "type": "VECTOR"},
		},
	}

	assets := ExtractAssets([]any{document})

	assert.Equal(t, []Asset{
		{ID: "1:2", Name: "Logo", Kind: "instance", Format: "svg"},
		{ID: "1:4", Name: "Photo", Kind: "image", Format: "png"},
		{ID: "1:5", Name: "Decoration", Kind: "vector", Format: "svg"},
	}, assets)
}

func TestExtractAssets_DeduplicatesIDsAndIgnoresInternalVectors(t *testing.T) {
	instance := map[string]any{
		"id": "1:2", "name": "Icon", "type": "INSTANCE",
		"children": []any{map[string]any{"id": "I1:2;1:3", "name": "Path", "type": "BOOLEAN_OPERATION"}},
	}

	assets := ExtractAssets([]any{instance, instance})

	assert.Equal(t, []Asset{{ID: "1:2", Name: "Icon", Kind: "instance", Format: "svg"}}, assets)
}

func TestExtractAssets_EmptyDocument(t *testing.T) {
	assert.Empty(t, ExtractAssets(nil))
}
