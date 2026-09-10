package figma

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// strPtr is a tiny test helper to build *string without cluttering the cases.
func strPtr(s string) *string { return &s }

func TestMapUserPreservesStableFieldNames(t *testing.T) {
	got := MapUser(api.User{Handle: "Ada", Id: "1", ImgUrl: "https://img/x.png"})

	assert.Equal(t, "Ada", got.Handle)
	assert.Equal(t, "1", got.ID)
	assert.Equal(t, "https://img/x.png", got.ImageURL, "snake_case img_url becomes ImageURL")
}

func TestMapProjectFilesHandlesMissingOptionalThumbnail(t *testing.T) {
	modified := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	got := MapProjectFiles(api.GetProjectFilesResponse{
		Name: "Empty",
		Files: []struct {
			Key          string    `json:"key"`
			LastModified time.Time `json:"last_modified"`
			Name         string    `json:"name"`
			ThumbnailUrl *string   `json:"thumbnail_url,omitempty"`
		}{
			{Key: "abc", Name: "No Thumb", LastModified: modified, ThumbnailUrl: nil},
		},
	})

	require.Len(t, got.Files, 1)
	assert.Equal(t, "abc", got.Files[0].Key)
	assert.Equal(t, "No Thumb", got.Files[0].Name)
	assert.True(t, got.Files[0].LastModified.Equal(modified), "timestamp preserved through mapper")
	assert.Nil(t, got.Files[0].ThumbnailURL, "nil thumbnail remains nil so callers can distinguish missing")
}

func TestMapProjectFilesPreservesThumbnailPointer(t *testing.T) {
	thumb := "https://example.com/thumb.png"
	got := MapProjectFiles(api.GetProjectFilesResponse{
		Files: []struct {
			Key          string    `json:"key"`
			LastModified time.Time `json:"last_modified"`
			Name         string    `json:"name"`
			ThumbnailUrl *string   `json:"thumbnail_url,omitempty"`
		}{
			{Key: "abc", Name: "With Thumb", ThumbnailUrl: &thumb},
		},
	})

	require.Len(t, got.Files, 1)
	require.NotNil(t, got.Files[0].ThumbnailURL)
	assert.Equal(t, "https://example.com/thumb.png", *got.Files[0].ThumbnailURL)
}

func TestMapTeamProjects(t *testing.T) {
	got := MapTeamProjects(api.GetTeamProjectsResponse{
		Name:     "Wire",
		Projects: []api.Project{{Id: "p1", Name: "Alpha"}, {Id: "p2", Name: "Beta"}},
	})

	assert.Equal(t, "Wire", got.Name)
	require.Len(t, got.Projects, 2)
	assert.Equal(t, "p1", got.Projects[0].ID)
	assert.Equal(t, "Alpha", got.Projects[0].Name)
	assert.Equal(t, "p2", got.Projects[1].ID)
}

func TestMapTeamProjectsEmptyResponseYieldsEmptySlice(t *testing.T) {
	got := MapTeamProjects(api.GetTeamProjectsResponse{Name: "Empty"})

	assert.Equal(t, "Empty", got.Name)
	assert.NotNil(t, got.Projects, "empty projects slice so JSON renders [] not null")
	assert.Empty(t, got.Projects)
}

func TestMapFileVersionsCollapsesPaginationToPresenceFlags(t *testing.T) {
	next := "https://api.figma.com/v1/files/x/versions?after=last"
	prev := "https://api.figma.com/v1/files/x/versions?before=first"
	created := time.Date(2026, 7, 10, 8, 23, 53, 0, time.UTC)

	got := MapFileVersions(api.GetFileVersionsResponse{
		Versions: []api.Version{
			{
				Id:          "v1",
				CreatedAt:   created,
				Label:       strPtr("Launch"),
				Description: strPtr("Final"),
				User:        api.User{Handle: "Astrid", Id: "1", ImgUrl: "https://img/a.png"},
			},
		},
		Pagination: api.ResponsePagination{NextPage: &next, PrevPage: &prev},
	})

	require.Len(t, got.Versions, 1)
	assert.Equal(t, "v1", got.Versions[0].ID)
	require.NotNil(t, got.Versions[0].Label)
	assert.Equal(t, "Launch", *got.Versions[0].Label)
	assert.True(t, got.Versions[0].CreatedAt.Equal(created))
	assert.Equal(t, "Astrid", got.Versions[0].User.Handle)
	assert.Equal(t, "1", got.Versions[0].User.ID)
	assert.Equal(t, "https://img/a.png", got.Versions[0].User.ImageURL)
	assert.True(t, got.Pagination.HasNextPage)
	assert.True(t, got.Pagination.HasPrevPage)
}

func TestMapFileVersionsHandlesAutoSaveWithoutLabelOrDescription(t *testing.T) {
	got := MapFileVersions(api.GetFileVersionsResponse{
		Versions: []api.Version{{Id: "auto", User: api.User{Handle: "Wolfgang", Id: "2"}}},
	})

	require.Len(t, got.Versions, 1)
	assert.Nil(t, got.Versions[0].Label)
	assert.Nil(t, got.Versions[0].Description)
	assert.Nil(t, got.Versions[0].ThumbnailURL)
	assert.False(t, got.Pagination.HasNextPage)
	assert.False(t, got.Pagination.HasPrevPage)
}

func TestMapCommentsConvertsResolvedAtToBoolean(t *testing.T) {
	resolved := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	got := MapComments([]api.Comment{
		{Id: "open", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), User: api.User{Handle: "Ada"}},
		{Id: "closed", CreatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), ResolvedAt: &resolved, User: api.User{Handle: "Linus"}},
	})

	require.Len(t, got, 2)
	assert.False(t, got[0].Resolved, "no ResolvedAt -> Resolved false")
	assert.True(t, got[1].Resolved, "ResolvedAt set -> Resolved true")
}

func TestMapCommentsPropagatesParentIDAndClientMeta(t *testing.T) {
	parent := "root"
	var meta api.Comment_ClientMeta
	require.NoError(t, meta.FromFrameOffset(api.FrameOffset{NodeId: "1:2"}))

	got := MapComments([]api.Comment{{
		Id:         "reply",
		ParentId:   &parent,
		CreatedAt:  time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		ClientMeta: meta,
		User:       api.User{Handle: "Ada"},
	}})

	require.Len(t, got, 1)
	require.NotNil(t, got[0].ParentID)
	assert.Equal(t, "root", *got[0].ParentID)
	assert.NotEmpty(t, got[0].ClientMeta, "client meta preserved as raw JSON for downstream decoders")

	var decoded struct {
		NodeID string `json:"node_id"`
	}
	require.NoError(t, json.Unmarshal(got[0].ClientMeta, &decoded))
	assert.Equal(t, "1:2", decoded.NodeID)
}

func TestMapCommentsEmptyClientMetaProducesNilRawMessage(t *testing.T) {
	got := MapComments([]api.Comment{{Id: "no-meta"}})

	require.Len(t, got, 1)
	assert.Nil(t, got[0].ClientMeta, "missing client_meta union becomes nil so callers can detect absence")
}

func TestMapCommentsEmptyInputYieldsEmptySlice(t *testing.T) {
	got := MapComments(nil)

	assert.NotNil(t, got, "nil input -> empty slice so JSON renders [] not null")
	assert.Empty(t, got)
}
