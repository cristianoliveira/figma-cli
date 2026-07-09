package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffText(t *testing.T) {
	from := []TextNode{
		{ID: "1:1", Name: "Title", Text: "Old title"},
		{ID: "1:2", Name: "Removed", Text: "Gone"},
		{ID: "1:3", Name: "Same", Text: "Keep"},
	}
	to := []TextNode{
		{ID: "1:1", Name: "Title", Text: "New title"},
		{ID: "1:3", Name: "Same", Text: "Keep"},
		{ID: "1:4", Name: "Added", Text: "New"},
	}

	got := DiffText(from, to)

	assert.Len(t, got.Changed, 1)
	assert.Equal(t, "Old title", got.Changed[0].From)
	assert.Equal(t, "New title", got.Changed[0].To)

	assert.Len(t, got.Removed, 1)
	assert.Equal(t, "Gone", got.Removed[0].Text)

	assert.Len(t, got.Added, 1)
	assert.Equal(t, "New", got.Added[0].Text)
}
