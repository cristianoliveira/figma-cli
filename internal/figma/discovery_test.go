package figma

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchTeamProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"name":"Wire","projects":[{"id":"123","name":"Design System"}]}`))
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	got, err := FetchTeamProjects(client, server.URL)

	require.NoError(t, err)
	assert.Equal(t, "Wire", got.Name)
	require.Len(t, got.Projects, 1)
	assert.Equal(t, "123", got.Projects[0].Id)
}

func TestFetchTeamProjectsErrorStatus(t *testing.T) {
	server := forbiddenServer(t)
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	_, err := FetchTeamProjects(client, server.URL)

	require.Error(t, err)
	assert.ErrorContains(t, err, "fetching team projects")
}

func TestFetchProjectFiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"name":"Design System","files":[{"key":"abc","name":"Primitives","last_modified":"2026-07-07T12:00:00Z","thumbnail_url":"https://example.com/thumb.png"}]}`))
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	got, err := FetchProjectFiles(client, server.URL)

	require.NoError(t, err)
	assert.Equal(t, "Design System", got.Name)
	require.Len(t, got.Files, 1)
	assert.Equal(t, "abc", got.Files[0].Key)
}

func TestFetchProjectFilesErrorStatus(t *testing.T) {
	server := forbiddenServer(t)
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	_, err := FetchProjectFiles(client, server.URL)

	require.Error(t, err)
	assert.ErrorContains(t, err, "fetching project files")
}

func forbiddenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
}
