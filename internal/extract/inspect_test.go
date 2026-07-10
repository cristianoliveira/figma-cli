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

func TestNodeToInspectOutputIncludesGradientAndImagePaints(t *testing.T) {
	node := map[string]any{"fills": []any{
		map[string]any{"type": "GRADIENT_LINEAR", "gradientStops": []any{
			map[string]any{"position": 0.0, "color": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0, "a": 1.0}},
		}},
		map[string]any{"type": "IMAGE", "imageRef": "image-1", "scaleMode": "FIT"},
	}}

	out := NodeToInspectOutput(node)

	if assert.NotNil(t, out.Paints) {
		assert.Equal(t, "#FF0000", out.Paints.Fills[0].GradientStops[0].Color)
		assert.Equal(t, "image-1", out.Paints.Fills[1].ImageRef)
		assert.Equal(t, "FIT", out.Paints.Fills[1].ScaleMode)
	}
}

func TestNodeToInspectOutputIncludesResponsiveLayoutAndEffects(t *testing.T) {
	node := map[string]any{
		"layoutMode":         "HORIZONTAL",
		"layoutWrap":         "WRAP",
		"counterAxisSpacing": 24.0,
		"layoutPositioning":  "ABSOLUTE",
		"minWidth":           120.0,
		"maxWidth":           480.0,
		"minHeight":          40.0,
		"maxHeight":          200.0,
		"constraints":        map[string]any{"horizontal": "STRETCH", "vertical": "TOP"},
		"effects": []any{map[string]any{
			"type": "DROP_SHADOW", "visible": true, "radius": 8.0, "spread": 2.0,
			"blendMode": "MULTIPLY", "offset": map[string]any{"x": 0.0, "y": 4.0},
		}},
	}

	out := NodeToInspectOutput(node)

	assert.Equal(t, "WRAP", out.Layout.Wrap)
	assert.Equal(t, 24.0, out.Layout.CounterAxisSpacing)
	assert.Equal(t, "ABSOLUTE", out.Layout.Positioning)
	assert.Equal(t, 120.0, out.Layout.MinWidth)
	assert.Equal(t, 480.0, out.Layout.MaxWidth)
	assert.Equal(t, "STRETCH", out.Layout.Constraints.Horizontal)
	assert.Equal(t, "TOP", out.Layout.Constraints.Vertical)
	assert.Equal(t, 4.0, out.Effects[0].OffsetY)
	assert.Equal(t, 2.0, out.Effects[0].Spread)
	assert.Equal(t, "MULTIPLY", out.Effects[0].BlendMode)
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

func TestResolveInspectStyleBindingsAddsNamesAndTypes(t *testing.T) {
	output := NodeToInspectOutput(map[string]any{"styles": map[string]any{"fill": "S:fill", "effect": "S:missing"}})
	styles := map[string]map[string]any{
		"S:fill": {"name": "Brand/Primary", "styleType": "FILL"},
	}

	ResolveInspectStyleBindings(&output, styles)

	assert.Equal(t, map[string]StyleBinding{
		"fill":   {ID: "S:fill", Name: "Brand/Primary", Type: "FILL"},
		"effect": {ID: "S:missing"},
	}, output.ResolvedStyles)
}

func TestResolveInspectVariableBindingsAddsVariableAndCollectionNames(t *testing.T) {
	output := NodeToInspectOutput(map[string]any{"boundVariables": map[string]any{
		"fills": []any{map[string]any{"type": "VARIABLE_ALIAS", "id": "V:brand"}},
	}})
	meta := map[string]any{
		"variables": map[string]any{"V:brand": map[string]any{
			"name": "Color/Brand", "resolvedType": "COLOR", "variableCollectionId": "C:theme",
		}},
		"variableCollections": map[string]any{"C:theme": map[string]any{"name": "Theme"}},
	}

	ResolveInspectVariableBindings(&output, meta)

	assert.Equal(t, []VariableBinding{{
		ID: "V:brand", Name: "Color/Brand", Type: "COLOR", CollectionID: "C:theme", CollectionName: "Theme",
	}}, output.ResolvedVariables["fills"])
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
