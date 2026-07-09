package figma

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/files/FILE", r.URL.Path, "unexpected path")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"document": map[string]any{"id": "0:0", "name": "Root", "type": "DOCUMENT"},
		})
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	client := NewClient("test-token")
	doc, err := FetchDocument(client, "FILE", nil, "", "")
	require.NoError(t, err)
	name := doc.(map[string]any)["name"]
	assert.Equal(t, "Root", name)
}

func TestFetchDocument_NodeIDsVersionDepth(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"document": map[string]any{"id": "0:0", "name": "Node", "type": "FRAME"},
		})
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	client := NewClient("test-token")
	_, err := FetchDocument(client, "FILE", []string{"1:2"}, "v9", "2")
	require.NoError(t, err)
	for _, want := range []string{"ids=1%3A2", "version=v9", "depth=2"} {
		assert.Contains(t, gotQuery, want)
	}
}

func TestFetchNodeDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/files/FILE/nodes", r.URL.Path)
		assert.Equal(t, "1:2,3:4", r.URL.Query().Get("ids"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"nodes": map[string]any{
				"1:2": map[string]any{"document": map[string]any{"id": "1:2", "name": "First", "type": "FRAME"}},
				"3:4": map[string]any{"document": map[string]any{"id": "3:4", "name": "Second", "type": "FRAME"}},
			},
		})
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	docs, err := FetchNodeDocuments(NewClient("test-token"), "FILE", []string{"1:2", "3:4"})

	require.NoError(t, err)
	require.Len(t, docs, 2)
	assert.Equal(t, "First", docs[0].(map[string]any)["name"])
	assert.Equal(t, "Second", docs[1].(map[string]any)["name"])
}

func TestFetchNodeDocumentsRequiresNodeIDs(t *testing.T) {
	_, err := FetchNodeDocuments(NewClient("test-token"), "FILE", nil)

	require.Error(t, err)
}

func TestFetchNodeDocumentsMissingNode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"nodes":{"1:2":null}}`))
	}))
	defer server.Close()
	withBaseURL(t, server.URL)

	_, err := FetchNodeDocuments(NewClient("test-token"), "FILE", []string{"1:2"})

	require.Error(t, err)
	assert.ErrorContains(t, err, "1:2")
}

func TestFetchDocument_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	client := NewClient("test-token")
	_, err := FetchDocument(client, "FILE", nil, "", "")
	require.Error(t, err, "FetchDocument() expected error for 500")
}

// withBaseURL swaps the package baseURL for the duration of the test.
func withBaseURL(t *testing.T, base string) {
	t.Helper()
	orig := baseURL
	baseURL = base
	t.Cleanup(func() { baseURL = orig })
}
