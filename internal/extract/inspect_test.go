package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestInspectTreeReturnsImplementationSpecsInTreeOrder(t *testing.T) {
	document := map[string]any{"id": "1", "name": "Card", "type": "COMPONENT", "children": []any{
		map[string]any{"id": "2", "name": "Label", "type": "TEXT", "characters": "Save"},
	}}

	outputs := InspectTree(document)

	require.Len(t, outputs, 2)
	assert.Equal(t, "1", outputs[0].ID)
	assert.Equal(t, "2", outputs[1].ID)
	assert.Equal(t, "Save", outputs[1].Text)
	assert.Nil(t, outputs[0].RelativeBounds)
	assert.Nil(t, outputs[1].RelativeBounds)
}

func TestInspectTreeMeasuresAdjacentAutoLayoutSiblingSpacing(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Stack", "type": "FRAME", "layoutMode": "VERTICAL", "itemSpacing": 0.0,
		"children": []any{
			map[string]any{"id": "1:2", "name": "Sales copy", "type": "TEXT", "characters": "Talk to sales", "absoluteBoundingBox": map[string]any{"x": 0.0, "y": 253.718, "width": 413.0, "height": 48.0}},
			map[string]any{"id": "1:3", "name": "Contact sales", "type": "TEXT", "characters": "Contact sales", "absoluteBoundingBox": map[string]any{"x": 0.0, "y": 301.718, "width": 107.0, "height": 24.0}},
		},
	}

	outputs := InspectTree(document)

	require.Len(t, outputs, 3)
	assert.Nil(t, outputs[1].SpacingFromPrevious)
	assert.Equal(t, &LayoutSpacing{
		ParentID: "1:1", PreviousID: "1:2", Axis: "vertical", Measured: 0, Declared: numberPointer(0), MatchesDeclared: true,
	}, outputs[2].SpacingFromPrevious)
}

func TestInspectTreeSkipsHiddenAndAbsoluteChildrenWhenMeasuringSpacing(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Row", "type": "FRAME", "layoutMode": "HORIZONTAL", "itemSpacing": 8.0,
		"children": []any{
			map[string]any{"id": "1:2", "name": "A", "type": "RECTANGLE", "absoluteBoundingBox": map[string]any{"x": 0.0, "y": 0.0, "width": 10.0, "height": 10.0}},
			map[string]any{"id": "1:9", "name": "Hidden", "type": "RECTANGLE", "visible": false, "absoluteBoundingBox": map[string]any{"x": 18.0, "y": 0.0, "width": 10.0, "height": 10.0}},
			map[string]any{"id": "1:8", "name": "Absolute", "type": "RECTANGLE", "layoutPositioning": "ABSOLUTE", "absoluteBoundingBox": map[string]any{"x": 18.0, "y": 0.0, "width": 10.0, "height": 10.0}},
			map[string]any{"id": "1:3", "name": "B", "type": "RECTANGLE", "absoluteBoundingBox": map[string]any{"x": 25.0, "y": 0.0, "width": 10.0, "height": 10.0}},
		},
	}

	outputs := InspectTree(document)

	require.Len(t, outputs, 5)
	assert.Nil(t, outputs[2].SpacingFromPrevious)
	assert.Nil(t, outputs[3].SpacingFromPrevious)
	assert.Equal(t, &LayoutSpacing{
		ParentID: "1:1", PreviousID: "1:2", Axis: "horizontal", Measured: 15, Declared: numberPointer(8), MatchesDeclared: false,
	}, outputs[4].SpacingFromPrevious)
}

func TestInspectTreeOmitsSpacingForNonAutoLayoutParents(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Group", "type": "GROUP",
		"children": []any{
			map[string]any{"id": "1:2", "name": "A", "type": "RECTANGLE", "absoluteBoundingBox": map[string]any{"x": 0.0, "y": 0.0, "width": 10.0, "height": 10.0}},
			map[string]any{"id": "1:3", "name": "B", "type": "RECTANGLE", "absoluteBoundingBox": map[string]any{"x": 20.0, "y": 0.0, "width": 10.0, "height": 10.0}},
		},
	}

	outputs := InspectTree(document)

	require.Len(t, outputs, 3)
	assert.Nil(t, outputs[2].SpacingFromPrevious)
}

func TestInspectTreeRelativeToScopePreservesAbsoluteBoundsAndFractionalRelativeBounds(t *testing.T) {
	document := map[string]any{
		"id": "13576:15248", "name": "Scope", "type": "FRAME",
		"absoluteBoundingBox": map[string]any{"x": 136.0, "y": 562.0, "width": 400.0, "height": 300.0},
		"children": []any{
			map[string]any{
				"id": "13576:15249", "name": "Icon", "type": "VECTOR",
				"absoluteBoundingBox": map[string]any{"x": 435.464, "y": 623.0, "width": 9.071, "height": 16.0},
				"children": []any{
					map[string]any{
						"id": "13576:15250", "name": "Nested", "type": "VECTOR",
						"absoluteBoundingBox": map[string]any{"x": 436.464, "y": 624.5, "width": 2.5, "height": 3.25},
					},
				},
			},
		},
	}

	outputs := InspectTreeRelativeToScope(document, "13576:15248")

	require.Len(t, outputs, 3)
	assert.Equal(t, boundsOutput{X: 136, Y: 562, Width: 400, Height: 300}, outputs[0].Bounds)
	assert.Equal(t, &relativeBoundsOutput{X: 0, Y: 0, Width: 400, Height: 300, RelativeTo: "13576:15248"}, outputs[0].RelativeBounds)
	assert.Equal(t, boundsOutput{X: 435.464, Y: 623, Width: 9.071, Height: 16}, outputs[1].Bounds)
	assert.Equal(t, &relativeBoundsOutput{X: 299.464, Y: 61, Width: 9.071, Height: 16, RelativeTo: "13576:15248"}, outputs[1].RelativeBounds)
	assert.Equal(t, &relativeBoundsOutput{X: 300.464, Y: 62.5, Width: 2.5, Height: 3.25, RelativeTo: "13576:15248"}, outputs[2].RelativeBounds)
}

func TestNodeToInspectOutputIncludesComponentAndMixedTextProperties(t *testing.T) {
	node := map[string]any{
		"variantProperties": map[string]any{"State": "Default"},
		"componentProperties": map[string]any{
			"Label#1:0": map[string]any{"type": "TEXT", "value": "Save"},
		},
		"componentPropertyDefinitions": map[string]any{
			"Disabled#1:1": map[string]any{"type": "BOOLEAN", "defaultValue": false},
		},
		"characterStyleOverrides": []any{0.0, 1.0, 1.0},
		"styleOverrideTable": map[string]any{
			"1": map[string]any{"fontFamily": "Inter", "fontSize": 16.0, "fontWeight": 700.0},
		},
	}

	out := NodeToInspectOutput(node)

	assert.Equal(t, "Default", out.VariantProperties["State"])
	assert.Equal(t, "Save", out.ComponentProperties["Label#1:0"].(map[string]any)["value"])
	assert.Equal(t, false, out.PropertyDefinitions["Disabled#1:1"].(map[string]any)["defaultValue"])
	assert.Equal(t, []int{1}, out.StyleOverrideIDs)
	assert.Equal(t, typographyOutput{FontFamily: "Inter", FontSize: 16, FontWeight: 700}, out.StyleOverrides["1"])
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
