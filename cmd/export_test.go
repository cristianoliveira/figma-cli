package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

func TestBuildExportURL(t *testing.T) {
	got, err := figma.BuildExportURL("file123", []string{"1:2"}, "png")
	if err != nil {
		t.Fatalf("BuildExportURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/images/file123?format=png&ids=1%3A2"
	if got != expected {
		t.Fatalf("BuildExportURL() = %v, expected %v", got, expected)
	}
}

func TestBuildExportURLRequiresNodeID(t *testing.T) {
	_, err := figma.BuildExportURL("file123", nil, "png")
	if err == nil {
		t.Fatal("BuildExportURL() error = nil, expected error")
	}
}

func TestBuildExportURLRequiresFormat(t *testing.T) {
	_, err := figma.BuildExportURL("file123", []string{"1:2"}, "")
	if err == nil {
		t.Fatal("BuildExportURL() error = nil, expected error")
	}
}

func TestDefaultExportOutputPath(t *testing.T) {
	got := defaultExportOutputPath("file123", "1:2", "png")
	expected := "file123_1-2.png"
	if got != expected {
		t.Fatalf("defaultExportOutputPath() = %v, expected %v", got, expected)
	}
}

func TestFetchExportURL(t *testing.T) {
	imageURL := "https://cdn.example.com/asset.png"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := api.GetImagesResponse{Images: map[string]*string{"1:2": &imageURL}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &figma.Client{Token: "test-token", HTTP: server.Client()}
	got, err := fetchExportURL(client, server.URL, "1:2")

	if err != nil {
		t.Fatalf("fetchExportURL() error = %v", err)
	}
	if got != "https://cdn.example.com/asset.png" {
		t.Errorf("fetchExportURL() = %v, want https://cdn.example.com/asset.png", got)
	}
}

func TestFetchExportURLMissingNodeID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(api.GetImagesResponse{Images: map[string]*string{}})
	}))
	defer server.Close()

	client := &figma.Client{Token: "test-token", HTTP: server.Client()}
	_, err := fetchExportURL(client, server.URL, "1:2")

	if err == nil {
		t.Fatal("fetchExportURL() expected error for missing node, got nil")
	}
}

func TestFetchExportURLErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	client := &figma.Client{Token: "test-token", HTTP: server.Client()}
	_, err := fetchExportURL(client, server.URL, "1:2")

	if err == nil {
		t.Fatal("fetchExportURL() expected error for 403, got nil")
	}
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fake-image-data"))
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := downloadFile(server.Client(), outputPath, server.URL)

	if err != nil {
		t.Fatalf("downloadFile() error = %v", err)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if string(data) != "fake-image-data" {
		t.Errorf("downloadFile() wrote %q, want fake-image-data", string(data))
	}
}

func TestDownloadFileErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "export.png")
	err := downloadFile(server.Client(), outputPath, server.URL)

	if err == nil {
		t.Fatal("downloadFile() expected error for 404, got nil")
	}
}
