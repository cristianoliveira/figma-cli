package extract

import "testing"

func paint(typ string, r, g, b float64, visible bool) map[string]any {
	return map[string]any{
		"type":    typ,
		"visible": visible,
		"color":   map[string]any{"r": r, "g": g, "b": b, "a": 1.0},
	}
}

func TestCollectColors(t *testing.T) {
	doc := map[string]any{
		"name":  "root",
		"fills": []any{paint("SOLID", 1, 0, 0, true)},
		"children": []any{
			map[string]any{"name": "a", "fills": []any{paint("SOLID", 1, 0, 0, true), paint("SOLID", 0, 1, 0, true)}},
			map[string]any{"name": "b", "fills": []any{paint("SOLID", 0, 0, 0, true)}},  // black -> filtered
			map[string]any{"name": "c", "fills": []any{paint("SOLID", 1, 0, 0, false)}}, // hidden -> filtered
			map[string]any{"name": "d", "backgroundColor": map[string]any{"color": map[string]any{"r": 0.0, "g": 0.0, "b": 1.0, "a": 1.0}}},
		},
	}

	palette := CollectColors(doc)

	if len(palette) != 3 {
		t.Fatalf("got %d entries, want 3: %#v", len(palette), palette)
	}
	if palette[0].Color != "#FF0000" || palette[0].Count != 2 {
		t.Errorf("first = %#v, want #FF0000 count 2 (sorted desc)", palette[0])
	}
	// background color collected with "(bg)" suffix
	var bg *ColorEntry
	for i := range palette {
		if palette[i].Color == "#0000FF" {
			bg = &palette[i]
		}
	}
	if bg == nil || len(bg.Usage) == 0 || bg.Usage[0] != "d (bg)" {
		t.Errorf("background entry = %#v, want Usage containing 'd (bg)'", bg)
	}
}

func TestCollectColorsEmpty(t *testing.T) {
	palette := CollectColors(map[string]any{"id": "0:0", "name": "empty", "type": "FRAME"})
	if palette == nil {
		t.Fatal("got nil, want non-empty slice")
	}
	if len(palette) != 0 {
		t.Errorf("got %d entries, want 0", len(palette))
	}
}
