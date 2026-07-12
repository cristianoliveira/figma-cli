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

func TestOpenAIDescribeUsesResponsesVisionAPI(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "image.png")
	require.NoError(t, os.WriteFile(imagePath, []byte("png"), 0o600))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/responses", r.URL.Path)
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		input := payload["input"].([]any)[0].(map[string]any)["content"].([]any)
		assert.Equal(t, "input_text", input[0].(map[string]any)["type"])
		assert.Equal(t, "input_image", input[1].(map[string]any)["type"])
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"{\"regions\":[{\"region\":\"r1\",\"visualContext\":\"card\"}]}"}]}]}`))
	}))
	defer server.Close()

	result, err := NewOpenAI("key", "model", server.URL).Describe(context.Background(), Input{ReferencePath: imagePath, ActualPath: imagePath, Regions: []Region{{ID: "r1"}}})

	require.NoError(t, err)
	assert.Equal(t, "openai", result.Provider)
	assert.Equal(t, "card", result.Regions[0].VisualContext)
}
