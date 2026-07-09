package figma

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubFigma serves canned JSON for the endpoints tokens use, so the whole
// fetch → decode → map round-trip can be exercised without the network.
func stubFigma(t *testing.T, varsBody, stylesBody, nodesBody string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/files/file123/variables/local", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(varsBody))
	})
	mux.HandleFunc("/v1/files/file123/styles", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(stylesBody))
	})
	mux.HandleFunc("/v1/files/file123/nodes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(nodesBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchVariablesRoundTrip(t *testing.T) {
	body := `{"meta":{"variableCollections":{"col-1":{"id":"col-1","name":"C","defaultModeId":"m1","modes":[{"modeId":"m1","name":"Light"}],"variableIds":["v1"]}},"variables":{"v1":{"id":"v1","name":"Color/Red","resolvedType":"COLOR","variableCollectionId":"col-1","valuesByMode":{"m1":{"r":1,"g":0,"b":0,"a":1}}}}},"status":200,"error":false}`
	srv := stubFigma(t, body, "", "")
	old := baseURL
	baseURL = srv.URL + "/v1"
	t.Cleanup(func() { baseURL = old })

	meta, err := FetchVariables(NewClient("token"), "file123")
	require.NoError(t, err)
	vars, _ := meta["variables"].(map[string]any)
	require.Len(t, vars, 1)
	v := vars["v1"].(map[string]any)
	assert.Equal(t, "Color/Red", v["name"])
}

func TestFetchStylesAndNodesRoundTrip(t *testing.T) {
	stylesBody := `{"meta":{"styles":[{"key":"s1","name":"Brand","style_type":"FILL","node_id":"1:1"}]},"status":200,"error":false}`
	nodesBody := `{"nodes":{"1:1":{"document":{"type":"RECTANGLE","fills":[{"type":"SOLID","color":{"r":0,"g":0,"b":0,"a":1}}]}}}}`
	srv := stubFigma(t, "", stylesBody, nodesBody)
	old := baseURL
	baseURL = srv.URL + "/v1"
	t.Cleanup(func() { baseURL = old })

	c := NewClient("token")
	styles, err := FetchStyles(c, "file123")
	require.NoError(t, err)
	require.Len(t, styles, 1)
	assert.Equal(t, "Brand", styles[0]["name"])

	nodes, err := FetchNodes(c, "file123", []string{"1:1"})
	require.NoError(t, err)
	entry := nodes["1:1"].(map[string]any)
	assert.NotNil(t, entry["document"])
}

// TestTokensStylesEndToEnd proves URL → HTTP → typed decode → toMap → extract
// produces a real color token end to end.
func TestTokensStylesEndToEnd(t *testing.T) {
	stylesBody := `{"meta":{"styles":[{"key":"s1","name":"Brand","style_type":"FILL","node_id":"1:1"}]},"status":200,"error":false}`
	nodesBody := `{"nodes":{"1:1":{"document":{"type":"RECTANGLE","fills":[{"type":"SOLID","color":{"r":1,"g":0,"b":0,"a":1}}]}}}}`
	srv := stubFigma(t, "", stylesBody, nodesBody)
	old := baseURL
	baseURL = srv.URL + "/v1"
	t.Cleanup(func() { baseURL = old })

	c := NewClient("token")
	styles, err := FetchStyles(c, "file123")
	require.NoError(t, err)
	nodes, err := FetchNodes(c, "file123", []string{"1:1"})
	require.NoError(t, err)

	tokens := extract.ExtractTokensFromStyles(styles, nodes)
	var got string
	for _, tk := range tokens {
		if tk.Category == "color" {
			got = tk.Value
		}
	}
	assert.Equal(t, "#FF0000", got)
}
