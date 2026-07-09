package extract

import "testing"

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

	if got := FindNodeByID(doc, "missing"); got != nil {
		t.Errorf("missing returned %#v, want nil", got)
	}

	got := FindNodeByID(doc, "1:3")
	if got == nil || got["name"] != "C" {
		t.Fatalf("FindNodeByID(1:3) = %#v, want node C", got)
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

	if out.ID != "1:1" || out.Name != "Card" || out.Type != "FRAME" {
		t.Errorf("identity = %#v", out)
	}
	if out.Bounds.Width != 10 || out.Bounds.Height != 20 {
		t.Errorf("bounds = %#v", out.Bounds)
	}
	if len(out.Fills) != 1 || out.Fills[0] != "#FF0000" {
		t.Errorf("fills = %v", out.Fills)
	}
	if out.BackgroundColor != "#000000" {
		t.Errorf("backgroundColor = %q, want #000000", out.BackgroundColor)
	}
}
