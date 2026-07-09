package figma

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMe(t *testing.T) {
	want := api.GetMeResponse{
		Id:     "1",
		Handle: "cristian",
		Email:  "c@x.com",
		ImgUrl: "https://img.example.com/x.png",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	got, err := FetchMe(client, server.URL)

	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestFetchMeErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	client := &Client{Token: "test-token", HTTP: server.Client()}
	_, err := FetchMe(client, server.URL)

	require.Error(t, err, "FetchMe() expected error for 403")
}
