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
		"type":                "FRAME",
		"opacity":             0.5,
		"cornerRadius":        8.0,
		"absoluteBoundingBox": map[string]any{"x": 1.0, "y": 2.0, "width": 100.0, "height": 50.0},
		"layoutMode":          "HORIZONTAL",
		"itemSpacing":         12.0,
		"paddingLeft":         16.0,
		"componentId":         "component-1",
		"componentSetId":      "set-1",
		"variantProperties":   map[string]any{"State": "Default"},
		"strokes":             []any{map[string]any{"type": "SOLID", "color": map[string]any{"r": 1.0, "g": 0.0, "b": 0.0, "a": 1.0}}},
		"strokeWeight":        2.0,
		"strokeAlign":         "INSIDE",
		"children": []any{
			map[string]any{"id": buttonID, "name": "Button", "type": "INSTANCE"},
			map[string]any{"id": "1:2", "name": "Title", "type": "TEXT", "characters": "Hello", "style": map[string]any{"fontFamily": "Inter", "fontSize": 14.0, "fontWeight": 700.0, "letterSpacing": 0.2, "textAlignHorizontal": "CENTER"}},
		},
	}

	components := ExtractComponents(document)

	t.Run("count", func(t *testing.T) {
		assert.Len(t, components, 3)
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

	t.Run("text child", func(t *testing.T) {
		text := components[2]
		assert.Equal(t, "Hello", text.Text, "Text")
		assert.Equal(t, "Inter", text.Typography.FontFamily, "FontFamily")
		assert.Equal(t, float64(700), text.Typography.FontWeight, "FontWeight")
		assert.Equal(t, 0.2, text.Typography.LetterSpacing, "LetterSpacing")
		assert.Equal(t, "CENTER", text.Typography.TextAlignHorizontal, "TextAlignHorizontal")
	})
}

func TestExtractComponentsFromDocuments(t *testing.T) {
	documents := []any{
		map[string]any{"id": "1:1", "name": "First", "type": "FRAME"},
		map[string]any{"id": "2:2", "name": "Second", "type": "FRAME"},
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
