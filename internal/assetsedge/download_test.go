package assetsedge

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	t.Cleanup(server.Close)

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := DownloadFile(server.Client(), outputPath, server.URL)

	require.NoError(t, err)
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "fake-image-data", string(data))
}

func TestDownloadFileCreatesMissingParentDirectories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	t.Cleanup(server.Close)

	outputPath := filepath.Join(t.TempDir(), "missing", "nested", "export.png")
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
	t.Cleanup(server.Close)

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := DownloadFile(server.Client(), outputPath, server.URL)

	require.Error(t, err, "DownloadFile() expected error for 404")
}
