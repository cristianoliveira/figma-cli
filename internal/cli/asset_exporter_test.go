package cli

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetExporterRecordsSuccessAndFailureInRequestOrder(t *testing.T) {
	downloadServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/broken" {
			http.Error(w, "broken", http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte("svg"))
	}))
	t.Cleanup(downloadServer.Close)

	exporter := AssetExporter{
		HTTPClient: downloadServer.Client(),
		FetchURL: func(nodeID, _ string) (string, error) {
			if nodeID == "3:4" {
				return downloadServer.URL + "/broken", nil
			}
			return downloadServer.URL + "/ok", nil
		},
	}

	manifest := exporter.Export(t.TempDir(), []extract.Asset{
		{ID: "1:2", Name: "Icon", Kind: "vector", Format: "svg"},
		{ID: "3:4", Name: "Photo", Kind: "image", Format: "png"},
	})

	require.Len(t, manifest.Items, 2)
	assert.Equal(t, "1:2", manifest.Items[0].NodeID)
	assert.FileExists(t, manifest.Items[0].Path)
	assert.Empty(t, manifest.Items[0].Error)
	assert.Equal(t, "3:4", manifest.Items[1].NodeID)
	assert.Contains(t, manifest.Items[1].Error, "status 502")
	assert.Equal(t, 1, manifest.Succeeded)
	assert.Equal(t, 1, manifest.Failed)
}

func TestAssetExporterRequiresConfiguredHTTPClient(t *testing.T) {
	exporter := AssetExporter{FetchURL: func(_, _ string) (string, error) { return "https://cdn.example/asset", nil }}

	manifest := exporter.Export(t.TempDir(), []extract.Asset{{ID: "1:2", Name: "Icon", Kind: "vector", Format: "svg"}})

	require.Len(t, manifest.Items, 1)
	assert.Equal(t, "HTTP client is required", manifest.Items[0].Error)
	assert.Equal(t, 1, manifest.Failed)
}

func TestAssetExporterRecordsExportURLFailure(t *testing.T) {
	exporter := AssetExporter{FetchURL: func(_, _ string) (string, error) {
		return "", errors.New("not exportable")
	}}

	manifest := exporter.Export(t.TempDir(), []extract.Asset{{ID: "1:2", Name: "Icon", Format: "svg"}})

	require.Len(t, manifest.Items, 1)
	assert.Equal(t, "not exportable", manifest.Items[0].Error)
	assert.Equal(t, 1, manifest.Failed)
}

func TestAssetExporterResolvesFilenameCollisionsDeterministically(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("svg"))
	}))
	t.Cleanup(server.Close)
	exporter := AssetExporter{
		HTTPClient: server.Client(),
		FetchURL:   func(_, _ string) (string, error) { return server.URL, nil },
		Filename:   func(extract.Asset) string { return "icon" },
	}
	output := t.TempDir()

	manifest := exporter.Export(output, []extract.Asset{
		{ID: "1:2", Name: "Icon", Format: "svg"},
		{ID: "3:4", Name: "Icon", Format: "svg"},
	})

	require.Len(t, manifest.Items, 2)
	assert.Equal(t, filepath.Join(output, "icon.svg"), manifest.Items[0].Path)
	assert.Equal(t, filepath.Join(output, "icon-2.svg"), manifest.Items[1].Path)
}
