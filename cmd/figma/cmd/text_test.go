package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/api"
)

func TestCollectTexts(t *testing.T) {
	// Build a tree with TEXT nodes at different depths
	root := &api.Node{
		ID:   "0",
		Name: "Root",
		Type: "FRAME",
		Children: []api.Node{
			{
				ID:         "1",
				Name:       "Text1",
				Type:       "TEXT",
				Characters: "Hello",
			},
			{
				ID:   "2",
				Name: "Frame1",
				Type: "FRAME",
				Children: []api.Node{
					{
						ID:         "3",
						Name:       "Text2",
						Type:       "TEXT",
						Characters: "World",
					},
					{
						ID:         "4",
						Name:       "EmptyText",
						Type:       "TEXT",
						Characters: "", // empty, should be skipped
					},
					{
						ID:   "5",
						Name: "Rect",
						Type: "RECTANGLE",
					},
				},
			},
			{
				ID:   "6",
				Name: "Component1",
				Type: "COMPONENT",
				Children: []api.Node{
					{
						ID:         "7",
						Name:       "Text3",
						Type:       "TEXT",
						Characters: "Nested",
					},
				},
			},
		},
	}

	// Test non-recursive (maxDepth = 1)
	items := collectTexts(root, false)
	if len(items) != 1 {
		t.Errorf("Expected 1 text item with non-recursive, got %d", len(items))
	}
	if len(items) > 0 && items[0].Characters != "Hello" {
		t.Errorf("Expected 'Hello', got %s", items[0].Characters)
	}

	// Test recursive (unlimited depth)
	items = collectTexts(root, true)
	if len(items) != 3 {
		t.Errorf("Expected 3 text items with recursive, got %d", len(items))
	}
	// Check that we have Hello, World, Nested
	expected := map[string]bool{"Hello": false, "World": false, "Nested": false}
	for _, item := range items {
		expected[item.Characters] = true
	}
	for chars, found := range expected {
		if !found {
			t.Errorf("Expected text %q not found", chars)
		}
	}

	// Test empty characters skipped
	for _, item := range items {
		if item.Characters == "" {
			t.Errorf("Found empty characters, should be skipped")
		}
	}
}

func TestCollectTextsDepth(t *testing.T) {
	// Build a deeper tree
	root := &api.Node{
		ID:   "0",
		Name: "Root",
		Type: "FRAME",
		Children: []api.Node{
			{
				ID:   "1",
				Name: "Level1",
				Type: "FRAME",
				Children: []api.Node{
					{
						ID:         "2",
						Name:       "Level2",
						Type:       "TEXT",
						Characters: "Deep",
					},
				},
			},
		},
	}

	// Non-recursive should not find deep text
	items := collectTexts(root, false)
	if len(items) != 0 {
		t.Errorf("Expected 0 text items with non-recursive (depth 2), got %d", len(items))
	}

	// Recursive should find it
	items = collectTexts(root, true)
	if len(items) != 1 || items[0].Characters != "Deep" {
		t.Errorf("Expected 'Deep' text with recursive, got %v", items)
	}
}
