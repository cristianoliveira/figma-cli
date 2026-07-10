package figma

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "test-token", r.Header.Get("X-Figma-Token"), "missing auth header")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"key": "value"})
	}))
	defer server.Close()

	client := &Client{Token: "test-token"}
	var got map[string]string
	err := client.Fetch(server.URL, &got)

	require.NoError(t, err)
	assert.Equal(t, "value", got["key"])
}

func TestNewClientConfiguresHTTPClient(t *testing.T) {
	client := NewClient("test-token")

	require.NotNil(t, client.HTTP)
	assert.Equal(t, defaultHTTPTimeout, client.HTTP.Timeout)
}

func TestClientFetchHonorsContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client := NewClient("test-token")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var got map[string]any
	err := client.WithContext(ctx).Fetch(server.URL, &got)

	require.ErrorIs(t, err, context.Canceled)
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

	require.NoError(t, err)
	assert.Equal(t, 1, got["ok"])
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

	require.Error(t, err, "Fetch() expected error for 404")
}

func TestClientFetchJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []any{1, 2, 3}})
	}))
	defer server.Close()

	client := NewClient("test-token")
	got, err := client.FetchJSON(server.URL)

	require.NoError(t, err)
	items, ok := got["items"].([]any)
	require.True(t, ok)
	assert.Len(t, items, 3)
}
