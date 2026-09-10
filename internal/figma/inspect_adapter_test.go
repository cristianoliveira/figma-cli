package figma

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInspectAdapterFetchNodeDetailsNoVectorPaths confirms the adapter
// routes to the non-vector path and returns the documents/styles payload.
func TestInspectAdapterFetchNodeDetailsNoVectorPaths(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.NotContains(t, request.URL.RawQuery, "geometry=")
		assert.Equal(t, "/v1/files/FILE/nodes", request.URL.Path)
		body := `{"name":"File","lastModified":"2026-07-10T08:23:53Z","editorType":"figma","thumbnailUrl":"","nodes":{"42:1":{"document":{"id":"42:1","name":"Button","type":"COMPONENT"},"styles":{"S:fill":{"key":"S:fill","name":"Brand/Primary","styleType":"FILL","remote":false,"description":""}}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &Client{HTTP: &http.Client{Transport: transport}}
	adapter := NewInspectAdapter(client)

	details, err := adapter.FetchNodeDetails(context.Background(), "FILE", "42:1", false, "")
	require.NoError(t, err)
	require.Len(t, details.Documents, 1)
	assert.Contains(t, details.Styles, "S:fill")
}

// TestInspectAdapterFetchNodeDetailsVectorPaths confirms the adapter passes
// the bounded depth when vector paths are requested.
func TestInspectAdapterFetchNodeDetailsVectorPaths(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "paths", request.URL.Query().Get("geometry"))
		assert.Equal(t, "3", request.URL.Query().Get("depth"))
		body := `{"name":"File","lastModified":"2026-07-10T08:23:53Z","editorType":"figma","thumbnailUrl":"","nodes":{"42:1":{"document":{"id":"42:1","name":"Wave","type":"VECTOR"}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &Client{HTTP: &http.Client{Transport: transport}}
	adapter := NewInspectAdapter(client)

	_, err := adapter.FetchNodeDetails(context.Background(), "FILE", "42:1", true, "3")
	require.NoError(t, err)
}

// TestInspectAdapterFetchVariables confirms the adapter returns the
// variables metadata without touching transport details.
func TestInspectAdapterFetchVariables(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/files/FILE/variables/local", request.URL.Path)
		body := `{"meta":{"variables":{"V:brand":{"id":"V:brand","name":"Brand","resolvedType":"COLOR"}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &Client{HTTP: &http.Client{Transport: transport}}
	adapter := NewInspectAdapter(client)

	vars, err := adapter.FetchVariables(context.Background(), "FILE")
	require.NoError(t, err)
	require.Contains(t, vars, "variables")
}

// roundTripFunc is a tiny test helper that adapts a function into an
// http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
