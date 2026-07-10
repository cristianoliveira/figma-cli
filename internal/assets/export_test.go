package assets

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultExportOutputPath(t *testing.T) {
	got := DefaultExportOutputPath("file123", "1:2", "png")
	assert.Equal(t, "file123_1-2.png", got)
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := DownloadFile(server.Client(), outputPath, server.URL)

	require.NoError(t, err)
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "fake-image-data", string(data))
}

func TestDownloadFileErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := DownloadFile(server.Client(), outputPath, server.URL)

	require.Error(t, err, "DownloadFile() expected error for 404")
}
