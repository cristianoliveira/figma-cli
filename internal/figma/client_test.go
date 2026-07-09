package figma

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Figma-Token") != "test-token" {
			t.Errorf("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer server.Close()

	client := &Client{Token: "test-token"}
	var got map[string]string
	err := client.Fetch(server.URL, &got)

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if got["key"] != "value" {
		t.Errorf("Fetch() got = %v, want key=value", got)
	}
}

func TestClientFetchNilHTTPDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{"ok": 1})
	}))
	defer server.Close()

	// Construct Client without calling NewClient — HTTP is nil
	client := &Client{Token: "test-token"}
	var got map[string]int
	err := client.Fetch(server.URL, &got)

	if err != nil {
		t.Fatalf("Fetch() with nil HTTP error = %v", err)
	}
	if got["ok"] != 1 {
		t.Errorf("Fetch() got = %v, want ok=1", got)
	}
}

func TestClientFetchErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	client := NewClient("test-token")
	var got map[string]any
	err := client.Fetch(server.URL, &got)

	if err == nil {
		t.Fatal("Fetch() expected error for 404, got nil")
	}
}

func TestClientFetchJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{1, 2, 3}})
	}))
	defer server.Close()

	client := NewClient("test-token")
	got, err := client.FetchJSON(server.URL)

	if err != nil {
		t.Fatalf("FetchJSON() error = %v", err)
	}
	items, ok := got["items"].([]any)
	if !ok || len(items) != 3 {
		t.Errorf("FetchJSON() got = %v", got)
	}
}
