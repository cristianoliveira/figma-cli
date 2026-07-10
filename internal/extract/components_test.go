package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractComponents(t *testing.T) {
	const buttonID = "1:1"
	document := map[string]any{
		"id":                  "root",
		"name":                "Root",
		"type":                "COMPONENT",
		"opacity":             0.5,
		"cornerRadius":        8.0,
		"absoluteBoundingBox": map[string]any{"x": 1.0, "y": 2.0, "width": 100.0, "height": 50.0},
		"layoutMode":          "HORIZONTAL",
		"itemSpacing":         12.0,
		"paddingLeft":         16.0,
		"componentId":         "component-1",
		"componentSetId":      "set-1",
		"variantProperties":   map[string]any{"State": "Default"},
		"componentProperties": map[string]any{"Label#1:0": map[string]any{"type": "TEXT", "value": "Save"}},
		"componentPropertyDefinitions": map[string]any{
			"Label#1:0": map[string]any{"type": "TEXT", "defaultValue": "Button"},
		},
		"strokes":      []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0, "a": 1.0}}},
		"strokeWeight": 2.0,
		"strokeAlign":  "INSIDE",
		"children": []any{
			map[string]any{"id": buttonID, "name": "Button", "type": "INSTANCE"},
			map[string]any{"id": "1:2", "name": "Title", "type": "TEXT", "characters": "Hello", "style": map[string]any{"fontFamily": "Inter", "fontSize": 14.0, "fontWeight": 700.0, "letterSpacing": 0.2, "textAlignHorizontal": "CENTER"}},
		},
	}

	components := ExtractComponents(document)

	t.Run("count", func(t *testing.T) {
		assert.Len(t, components, 2)
	})

	root := components[0]
	rootTests := []struct {
		name string
		got  any
		want any
	}{
		{"opacity", *root.Opacity, 0.5},
		{"cornerRadius", *root.CornerRadius, 8.0},
		{"bounds.x", root.Bounds.X, 1.0},
		{"bounds.y", root.Bounds.Y, 2.0},
		{"bounds.width", root.Bounds.Width, 100.0},
		{"bounds.height", root.Bounds.Height, 50.0},
		{"layout.mode", root.Layout.Mode, "HORIZONTAL"},
		{"layout.gap", root.Layout.Gap, 12.0},
		{"layout.paddingLeft", root.Layout.PaddingLeft, 16.0},
		{"componentId", root.ComponentID, "component-1"},
		{"componentSetId", root.ComponentSetID, "set-1"},
		{"strokeWeight", root.StrokeWeight, 2.0},
		{"strokeAlign", root.StrokeAlign, "INSIDE"},
		{"variantProperties.State", root.VariantProperties["State"], "Default"},
		{"componentProperties.Label", root.ComponentProperties["Label#1:0"].(map[string]any)["value"], "Save"},
		{"propertyDefinitions.Label", root.PropertyDefinitions["Label#1:0"].(map[string]any)["defaultValue"], "Button"},
		{"path", root.Path, []string{"Root"}},
		{"paints.strokes[0].color", root.Paints.Strokes[0].Color, "#FF0000"},
	}

	for _, tt := range rootTests {
		t.Run("root/"+tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got, tt.name)
		})
	}

	t.Run("instance child", func(t *testing.T) {
		child := components[1]
		assert.Equal(t, buttonID, child.ID, "ID")
		assert.Equal(t, "Button", child.Name, "Name")
		assert.Equal(t, "INSTANCE", child.Type, "Type")
	})

	t.Run("generic text child is omitted", func(t *testing.T) {
		for _, component := range components {
			assert.NotEqual(t, "TEXT", component.Type)
		}
	})
}

func TestExtractComponentsFindsNestedDomainNodes(t *testing.T) {
	document := map[string]any{"id": "root", "type": "FRAME", "children": []any{
		map[string]any{"id": "frame", "type": "FRAME", "children": []any{
			map[string]any{"id": "set", "name": "Button", "type": "COMPONENT_SET"},
			map[string]any{"id": "detached", "name": "Detached", "type": "INSTANCE"},
		}},
	}}

	components := ExtractComponents(document)

	assert.Equal(t, []string{"set", "detached"}, []string{components[0].ID, components[1].ID})
	assert.Equal(t, []string{"Button"}, components[0].Path)
	assert.Equal(t, []string{"Detached"}, components[1].Path)
}

func TestAggregateComponentUsageGroupsInstancesByExactComponentID(t *testing.T) {
	components := []ComponentOutput{
		{ID: "1:1", Name: "Button", Type: "INSTANCE", ComponentID: "component-a", Path: []string{"Screen", "Primary"}, VariantProperties: map[string]any{"Size": "Large"}},
		{ID: "1:2", Name: "Renamed Button", Type: "INSTANCE", ComponentID: "component-a", Path: []string{"Screen", "Secondary"}, ComponentProperties: map[string]any{"Label": map[string]any{"value": "Cancel"}}},
		{ID: "1:3", Name: "Button", Type: "INSTANCE", ComponentID: "component-b", Path: []string{"Modal", "Primary"}},
		{ID: "1:4", Name: "Detached", Type: "INSTANCE", Path: []string{"Screen", "Detached"}},
		{ID: "1:5", Name: "Master", Type: "COMPONENT", ComponentID: "component-a"},
	}

	usage := AggregateComponentUsage(components)

	assert.Equal(t, []ComponentUsageOutput{
		{ComponentID: "component-a", Name: "Button", Count: 2, Instances: []ComponentInstanceUsage{
			{ID: "1:1", Name: "Button", Path: []string{"Screen", "Primary"}, VariantProperties: map[string]any{"Size": "Large"}},
			{ID: "1:2", Name: "Renamed Button", Path: []string{"Screen", "Secondary"}, ComponentProperties: map[string]any{"Label": map[string]any{"value": "Cancel"}}},
		}},
		{ComponentID: "component-b", Name: "Button", Count: 1, Instances: []ComponentInstanceUsage{{ID: "1:3", Name: "Button", Path: []string{"Modal", "Primary"}}}},
	}, usage)
}

func TestFilterComponentsByKind(t *testing.T) {
	components := []ComponentOutput{{Type: "COMPONENT"}, {Type: "COMPONENT_SET"}, {Type: "INSTANCE"}}

	assert.Equal(t, []ComponentOutput{{Type: "INSTANCE"}}, FilterComponentsByKind(components, "instance"))
	assert.Equal(t, components, FilterComponentsByKind(components, ""))
}

func TestPaintOutputsIncludeGradientAndImageIntent(t *testing.T) {
	paints := paintOutputsFromValue([]any{
		map[string]any{
			"type": "GRADIENT_LINEAR",
			"gradientStops": []any{
				map[string]any{"position": 0.0, "color": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0, "a": 1.0}},
				map[string]any{"position": 1.0, "color": map[string]any{"r": 0.0, "g": 0.0, "b": 1.0, "a": 0.5}},
			},
		},
		map[string]any{"type": "IMAGE", "imageRef": "image-1", "scaleMode": "FILL"},
	})

	assert.Equal(t, "GRADIENT_LINEAR", paints[0].Type)
	assert.Equal(t, []gradientStopOutput{
		{Position: 0, Color: "#FF0000", Opacity: 1},
		{Position: 1, Color: "#0000FF", Opacity: 0.5},
	}, paints[0].GradientStops)
	assert.Equal(t, "image-1", paints[1].ImageRef)
	assert.Equal(t, "FILL", paints[1].ScaleMode)
}

func TestPaintOutputsIgnoreMalformedGradientStops(t *testing.T) {
	paints := paintOutputsFromValue([]any{map[string]any{
		"type": "GRADIENT_LINEAR", "gradientStops": []any{"invalid"},
	}})

	assert.Empty(t, paints[0].GradientStops)
}

func TestExtractComponentsFromDocuments(t *testing.T) {
	documents := []any{
		map[string]any{"id": "1:1", "name": "First", "type": "COMPONENT"},
		map[string]any{"id": "2:2", "name": "Second", "type": "INSTANCE"},
	}

	got := ExtractComponentsFromDocuments(documents)

	assert.Len(t, got, 2)
	assert.Equal(t, "1:1", got[0].ID)
	assert.Equal(t, "2:2", got[1].ID)
}

func TestExtractRawComponentsFromDocuments(t *testing.T) {
	documents := []any{
		map[string]any{"id": "1:1", "children": []any{map[string]any{"id": "1:2"}}},
		map[string]any{"id": "2:2"},
	}

	got := ExtractRawComponentsFromDocuments(documents)

	assert.Len(t, got, 3)
	assert.Equal(t, "1:1", got[0]["id"])
	assert.Equal(t, "1:2", got[1]["id"])
	assert.Equal(t, "2:2", got[2]["id"])
}

func TestExtractRawComponents(t *testing.T) {
	const buttonID = "1:1"
	document := map[string]any{
		"id":   "root",
		"name": "Root",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": buttonID, "name": "Button", "type": "INSTANCE"},
		},
	}

	got := ExtractRawComponents(document)

	assert.Len(t, got, 2)
	assert.Equal(t, buttonID, got[1]["id"])
	assert.Equal(t, "Button", got[1]["name"])
}

func TestColorHexFromPaint(t *testing.T) {
	paint := map[string]any{"color": map[string]any{"r": 0.0, "g": 0.4, "b": 0.8, "a": 1.0}}

	got := colorHexFromPaint(paint)

	assert.Equal(t, "#0066CC", got)
}
