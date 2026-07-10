package diff

import (
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTextHistory struct {
	versions []Version
	text     map[string][]extract.TextNode
}

func (f *fakeTextHistory) Versions() ([]Version, error) { return f.versions, nil }
func (f *fakeTextHistory) TextAt(id string) ([]extract.TextNode, error) {
	return f.text[id], nil
}

func textHistoryFixture(hasNewText map[string]bool) *fakeTextHistory {
	versions := make([]Version, 0, 4)
	text := make(map[string][]extract.TextNode, 4)
	for i, id := range []string{"3", "2", "1", "0"} {
		versions = append(versions, Version{
			ID:        id,
			CreatedAt: time.Date(2026, 7, 10-i, 8, 0, 0, 0, time.UTC),
			User:      "user-" + id,
		})
		value := "old"
		if hasNewText[id] {
			value = "new"
		}
		text[id] = []extract.TextNode{{ID: "4707:15615", Name: "Configure", Text: value}}
	}
	return &fakeTextHistory{versions: versions, text: text}
}

func TestFindTextChangeFindsIntroducingVersion(t *testing.T) {
	got, err := FindTextChange(textHistoryFixture(map[string]bool{"3": true, "2": true}), "3", "")
	require.NoError(t, err)

	assert.Equal(t, "2", got.IntroducedIn.ID)
	require.NotNil(t, got.Previous)
	assert.Equal(t, "1", got.Previous.ID)
	assert.Len(t, got.Changes.Changed, 1)
	assert.Equal(t, "old", got.Changes.Changed[0].From)
	assert.Equal(t, "new", got.Changes.Changed[0].To)
	assert.False(t, got.PredatesHistory)
}

func TestFindTextChangePredatesHistory(t *testing.T) {
	got, err := FindTextChange(textHistoryFixture(map[string]bool{"3": true, "2": true, "1": true, "0": true}), "3", "")
	require.NoError(t, err)

	assert.Equal(t, "0", got.IntroducedIn.ID)
	assert.Nil(t, got.Previous)
	assert.True(t, got.PredatesHistory)
}

func TestFindTextChangeRejectsInvalidBounds(t *testing.T) {
	_, err := FindTextChange(textHistoryFixture(map[string]bool{"3": true}), "2", "3")
	require.Error(t, err)
	assert.ErrorContains(t, err, "older than --to")
}
