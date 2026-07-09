package extract

import "testing"

func TestFindLayersByName(t *testing.T) {
	doc := map[string]any{
		"id": "0:0", "name": "root", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Button", "type": "FRAME"},
			map[string]any{"id": "1:2", "name": "Other", "type": "FRAME", "children": []any{
				map[string]any{"id": "1:3", "name": "Button", "type": "INSTANCE"},
			}},
		},
	}

	matches := FindLayersByName(doc, "Button")
	if len(matches) != 2 {
		t.Fatalf("got %d matches, want 2: %#v", len(matches), matches)
	}
	if matches[0].ID != "1:1" || matches[1].ID != "1:3" {
		t.Errorf("match ids = %v, want 1:1 and 1:3", matches)
	}

	// Matching is exact, not substring.
	if len(FindLayersByName(doc, "But")) != 0 {
		t.Errorf("partial name matched; want exact only")
	}
}

func TestFindLayersByNameNil(t *testing.T) {
	if got := FindLayersByName(nil, "x"); got != nil {
		t.Errorf("FindLayersByName(nil) = %#v, want nil", got)
	}
}
