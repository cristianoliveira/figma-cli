package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	toon "github.com/toon-format/toon-go"
)

func TestStructuredDefaultsToDeterministicTOONAndRoundTrips(t *testing.T) {
	value := map[string]any{
		"count":   2,
		"empty":   []string{},
		"nested":  map[string]any{"message": "line one\nline two"},
		"results": []map[string]any{{"id": 1, "name": "Ada"}, {"id": 2, "name": "Linus"}},
	}
	var first bytes.Buffer
	var second bytes.Buffer
	require.NoError(t, New(&first, FormatTOON).Structured(value))
	require.NoError(t, New(&second, FormatTOON).Structured(value))
	assert.Equal(t, first.String(), second.String())
	assert.Contains(t, first.String(), "results[2]{id,name}:")
	assert.Contains(t, first.String(), "empty[0]:")
	assert.True(t, bytes.HasSuffix(first.Bytes(), []byte("\n")))

	var decoded any
	require.NoError(t, toon.Unmarshal(first.Bytes(), &decoded))
	var original any
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(encoded, &original))
	assert.Equal(t, original, decoded)
}

func TestStructuredJSONPreservesCompatibilityBytes(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(&buf, FormatJSON).Structured(map[string]int{"a": 1}))
	assert.Equal(t, "{\n  \"a\": 1\n}\n", buf.String())
}

func TestTextRawByDefaultAndWrappedUnderJSON(t *testing.T) {
	var toonOutput bytes.Buffer
	require.NoError(t, New(&toonOutput, FormatTOON).Text("css", ".a { color: red; }"))
	assert.Equal(t, ".a { color: red; }", toonOutput.String())

	var jsonOutput bytes.Buffer
	require.NoError(t, New(&jsonOutput, FormatJSON).Text("css", ".a { color: red; }"))
	assert.JSONEq(t, `{"css":".a { color: red; }"}`, jsonOutput.String())
}

func TestFileRawByDefaultAndWrappedUnderJSON(t *testing.T) {
	var toonOutput bytes.Buffer
	require.NoError(t, New(&toonOutput, FormatTOON).File("/tmp/a.svg", nil))
	assert.Equal(t, "/tmp/a.svg\n", toonOutput.String())

	var jsonOutput bytes.Buffer
	require.NoError(t, New(&jsonOutput, FormatJSON).File("/tmp/a.svg", map[string]any{"format": "svg"}))
	assert.JSONEq(t, `{"path":"/tmp/a.svg","format":"svg"}`, jsonOutput.String())
}

func TestRenderUsesSelectedStructuredFormat(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(&buf, FormatTOON).Render(map[string]int{"a": 1}, "ignored human view"))
	assert.Equal(t, "a: 1\n", buf.String())
}
