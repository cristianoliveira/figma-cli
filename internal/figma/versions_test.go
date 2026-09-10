package figma

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
		  "versions": [
		    {
		      "id": "2374505616843859677",
		      "created_at": "2026-07-10T08:23:53Z",
		      "label": "Launch",
		      "description": "Final pre-launch cut",
		      "user": {"handle": "Astrid Pahl", "id": "1042800039466656116", "img_url": "https://example.com/a.png"},
		      "thumbnail_url": "https://example.com/thumb.png"
		    }
		  ],
		  "pagination": {"next_page": "https://api.figma.com/v1/files/x/versions?after=1"}
		}`))
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	got, err := FetchVersions(client, server.URL)

	require.NoError(t, err)
	require.Len(t, got.Versions, 1)
	assert.Equal(t, "2374505616843859677", got.Versions[0].ID)
	require.NotNil(t, got.Versions[0].Label)
	assert.Equal(t, "Launch", *got.Versions[0].Label)
	assert.True(t, got.Pagination.HasNextPage)
	assert.False(t, got.Pagination.HasPrevPage)
	assert.Equal(t, "Astrid Pahl", got.Versions[0].User.Handle)
	assert.Equal(t, "1042800039466656116", got.Versions[0].User.ID)
}

func TestFetchVersionsErrorStatus(t *testing.T) {
	server := forbiddenServer(t)
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	_, err := FetchVersions(client, server.URL)

	require.Error(t, err)
	assert.ErrorContains(t, err, "fetching versions")
}
