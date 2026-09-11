package assetsedge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/operr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPAssetSinkWritesBytes confirms the edge sink performs a full
// download + file write and returns no error on 200.
func TestHTTPAssetSinkWritesBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-image"))
	}))
	t.Cleanup(server.Close)

	sink := NewHTTPAssetSink(server.Client())
	outPath := filepath.Join(t.TempDir(), "out", "asset.svg")

	err := sink.Write(context.Background(), outPath, server.URL)
	require.NoError(t, err)
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Equal(t, "fake-image", string(data))
}

// TestHTTPAssetSinkCreatesParentDirectory confirms the sink creates the
// destination directory.
func TestHTTPAssetSinkCreatesParentDirectory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("x"))
	}))
	t.Cleanup(server.Close)

	sink := NewHTTPAssetSink(server.Client())
	outPath := filepath.Join(t.TempDir(), "a", "b", "asset.svg")

	err := sink.Write(context.Background(), outPath, server.URL)
	require.NoError(t, err)
	assert.FileExists(t, outPath)
}

// TestHTTPAssetSinkNon200ReturnsError confirms a non-200 response maps
// to a classified error carrying the safe status code and never the body.
func TestHTTPAssetSinkNon200ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("SECRET-RESPONSE-BODY"))
	}))
	t.Cleanup(server.Close)

	sink := NewHTTPAssetSink(server.Client())
	err := sink.Write(context.Background(), filepath.Join(t.TempDir(), "asset.svg"), server.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
	assert.NotContains(t, err.Error(), "SECRET-RESPONSE-BODY", "response body must not leak")

	var classified *operr.ClassifiedError
	require.ErrorAs(t, err, &classified)
	assert.Equal(t, operr.CategoryDependencyUnavailable, classified.Category)
}

// TestHTTPAssetSinkNilClientErrors confirms a nil client is reported as
// an operational error without panic.
func TestHTTPAssetSinkNilClientErrors(t *testing.T) {
	sink := NewHTTPAssetSink(nil)
	err := sink.Write(context.Background(), "asset.svg", "https://x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client is required")
}
