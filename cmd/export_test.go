package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportCommandWritesFileAndJSONContract(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "image"
		if strings.Contains(request.URL.Path, "/v1/images/") {
			body = `{"images":{"42:1":"https://cdn.example/image"}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	httpClient := &http.Client{Transport: transport}
	client := &figma.Client{HTTP: httpClient}
	outputPath := t.TempDir() + "/button.svg"

	result := executeCommand(
		newExportCommand(func() (*figma.Client, error) { return client, nil }, httpClient),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--output", outputPath, "--json",
	)

	require.NoError(t, result.Err)
	expectedJSON, err := json.Marshal(map[string]string{
		"path":   outputPath,
		"format": "svg",
		"node":   "42:1",
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), result.Stdout)
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "image", string(content))
}

func TestExportCommandRejectsFormatBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}, http.DefaultClient), "abc", "--id", "1:2", "--format", "gif")

	assert.EqualError(t, result.Err, `invalid format "gif": expected png, jpg, svg, or pdf`)
	assert.False(t, loaded)
}
