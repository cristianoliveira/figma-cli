package figma

import (
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBlameFetcher serves a canned version list (newest-first) and text per
// version, so the binary search can be exercised without HTTP or the generated
// document types.
type fakeBlameFetcher struct {
	versions []api.Version
	text     map[string][]extract.TextNode
}

func (f *fakeBlameFetcher) Versions() ([]api.Version, error) { return f.versions, nil }
func (f *fakeBlameFetcher) TextAt(id string) ([]extract.TextNode, error) {
	return f.text[id], nil
}

// blameFixture builds a 4-version history (newest-first: 3,2,1,0) where each
// version's text is "new" or "old" per mapNew.
func blameFixture(mapNew map[string]bool) *fakeBlameFetcher {
	versions := make([]api.Version, 0, 4)
	text := make(map[string][]extract.TextNode, 4)
	for i, id := range []string{"3", "2", "1", "0"} {
		versions = append(versions, api.Version{
			Id:        id,
			CreatedAt: time.Date(2026, 7, 10-i, 8, 0, 0, 0, time.UTC),
			User:      api.User{Handle: "user-" + id, Id: "u" + id},
		})
		chars := "old"
		if mapNew[id] {
			chars = "new"
		}
		text[id] = []extract.TextNode{{ID: "4707:15615", Name: "Configure", Text: chars}}
	}
	return &fakeBlameFetcher{versions: versions, text: text}
}

func TestFindTextChangeFindsIntroducingVersion(t *testing.T) {
	// versions 3,2 have "new"; 1,0 have "old" -> introduced at 2
	got, err := findTextChange(blameFixture(map[string]bool{"3": true, "2": true}), "3", "")
	require.NoError(t, err)

	assert.Equal(t, "2", got.IntroducedIn.Id, "oldest version with new text")
	require.NotNil(t, got.Previous)
	assert.Equal(t, "1", got.Previous.Id)
	require.Len(t, got.Changes.Changed, 1)
	assert.Equal(t, "old", got.Changes.Changed[0].From)
	assert.Equal(t, "new", got.Changes.Changed[0].To)
	assert.False(t, got.PredatesHistory)
}

func TestFindTextChangeWhenIntroducedAtToItself(t *testing.T) {
	// only the target version has new text -> introduced at to itself
	got, err := findTextChange(blameFixture(map[string]bool{"3": true}), "3", "")
	require.NoError(t, err)

	assert.Equal(t, "3", got.IntroducedIn.Id)
	require.NotNil(t, got.Previous)
	assert.Equal(t, "2", got.Previous.Id)
}

func TestFindTextChangePredatesHistory(t *testing.T) {
	// every version already has the new text -> predates visible history
	got, err := findTextChange(blameFixture(map[string]bool{"3": true, "2": true, "1": true, "0": true}), "3", "")
	require.NoError(t, err)

	assert.Equal(t, "0", got.IntroducedIn.Id, "oldest visible version")
	assert.Nil(t, got.Previous)
	assert.True(t, got.PredatesHistory)
}

func TestFindTextChangeFromBoundPredatesWhenFromHasNewText(t *testing.T) {
	// 3,2 new; 1,0 old. True introducer is 2, but --from=2 caps the search range
	// to [3,2]; since 2 already has the new text, the change is reported as
	// predating the --from bound (at or before 2).
	got, err := findTextChange(blameFixture(map[string]bool{"3": true, "2": true}), "3", "2")
	require.NoError(t, err)

	assert.Equal(t, "2", got.IntroducedIn.Id)
	assert.True(t, got.PredatesHistory)
}

func TestFindTextChangeUnknownToVersion(t *testing.T) {
	_, err := findTextChange(blameFixture(map[string]bool{"3": true}), "nope", "")
	require.Error(t, err)
	assert.ErrorContains(t, err, "not found")
}

func TestFindTextChangeFromOlderThanTo(t *testing.T) {
	_, err := findTextChange(blameFixture(map[string]bool{"3": true}), "2", "3")
	require.Error(t, err)
	assert.ErrorContains(t, err, "older than --to")
}
