package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedTextForFramePreservesTreeOrderAndContext(t *testing.T) {
	frame := map[string]any{
		"id": "1:1", "name": "Dialog", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Title", "type": "TEXT", "characters": "Welcome"},
			map[string]any{
				"id": "1:3", "name": "Items", "type": "FRAME",
				"children": []any{
					map[string]any{
						"id": "1:4", "name": "Generic name", "type": "TEXT", "characters": "First item\nSecond item",
						"lineTypes": []any{"UNORDERED", "UNORDERED"}, "lineIndentations": []any{0.0, 1.0},
						"characterStyleOverrides": []any{0.0, 1.0, 1.0},
						"styleOverrideTable":      map[string]any{"1": map[string]any{"fontFamily": "Inter", "fontWeight": 700.0}},
					},
				},
			},
		},
	}

	got := OrderedTextForFrame(frame)

	assert.Equal(t, []OrderedTextOutput{
		{ID: "1:2", Name: "Title", Text: "Welcome", NodeKind: "textBlock", Depth: 1, Order: 0, ParentName: "Dialog"},
		{
			ID: "1:4", Name: "Generic name", Text: "First item\nSecond item", NodeKind: "textBlock", Depth: 2, Order: 1, ParentName: "Items",
			Lines: []TextLineOutput{
				{Index: 0, Text: "First item", ListType: "UNORDERED"},
				{Index: 1, Text: "Second item", ListType: "UNORDERED", Indentation: 1},
			},
			StyleOverrideIDs: []int{1},
			StyleOverrides:   map[string]typographyOutput{"1": {FontFamily: "Inter", FontWeight: 700}},
		},
	}, got)
}

func TestOrderedTextForFrameOmitsLinesWhenMetadataHasNoListIntent(t *testing.T) {
	frame := map[string]any{"type": "FRAME", "children": []any{
		map[string]any{"id": "1", "name": "Plain", "type": "TEXT", "characters": "Plain copy", "lineTypes": []any{"NONE"}, "lineIndentations": []any{0.0}},
	}}

	got := OrderedTextForFrame(frame)

	assert.Empty(t, got[0].Lines)
}

func TestOrderedTextForFrameDoesNotInferListsFromGlyphsOrNewlines(t *testing.T) {
	frame := map[string]any{"type": "FRAME", "children": []any{
		map[string]any{"id": "1", "name": "Copy", "type": "TEXT", "characters": "• First\n• Second"},
		map[string]any{"id": "2", "name": "Item", "type": "TEXT", "characters": "Separate first"},
		map[string]any{"id": "3", "name": "Item", "type": "TEXT", "characters": "Separate second"},
	}}

	got := OrderedTextForFrame(frame)

	assert.Len(t, got, 3)
	assert.Empty(t, got[0].Lines)
	assert.Equal(t, "textBlock", got[0].NodeKind)
	assert.Equal(t, []string{"Separate first", "Separate second"}, []string{got[1].Text, got[2].Text})
}

func TestOrderedTextForFrameReturnsEmptyForInvalidDocument(t *testing.T) {
	assert.Empty(t, OrderedTextForFrame(nil))
}

func TestTextEqual(t *testing.T) {
	a := []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "yo"}}

	assert.True(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "yo"}}))
	assert.False(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "no"}}), "changed text")
	assert.False(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}}), "removed node")
	assert.False(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "yo"}, {ID: "1:3", Text: "x"}}), "added node")
}
