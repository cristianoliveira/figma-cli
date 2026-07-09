package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultExportOutputPath(t *testing.T) {
	got := DefaultExportOutputPath("file123", "1:2", "png")
	expected := "file123_1-2.png"
	if got != expected {
		t.Fatalf("DefaultExportOutputPath() = %v, expected %v", got, expected)
	}
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := DownloadFile(server.Client(), outputPath, server.URL)

	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if string(data) != "fake-image-data" {
		t.Errorf("DownloadFile() wrote %q, want fake-image-data", string(data))
	}
}

func TestDownloadFileErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := DownloadFile(server.Client(), outputPath, server.URL)

	if err == nil {
		t.Fatal("DownloadFile() expected error for 404, got nil")
	}
}
