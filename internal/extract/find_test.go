package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func searchDoc() map[string]any {
	return map[string]any{
		"id": "0:0", "name": "root", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Login Button", "type": "COMPONENT"},
			map[string]any{"id": "1:2", "name": "Card", "type": "FRAME", "children": []any{
				map[string]any{"id": "1:3", "name": "Signup Button", "type": "INSTANCE"},
				map[string]any{"id": "1:4", "name": "Icon", "type": "VECTOR"},
			}},
			map[string]any{"id": "1:5", "name": "Footer", "type": "SECTION"},
			map[string]any{"id": "1:6", "name": "Stale title", "type": "TEXT", "characters": "Actual title"},
		},
	}
}

func TestSearchByName(t *testing.T) {
	matches := Search(searchDoc(), SearchCriteria{Name: "button"})

	assert.Len(t, matches, 2)
	assert.Equal(t, "1:1", matches[0].ID)
	assert.Equal(t, "1:3", matches[1].ID)
}

func TestSearchByType(t *testing.T) {
	matches := Search(searchDoc(), SearchCriteria{Type: "frame"})

	assert.Len(t, matches, 2)
	assert.Equal(t, "0:0", matches[0].ID)
	assert.Equal(t, "1:2", matches[1].ID)
}

func TestSearchByNameAndType(t *testing.T) {
	matches := Search(searchDoc(), SearchCriteria{Name: "button", Type: "COMPONENT"})

	assert.Len(t, matches, 1)
	assert.Equal(t, "1:1", matches[0].ID)
}

func TestSearchNoMatch(t *testing.T) {
	assert.Empty(t, Search(searchDoc(), SearchCriteria{Name: "nonexistent"}))
	assert.Empty(t, Search(searchDoc(), SearchCriteria{Type: "WIDGET"}))
}

func TestSearchNil(t *testing.T) {
	assert.Nil(t, Search(nil, SearchCriteria{Name: "x"}))
}

func TestSearchNoCriteriaReturnsAll(t *testing.T) {
	matches := Search(searchDoc(), SearchCriteria{})

	// root + 6 descendants = 7 nodes.
	assert.Len(t, matches, 7)
}

func TestSearchIncludesCharactersForTextNodes(t *testing.T) {
	matches := Search(searchDoc(), SearchCriteria{Type: "TEXT"})

	assert.Equal(t, []LayerMatch{{
		ID: "1:6", Name: "Stale title", Type: "TEXT", Text: "Actual title",
	}}, matches)
}

func TestSearchLeavesTextEmptyForNonTextNodes(t *testing.T) {
	matches := Search(searchDoc(), SearchCriteria{Name: "Footer"})

	assert.Equal(t, "", matches[0].Text)
}
