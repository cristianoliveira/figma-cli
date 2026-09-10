package assetsedge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
// to an error carrying the status and body.
func TestHTTPAssetSinkNon200ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("nope"))
	}))
	t.Cleanup(server.Close)

	sink := NewHTTPAssetSink(server.Client())
	err := sink.Write(context.Background(), filepath.Join(t.TempDir(), "asset.svg"), server.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 404")
}

// TestHTTPAssetSinkNilClientErrors confirms a nil client is reported as
// an operational error without panic.
func TestHTTPAssetSinkNilClientErrors(t *testing.T) {
	sink := NewHTTPAssetSink(nil)
	err := sink.Write(context.Background(), "asset.svg", "https://x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP client is required")
}
