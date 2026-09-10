package extract

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/document"
	"github.com/stretchr/testify/assert"
)

// searchDoc returns the stable document model used by every test in
// this file. It mirrors the JSON fixture the production Figma adapter
// maps from, but the test builds the model directly to avoid a
// figma<->extract import cycle.
func searchDoc() *document.Node {
	return &document.Node{
		ID: "0:0", Name: "root", Type: "FRAME",
		Children: []*document.Node{
			{ID: "1:1", Name: "Login Button", Type: "COMPONENT"},
			{ID: "1:2", Name: "Card", Type: "FRAME", Children: []*document.Node{
				{ID: "1:3", Name: "Signup Button", Type: "INSTANCE"},
				{ID: "1:4", Name: "Icon", Type: "VECTOR"},
			}},
			{ID: "1:5", Name: "Footer", Type: "SECTION"},
			{ID: "1:6", Name: "Stale title", Type: "TEXT", Text: "Actual title"},
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
