package figma

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchDocument(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/FILE" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"document": map[string]any{"id": "0:0", "name": "Root", "type": "DOCUMENT"},
		})
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	client := NewClient("test-token")
	doc, err := FetchDocument(client, "FILE", nil, "", "")
	if err != nil {
		t.Fatalf("FetchDocument() error = %v", err)
	}
	name := doc.(map[string]any)["name"]
	if name != "Root" {
		t.Errorf("FetchDocument() doc name = %v, want Root", name)
	}
}

func TestFetchDocument_NodeIDsVersionDepth(t *testing.T) {
	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode(map[string]any{
			"document": map[string]any{"id": "0:0", "name": "Node", "type": "FRAME"},
		})
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	client := NewClient("test-token")
	if _, err := FetchDocument(client, "FILE", []string{"1:2"}, "v9", "2"); err != nil {
		t.Fatalf("FetchDocument() error = %v", err)
	}
	for _, want := range []string{"ids=1%3A2", "version=v9", "depth=2"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query missing %q: %s", want, gotQuery)
		}
	}
}

func TestFetchDocument_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	withBaseURL(t, server.URL)

	client := NewClient("test-token")
	if _, err := FetchDocument(client, "FILE", nil, "", ""); err == nil {
		t.Fatal("FetchDocument() expected error for 500, got nil")
	}
}

// withBaseURL swaps the package baseURL for the duration of the test.
func withBaseURL(t *testing.T, base string) {
	t.Helper()
	orig := baseURL
	baseURL = base
	t.Cleanup(func() { baseURL = orig })
}
