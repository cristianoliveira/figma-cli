package figma

import "testing"

func TestFindTextByLayerNameNonRecursive(t *testing.T) {
	document := map[string]any{
		"id":   "1:1",
		"name": "Target",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Immediate", "type": "TEXT", "characters": "Hello"},
			map[string]any{
				"id":   "1:3",
				"name": "Nested frame",
				"type": "FRAME",
				"children": []any{
					map[string]any{"id": "1:4", "name": "Nested", "type": "TEXT", "characters": "Hidden"},
				},
			},
		},
	}

	got := FindTextByLayerName(document, "Target", false)

	if len(got) != 1 {
		t.Fatalf("matches = %#v", got)
	}
	if got[0].ID != "1:1" || len(got[0].Texts) != 1 || got[0].Texts[0].Text != "Hello" {
		t.Fatalf("match = %#v", got[0])
	}
}

func TestFindTextByLayerNameRecursive(t *testing.T) {
	document := map[string]any{
		"id":   "1:1",
		"name": "Target",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Immediate", "type": "TEXT", "characters": "Hello"},
			map[string]any{
				"id":   "1:3",
				"name": "Nested frame",
				"type": "FRAME",
				"children": []any{
					map[string]any{"id": "1:4", "name": "Nested", "type": "TEXT", "characters": "Hidden"},
				},
			},
		},
	}

	got := FindTextByLayerName(document, "Target", true)

	if len(got) != 1 {
		t.Fatalf("matches = %#v", got)
	}
	if len(got[0].Texts) != 2 {
		t.Fatalf("texts = %#v", got[0].Texts)
	}
	if got[0].Texts[0].Text != "Hello" || got[0].Texts[1].Text != "Hidden" {
		t.Fatalf("texts = %#v", got[0].Texts)
	}
}

func TestFindTextByLayerNameReturnsAllMatches(t *testing.T) {
	document := map[string]any{
		"id":   "root",
		"name": "Root",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Target", "type": "TEXT", "characters": "First"},
			map[string]any{"id": "1:2", "name": "Target", "type": "TEXT", "characters": "Second"},
		},
	}

	got := FindTextByLayerName(document, "Target", false)

	if len(got) != 2 {
		t.Fatalf("matches = %#v", got)
	}
}
