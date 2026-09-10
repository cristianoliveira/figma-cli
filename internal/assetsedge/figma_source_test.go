package assetsedge

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestFigmaAssetSourceRequestsExpectedNodePath confirms the Figma edge
// adapter hits /files/{key}/nodes and returns extractable candidates.
func TestFigmaAssetSourceRequestsExpectedNodePath(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/files/FILE/nodes", request.URL.Path)
		assert.Equal(t, "1:2", request.URL.Query().Get("ids"))
		body := `{"nodes":{"1:2":{"document":{"id":"1:2","name":"Screen","type":"FRAME","children":[{"id":"2:3","name":"Close","type":"VECTOR"}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}
	source := NewFigmaAssetSource(client, "FILE", []string{"1:2"})

	candidates, err := source.Candidates(context.Background())
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	assert.Equal(t, "2:3", candidates[0].ID)
	assert.Equal(t, "vector", candidates[0].Kind)
	assert.Equal(t, "svg", candidates[0].Format)
}

// TestFigmaExportURLSourcePreservesFormat confirms the URL adapter
// builds the right export URL for a node/format pair.
func TestFigmaExportURLSourcePreservesFormat(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/images/FILE", request.URL.Path)
		assert.Equal(t, "2:3", request.URL.Query().Get("ids"))
		assert.Equal(t, "png", request.URL.Query().Get("format"))
		body := `{"images":{"2:3":"https://cdn.example/close.png"}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}
	source := NewFigmaExportURLSource(client, "FILE")

	url, err := source.ExportURL(context.Background(), "2:3", "png")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/close.png", url)
}
