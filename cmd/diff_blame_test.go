package cmd

import (
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func blameAt(id, when, handle string) api.Version {
	t, _ := time.Parse(time.RFC3339, when)
	return api.Version{Id: id, CreatedAt: t, User: api.User{Handle: handle, Id: "u" + id}}
}

func ptrVersion(v api.Version) *api.Version { return &v }

func TestNewBlameOutputMapsFields(t *testing.T) {
	result := figma.BlameResult{
		IntroducedIn: blameAt("2360480101025964654", "2026-06-02T13:02:03Z", "Wolfgang"),
		Previous:     ptrVersion(blameAt("2358538525635412943", "2026-05-28T07:30:15Z", "Wolfgang")),
		Changes: extract.TextOutput{
			Changed: []extract.ChangedTextOutput{
				{ID: "4707:15615", Name: "Configure", From: "old", To: "new"},
			},
		},
	}

	got := newBlameOutput(result)

	assert.Equal(t, "2360480101025964654", got.IntroducedIn.ID)
	assert.Equal(t, "Wolfgang", got.IntroducedIn.User)
	assert.Equal(t, "2026-06-02T13:02:03Z", got.IntroducedIn.CreatedAt)
	require.NotNil(t, got.Previous)
	assert.Equal(t, "2358538525635412943", got.Previous.ID)
	require.Len(t, got.Changes.Changed, 1)
	assert.False(t, got.PredatesHistory)
}

func TestNewBlameOutputNilPrevious(t *testing.T) {
	result := figma.BlameResult{
		IntroducedIn:    blameAt("0", "2026-06-01T00:00:00Z", "Wolfgang"),
		PredatesHistory: true,
	}

	got := newBlameOutput(result)

	assert.Nil(t, got.Previous)
	assert.True(t, got.PredatesHistory)
}

func TestFormatBlameRendersIntroAndChange(t *testing.T) {
	out := newBlameOutput(figma.BlameResult{
		IntroducedIn: blameAt("2360480101025964654", "2026-06-02T13:02:03Z", "Wolfgang"),
		Previous:     ptrVersion(blameAt("2358538525635412943", "2026-05-28T07:30:15Z", "Wolfgang")),
		Changes: extract.TextOutput{
			Changed: []extract.ChangedTextOutput{
				{ID: "4707:15615", Name: "Configure", From: "old", To: "new"},
			},
		},
	})

	s := formatBlame(out)

	assert.Contains(t, s, "Introduced in 2360480101025964654 by Wolfgang on 2026-06-02 13:02:03")
	assert.Contains(t, s, "4707:15615")
	assert.Contains(t, s, `"old" -> "new"`)
}

func TestFormatTextPath(t *testing.T) {
	assert.Equal(t, " [Document/Checkout]", formatTextPath([]string{"Document", "Checkout"}))
	assert.Empty(t, formatTextPath(nil))
}

func TestFormatBlameNotesPredatesHistory(t *testing.T) {
	out := newBlameOutput(figma.BlameResult{
		IntroducedIn:    blameAt("0", "2026-06-01T00:00:00Z", "Wolfgang"),
		PredatesHistory: true,
	})

	s := formatBlame(out)

	assert.Contains(t, s, "oldest searched version")
}
