package cmd

import "testing"

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

	components := extractComponents(document)

	t.Run("count", func(t *testing.T) {
		if len(components) != 3 {
			t.Fatalf("got %d components, want 3", len(components))
		}
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
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	t.Run("instance child", func(t *testing.T) {
		child := components[1]
		if child.ID != buttonID {
			t.Errorf("ID = %v, want %v", child.ID, buttonID)
		}
		if child.Name != "Button" {
			t.Errorf("Name = %v, want Button", child.Name)
		}
		if child.Type != "INSTANCE" {
			t.Errorf("Type = %v, want INSTANCE", child.Type)
		}
	})

	t.Run("text child", func(t *testing.T) {
		text := components[2]
		if text.Text != "Hello" {
			t.Errorf("Text = %v, want Hello", text.Text)
		}
		if text.Typography.FontFamily != "Inter" {
			t.Errorf("FontFamily = %v, want Inter", text.Typography.FontFamily)
		}
		if text.Typography.FontWeight != 700 {
			t.Errorf("FontWeight = %v, want 700", text.Typography.FontWeight)
		}
		if text.Typography.LetterSpacing != 0.2 {
			t.Errorf("LetterSpacing = %v, want 0.2", text.Typography.LetterSpacing)
		}
		if text.Typography.TextAlignHorizontal != "CENTER" {
			t.Errorf("TextAlignHorizontal = %v, want CENTER", text.Typography.TextAlignHorizontal)
		}
	})
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

	got := extractRawComponents(document)

	if len(got) != 2 {
		t.Fatalf("raw components = %#v", got)
	}
	if got[1]["id"] != buttonID || got[1]["name"] != "Button" {
		t.Fatalf("raw component = %#v", got[1])
	}
}

func TestColorHexFromPaint(t *testing.T) {
	paint := map[string]any{"color": map[string]any{"r": 0.0, "g": 0.4, "b": 0.8, "a": 1.0}}

	got := colorHexFromPaint(paint)

	if got != "#0066CC" {
		t.Fatalf("colorHexFromPaint() = %v, expected #0066CC", got)
	}
}
