package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
)

func TestHasTextChangesEmpty(t *testing.T) {
	assert.False(t, hasTextChanges(extract.TextOutput{}))
}

func TestHasTextChangesDetectsChanged(t *testing.T) {
	assert.True(t, hasTextChanges(extract.TextOutput{
		Changed: []extract.ChangedTextOutput{{ID: "1:1", From: "a", To: "b"}},
	}))
}

func TestHasTextChangesDetectsAdded(t *testing.T) {
	assert.True(t, hasTextChanges(extract.TextOutput{
		Added: []extract.TextNodeOutput{{ID: "1:2", Text: "new"}},
	}))
}

func TestHasTextChangesDetectsRemoved(t *testing.T) {
	assert.True(t, hasTextChanges(extract.TextOutput{
		Removed: []extract.TextNodeOutput{{ID: "1:3", Text: "gone"}},
	}))
}
