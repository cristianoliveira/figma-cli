package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNodeByID(t *testing.T) {
	doc := map[string]any{
		"id": "0:0", "name": "root", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "A", "type": "FRAME"},
			map[string]any{"id": "1:2", "name": "B", "type": "FRAME", "children": []any{
				map[string]any{"id": "1:3", "name": "C", "type": "TEXT"},
			}},
		},
	}

	assert.Nil(t, FindNodeByID(doc, "missing"))

	got := FindNodeByID(doc, "1:3")
	if assert.NotNil(t, got) {
		assert.Equal(t, "C", got["name"])
	}
}

func TestNodeToInspectOutput(t *testing.T) {
	node := map[string]any{
		"id":                  "1:1",
		"name":                "Card",
		"type":                "FRAME",
		"fills":               []any{paint("SOLID", 1, 0, 0, true)},
		"absoluteBoundingBox": map[string]any{"x": 1.0, "y": 2.0, "width": 10.0, "height": 20.0},
		"backgroundColor":     map[string]any{"color": map[string]any{"r": 0.0, "g": 0.0, "b": 0.0, "a": 1.0}},
	}

	out := NodeToInspectOutput(node)

	assert.Equal(t, "1:1", out.ID)
	assert.Equal(t, "Card", out.Name)
	assert.Equal(t, "FRAME", out.Type)
	assert.Equal(t, 10.0, out.Bounds.Width)
	assert.Equal(t, 20.0, out.Bounds.Height)
	assert.Len(t, out.Fills, 1)
	assert.Equal(t, "#FF0000", out.Fills[0])
	assert.Equal(t, "#000000", out.BackgroundColor)
}

func TestNodeToInspectOutputIncludesStyleAndVariableBindings(t *testing.T) {
	node := map[string]any{
		"id": "1:1", "name": "Button", "type": "FRAME",
		"styles": map[string]any{"fill": "S:fill", "effect": "S:shadow"},
		"boundVariables": map[string]any{
			"fills":        []any{map[string]any{"type": "VARIABLE_ALIAS", "id": "V:brand"}},
			"cornerRadius": map[string]any{"type": "VARIABLE_ALIAS", "id": "V:radius"},
		},
	}

	out := NodeToInspectOutput(node)

	assert.Equal(t, map[string]string{"fill": "S:fill", "effect": "S:shadow"}, out.StyleBindings)
	assert.Equal(t, map[string][]string{"fills": {"V:brand"}, "cornerRadius": {"V:radius"}}, out.VariableBindings)
}

func TestNodeToInspectOutputIgnoresMalformedBindings(t *testing.T) {
	node := map[string]any{
		"styles":         map[string]any{"fill": 42},
		"boundVariables": map[string]any{"fills": []any{"invalid"}},
	}

	out := NodeToInspectOutput(node)

	assert.Empty(t, out.StyleBindings)
	assert.Empty(t, out.VariableBindings)
}
