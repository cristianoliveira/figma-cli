package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSON_AlwaysMarshalsWithTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(&buf, false).JSON(map[string]int{"a": 1}))

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got), "not valid JSON: %s", buf.String())
	assert.Equal(t, 1, got["a"])
	assert.True(t, bytes.HasSuffix(buf.Bytes(), []byte("\n")), "missing trailing newline: %q", buf.String())
}

func TestJSON_UnaffectedByJSONFlag(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	// asJSON=true must not wrap a JSON result.
	require.NoError(t, New(&buf, true).JSON(map[string]int{"a": 1}))

	var got map[string]int
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got), "asJSON wrapped a JSON result: %s", buf.String())
}

func TestText_RawByDefault(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(&buf, false).Text("css", ".a { color: red; }"))
	assert.Equal(t, ".a { color: red; }", buf.String())
}

func TestText_WrappedUnderJSON(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(&buf, true).Text("css", ".a { color: red; }"))

	var got map[string]string
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got), "not wrapped JSON: %s", buf.String())
	assert.Equal(t, ".a { color: red; }", got["css"])
}

func TestFile_RawPrintsPathLine(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, New(&buf, false).File("/tmp/a.svg", nil))
	assert.Equal(t, "/tmp/a.svg\n", buf.String())
}

func TestFile_JSONMergesExtra(t *testing.T) {
	var buf bytes.Buffer
	extra := map[string]any{"format": "svg", "node": "1:2"}
	require.NoError(t, New(&buf, true).File("/tmp/a.svg", extra))

	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got), "not JSON: %s", buf.String())
	assert.Equal(t, "/tmp/a.svg", got["path"])
	assert.Equal(t, "svg", got["format"])
	assert.Equal(t, "1:2", got["node"])
}
