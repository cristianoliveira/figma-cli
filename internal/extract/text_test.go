package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedTextForFramePreservesTreeOrderAndContext(t *testing.T) {
	frame := map[string]any{
		"id": "1:1", "name": "Dialog", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Title", "type": "TEXT", "characters": "Welcome", "lineTypes": []any{"NONE"}},
			map[string]any{
				"id": "1:3", "name": "Items", "type": "FRAME",
				"children": []any{
					map[string]any{
						"id": "1:4", "name": "Generic name", "type": "TEXT", "characters": "First item",
						"lineTypes": []any{"UNORDERED"},
					},
				},
			},
		},
	}

	got := OrderedTextForFrame(frame)

	assert.Equal(t, []OrderedTextOutput{
		{ID: "1:2", Name: "Title", Text: "Welcome", Depth: 1, Order: 0, ParentName: "Dialog"},
		{ID: "1:4", Name: "Generic name", Text: "First item", Depth: 2, Order: 1, ParentName: "Items", LineTypes: []string{"UNORDERED"}},
	}, got)
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
