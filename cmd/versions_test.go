package cmd

import (
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseTime(t *testing.T, v string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, v)
	require.NoError(t, err)
	return parsed
}

func strPtr(s string) *string { return &s }

func TestNewVersionsOutputMapsFields(t *testing.T) {
	response := api.GetFileVersionsResponse{
		Versions: []api.Version{
			{
				Id:           "v1",
				CreatedAt:    parseTime(t, "2026-07-10T08:23:53Z"),
				Label:        strPtr("Launch"),
				Description:  strPtr("Final pre-launch cut"),
				User:         api.User{Handle: "Astrid Pahl", Id: "1042800039466656116"},
				ThumbnailUrl: strPtr("https://example.com/thumb.png"),
			},
			{
				Id:        "v2",
				CreatedAt: parseTime(t, "2026-07-08T09:24:23Z"),
				Label:     nil,
				User:      api.User{Handle: "Wolfgang", Id: "1028228948377477971"},
			},
		},
		Pagination: api.ResponsePagination{
			NextPage: strPtr("https://api.figma.com/v1/files/x/versions?after=v2"),
			PrevPage: strPtr("https://api.figma.com/v1/files/x/versions?before=v1"),
		},
	}

	got := newVersionsOutput(response)

	require.Len(t, got.Versions, 2)

	first := got.Versions[0]
	assert.Equal(t, "v1", first.ID)
	assert.Equal(t, "2026-07-10T08:23:53Z", first.CreatedAt)
	assert.Equal(t, "Launch", *first.Label)
	assert.Equal(t, "Final pre-launch cut", *first.Description)
	assert.True(t, first.Named, "named version when label present")
	assert.Equal(t, "Astrid Pahl", first.User.Handle)
	assert.Equal(t, "https://example.com/thumb.png", *first.ThumbnailURL)

	second := got.Versions[1]
	assert.False(t, second.Named, "auto-save when label and description nil")
	assert.Nil(t, second.ThumbnailURL)

	assert.True(t, got.Pagination.HasMore)
	assert.Equal(t, "v2", got.Pagination.NextAfter, "nextAfter is oldest id in page")
	assert.True(t, got.Pagination.HasPrev)
	assert.Equal(t, "v1", got.Pagination.PrevBefore, "prevBefore is newest id in page")
}

func TestNewVersionsOutputEmptyVersions(t *testing.T) {
	got := newVersionsOutput(api.GetFileVersionsResponse{})

	assert.Empty(t, got.Versions)
	assert.NotNil(t, got.Versions)
	assert.False(t, got.Pagination.HasMore)
	assert.False(t, got.Pagination.HasPrev)
	assert.Empty(t, got.Pagination.NextAfter)
	assert.Empty(t, got.Pagination.PrevBefore)
}

func TestNewVersionsOutputOmitsCursorsWhenNoPagination(t *testing.T) {
	response := api.GetFileVersionsResponse{
		Versions: []api.Version{{Id: "only", User: api.User{Handle: "Wolfgang"}}},
	}

	got := newVersionsOutput(response)

	assert.False(t, got.Pagination.HasMore)
	assert.False(t, got.Pagination.HasPrev)
	assert.Empty(t, got.Pagination.NextAfter)
	assert.Empty(t, got.Pagination.PrevBefore)
}

func TestFormatVersionsTableRendersHeaderRowsAndPagination(t *testing.T) {
	out := newVersionsOutput(api.GetFileVersionsResponse{
		Versions: []api.Version{
			{Id: "v1", CreatedAt: parseTime(t, "2026-07-10T08:23:53Z"), Label: strPtr("Launch"), User: api.User{Handle: "Astrid Pahl"}},
			{Id: "v2", CreatedAt: parseTime(t, "2026-07-08T09:24:23Z"), User: api.User{Handle: "Wolfgang"}},
		},
		Pagination: api.ResponsePagination{NextPage: strPtr("https://api.figma.com/v1/files/x/versions?after=v2")},
	})

	table := formatVersionsTable(out)

	assert.Contains(t, table, "Created (UTC)")
	assert.Contains(t, table, "Version ID")
	assert.Contains(t, table, "v1")
	assert.Contains(t, table, "Astrid Pahl")
	assert.Contains(t, table, "named")
	assert.Contains(t, table, "Launch")
	assert.Contains(t, table, "v2")
	assert.Contains(t, table, "auto")
	assert.Contains(t, table, "nextAfter=v2")
	assert.Contains(t, table, "hasMore=true")
}

func TestFormatVersionsTableEmpty(t *testing.T) {
	table := formatVersionsTable(versionsOutput{})

	assert.Contains(t, table, "No versions")
}
