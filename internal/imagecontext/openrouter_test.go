package imagecontext

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenRouterDescribeSendsImagesAndPreservesRegionIDs(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	require.NoError(t, os.WriteFile(reference, []byte("reference"), 0o600))
	require.NoError(t, os.WriteFile(actual, []byte("actual"), 0o600))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "model", payload["model"])
		messages := payload["messages"].([]any)
		content := messages[0].(map[string]any)["content"].([]any)
		assert.Len(t, content, 3)
		prompt := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, prompt, "at most 15 words per field")
		assert.Contains(t, prompt, "Do not describe causes")
		assert.Contains(t, prompt, "Do not mention anything outside the supplied region")
		assert.Contains(t, prompt, "User focus: focus on shadows")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"regions\":[{\"region\":\"r1\",\"referenceAppearance\":\"blue rectangle\",\"actualAppearance\":\"blue rectangle\",\"visualContext\":\"lower left\"}]}"}}]}`))
	}))
	defer server.Close()

	explainer := NewOpenRouter("key", "model", server.URL)
	result, err := explainer.Describe(context.Background(), Input{ReferencePath: reference, ActualPath: actual, Regions: []Region{{ID: "r1", Bounds: Bounds{X: 1, Y: 2, Width: 3, Height: 4}}}, Prompt: " focus on shadows "})

	require.NoError(t, err)
	require.Len(t, result.Regions, 1)
	assert.Equal(t, "r1", result.Regions[0].Region)
	assert.Equal(t, "model", result.Model)
	assert.True(t, result.Advisory)
	assert.Equal(t, " focus on shadows ", result.Prompt)
}

func TestOpenRouterDescribeRejectsUnknownRegionID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"regions\":[{\"region\":\"invented\"}]}"}}]}`))
	}))
	defer server.Close()
	dir := t.TempDir()
	path := filepath.Join(dir, "image.png")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o600))

	_, err := NewOpenRouter("key", "model", server.URL).Describe(context.Background(), Input{ReferencePath: path, ActualPath: path, Regions: []Region{{ID: "r1"}}})

	assert.EqualError(t, err, `visual context returned unknown region "invented"`)
}
