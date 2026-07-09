package cmd

import "testing"

func TestFindLayersByName(t *testing.T) {
	document := map[string]any{
		"id":   "root",
		"name": "Root",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Target", "type": "FRAME"},
			map[string]any{"id": "1:2", "name": "Target", "type": "TEXT"},
		},
	}

	got := findLayersByName(document, "Target")

	if len(got) != 2 {
		t.Fatalf("matches = %#v", got)
	}
	if got[0].ID != "1:1" || got[1].ID != "1:2" {
		t.Fatalf("matches = %#v", got)
	}
}

func TestBuildFindAPIURLWithNodeIDs(t *testing.T) {
	got, err := buildFindAPIURL("file123", []string{"1:2"})
	if err != nil {
		t.Fatalf("buildFindAPIURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/files/file123/nodes?ids=1%3A2"
	if got != expected {
		t.Fatalf("buildFindAPIURL() = %v, expected %v", got, expected)
	}
}
