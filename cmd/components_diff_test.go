package cmd

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentsDiffComparesPublishedFigmaComponentsWithCodebase(t *testing.T) {
	codebase := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(codebase, "Button"), 0o755))
	require.NoError(t, os.Mkdir(filepath.Join(codebase, "LegacyButton"), 0o755))
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "/v1/files/abc/components", request.URL.Path)
		body := `{"error":false,"status":200,"meta":{"components":[{"name":"Button / Primary","node_id":"1:1"},{"name":"Avatar Group","node_id":"1:2"}]}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}

	result := executeCommand(newComponentsCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--diff", "--codebase", codebase)

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{
		"fileKey":"abc",
		"figmaCount":2,
		"codeCount":2,
		"matched":[{"figma":{"name":"Button / Primary","nodeId":"1:1"},"code":{"name":"Button","path":"`+filepath.Join(codebase, "Button")+`"}}],
		"missing":[{"name":"Avatar Group","nodeId":"1:2"}],
		"extra":[{"name":"LegacyButton","path":"`+filepath.Join(codebase, "LegacyButton")+`"}]
	}`, result.Stdout)
}

func TestComponentsDiffRejectsUnreadableCodebaseBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newComponentsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--diff", "--codebase", filepath.Join(t.TempDir(), "missing"))

	assert.ErrorContains(t, result.Err, "reading codebase")
	assert.False(t, loaded)
}

func TestComponentsDiffRejectsNodeScopedURLBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newComponentsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "https://www.figma.com/design/abc/Name?node-id=1-1", "--diff", "--codebase", t.TempDir())

	assert.EqualError(t, result.Err, "--diff compares a whole Figma file; remove node-id from the URL")
	assert.False(t, loaded)
}

func TestComponentsDiffRequiresCodebaseBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newComponentsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--diff")

	assert.EqualError(t, result.Err, "--diff requires --codebase")
	assert.False(t, loaded)
}
