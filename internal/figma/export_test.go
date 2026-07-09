package figma

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
)

func TestFetchExportURL(t *testing.T) {
	imageURL := "https://cdn.example.com/asset.png"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := api.GetImagesResponse{Images: map[string]*string{"1:2": &imageURL}}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	got, err := FetchExportURL(client, server.URL, "1:2")

	if err != nil {
		t.Fatalf("FetchExportURL() error = %v", err)
	}
	if got != imageURL {
		t.Errorf("FetchExportURL() = %v, want %v", got, imageURL)
	}
}

func TestFetchExportURLMissingNodeID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(api.GetImagesResponse{Images: map[string]*string{}})
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	if _, err := FetchExportURL(client, server.URL, "1:2"); err == nil {
		t.Fatal("FetchExportURL() expected error for missing node, got nil")
	}
}

func TestFetchExportURLErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	if _, err := FetchExportURL(client, server.URL, "1:2"); err == nil {
		t.Fatal("FetchExportURL() expected error for 403, got nil")
	}
}
