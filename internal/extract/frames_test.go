package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscoverFramesReturnsScreenFramesUnderCanvasAndSections(t *testing.T) {
	document := map[string]any{
		"id": "0:1", "name": "Page", "type": "CANVAS",
		"children": []any{
			map[string]any{
				"id": "1:1", "name": "Account", "type": "SECTION",
				"children": []any{
					map[string]any{
						"id": "2:1", "name": "Desktop", "type": "FRAME",
						"children": []any{
							map[string]any{"id": "3:1", "name": "Header", "type": "FRAME"},
						},
					},
				},
			},
			map[string]any{"id": "2:2", "name": "Mobile", "type": "FRAME"},
			map[string]any{"id": "2:3", "name": "Notes", "type": "TEXT"},
		},
	}

	assert.Equal(t, []FrameMatch{
		{ID: "2:1", Name: "Desktop", ParentID: "1:1", ParentName: "Account"},
		{ID: "2:2", Name: "Mobile", ParentID: "0:1", ParentName: "Page"},
	}, DiscoverFrames(document))
}

func TestDiscoverFramesFindsScopedCanvasInsideFileDocument(t *testing.T) {
	document := map[string]any{
		"id": "0:0", "name": "Document", "type": "DOCUMENT",
		"children": []any{
			map[string]any{
				"id": "0:1", "name": "Page", "type": "CANVAS",
				"children": []any{map[string]any{"id": "2:1", "name": "Desktop", "type": "FRAME"}},
			},
		},
	}

	assert.Equal(t, []FrameMatch{{
		ID: "2:1", Name: "Desktop", ParentID: "0:1", ParentName: "Page",
	}}, DiscoverFrames(document))
}

func TestDiscoverFramesReturnsFramesUnderNestedSections(t *testing.T) {
	document := map[string]any{
		"id": "0:1", "name": "Page", "type": "CANVAS",
		"children": []any{
			map[string]any{
				"id": "1:1", "name": "Area", "type": "SECTION",
				"children": []any{
					map[string]any{
						"id": "1:2", "name": "Flow", "type": "SECTION",
						"children": []any{map[string]any{"id": "2:1", "name": "Confirmation", "type": "FRAME"}},
					},
				},
			},
		},
	}

	assert.Equal(t, []FrameMatch{{
		ID: "2:1", Name: "Confirmation", ParentID: "1:2", ParentName: "Flow",
	}}, DiscoverFrames(document))
}

func TestDiscoverFramesReturnsEmptyForInvalidOrFrameScope(t *testing.T) {
	assert.Empty(t, DiscoverFrames(nil))
	assert.Empty(t, DiscoverFrames(map[string]any{"id": "1:1", "name": "Card", "type": "FRAME"}))
}
