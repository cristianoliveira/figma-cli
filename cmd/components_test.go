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

	got := extractComponents(document)

	if len(got) != 3 {
		t.Fatalf("components = %#v", got)
	}
	if got[0].Opacity == nil || *got[0].Opacity != 0.5 || got[0].CornerRadius == nil || *got[0].CornerRadius != 8 {
		t.Fatalf("root component = %#v", got[0])
	}
	if got[0].Bounds.Width != 100 || got[0].Layout.Mode != "HORIZONTAL" || got[0].Layout.PaddingLeft != 16 {
		t.Fatalf("root component = %#v", got[0])
	}
	if got[0].ComponentID != "component-1" || got[0].ComponentSetID != "set-1" || got[0].VariantProperties["State"] != "Default" {
		t.Fatalf("root component identity = %#v", got[0])
	}
	if got[0].StrokeWeight != 2 || got[0].StrokeAlign != "INSIDE" || got[0].Paints.Strokes[0].Color != "#FF0000" {
		t.Fatalf("root component strokes = %#v", got[0])
	}
	if got[1].ID != buttonID || got[1].Name != "Button" || got[1].Type != "INSTANCE" {
		t.Fatalf("component = %#v", got[1])
	}
	if got[2].Text != "Hello" || got[2].Typography.FontFamily != "Inter" || got[2].Typography.FontWeight != 700 {
		t.Fatalf("text component = %#v", got[2])
	}
	if got[2].Typography.LetterSpacing != 0.2 || got[2].Typography.TextAlignHorizontal != "CENTER" {
		t.Fatalf("text typography = %#v", got[2])
	}
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
