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

// TestTokensAdaptersVariablesFetcher confirms the adapter routes the
// Variables fetcher to /files/{key}/variables/local.
func TestTokensAdaptersVariablesFetcher(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/files/FILE/variables/local", request.URL.Path)
		body := `{"meta":{"variables":{}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &Client{HTTP: &http.Client{Transport: transport}}
	adapter := NewTokensAdapters(client)

	meta, err := adapter.VariablesFetcher()(context.Background(), "FILE")
	require.NoError(t, err)
	require.Contains(t, meta, "variables")
}

// TestTokensAdaptersStylesFetcher confirms the Styles fetcher routes to
// /files/{key}/styles.
func TestTokensAdaptersStylesFetcher(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/files/FILE/styles", request.URL.Path)
		body := `{"meta":{"styles":[]}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &Client{HTTP: &http.Client{Transport: transport}}
	adapter := NewTokensAdapters(client)

	styles, err := adapter.StylesFetcher()(context.Background(), "FILE")
	require.NoError(t, err)
	assert.Empty(t, styles)
}

// TestTokensAdaptersNodesFetcher confirms the Nodes fetcher routes to
// /files/{key}/nodes with the requested IDs.
func TestTokensAdaptersNodesFetcher(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/files/FILE/nodes", request.URL.Path)
		assert.Equal(t, "1:2,3:4", request.URL.Query().Get("ids"))
		body := `{"nodes":{}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &Client{HTTP: &http.Client{Transport: transport}}
	adapter := NewTokensAdapters(client)

	nodes, err := adapter.NodesFetcher()(context.Background(), "FILE", []string{"1:2", "3:4"})
	require.NoError(t, err)
	require.NotNil(t, nodes)
}
