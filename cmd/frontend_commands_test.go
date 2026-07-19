package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontendCommandsValidateBeforeLoadingClient(t *testing.T) {
	tests := []struct {
		name    string
		command func(func() (*figma.Client, error)) *cobra.Command
		args    []string
		wantErr string
	}{
		{name: "find criteria", command: newFindCommand, args: []string{"abc"}, wantErr: "at least one of --name or --type is required"},
		{name: "find result limit", command: newFindCommand, args: []string{"abc", "--name", "button", "--limit", "0"}, wantErr: "--limit must be greater than zero"},
		{name: "colors scope", command: newColorsCommand, args: []string{"abc"}, wantErr: "colors requires a Figma URL with node-id or --id"},
		{name: "layout scope", command: newLayoutCommand, args: []string{"abc"}, wantErr: "layout requires a Figma URL with node-id or --id"},
		{name: "css scope", command: newCSSCommand, args: []string{"abc"}, wantErr: "css requires a Figma URL with node-id or --id"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			loaded := false
			command := test.command(func() (*figma.Client, error) {
				loaded = true
				return nil, nil
			})

			result := executeCommand(command, test.args...)

			assert.EqualError(t, result.Err, test.wantErr)
			assert.False(t, loaded)
		})
	}
}

func TestFindCommandEmitsScopedResults(t *testing.T) {
	client := fixtureClient(t, `{"document":{"id":"0:0","name":"Document","type":"DOCUMENT","children":[{"id":"1:1","name":"Button","type":"COMPONENT"}]}}`)

	result := executeCommand(newFindCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--name", "button")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"query":{"name":"button","type":""},"total":1,"results":[{"id":"1:1","name":"Button","type":"COMPONENT"}]}`, result.Stdout)
}

func TestFindCommandReportsTruncationAndSupportsFullOutput(t *testing.T) {
	document := `{"document":{"id":"0:0","name":"Document","type":"DOCUMENT","children":[{"id":"1:1","name":"Button A","type":"COMPONENT"},{"id":"1:2","name":"Button B","type":"COMPONENT"}]}}`
	limitedClient := fixtureClient(t, document)
	fullClient := fixtureClient(t, document)

	limited := executeCommand(newFindCommand(func() (*figma.Client, error) { return limitedClient, nil }), "abc", "--name", "button", "--limit", "1")
	full := executeCommand(newFindCommand(func() (*figma.Client, error) { return fullClient, nil }), "abc", "--name", "button", "--full")

	require.NoError(t, limited.Err)
	var limitedOutput struct {
		Total     int   `json:"total"`
		Truncated bool  `json:"truncated"`
		Results   []any `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(limited.Stdout), &limitedOutput))
	assert.Equal(t, 2, limitedOutput.Total)
	assert.True(t, limitedOutput.Truncated)
	assert.Len(t, limitedOutput.Results, 1)

	require.NoError(t, full.Err)
	var fullOutput struct {
		Total     int   `json:"total"`
		Truncated bool  `json:"truncated"`
		Results   []any `json:"results"`
	}
	require.NoError(t, json.Unmarshal([]byte(full.Stdout), &fullOutput))
	assert.Equal(t, 2, fullOutput.Total)
	assert.False(t, fullOutput.Truncated)
	assert.Len(t, fullOutput.Results, 2)
}

func TestColorsCommandEmitsScopedResults(t *testing.T) {
	client := fixtureClient(t, `{"document":{"id":"0:0","name":"Document","type":"DOCUMENT"}}`)

	result := executeCommand(newColorsCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--id", "1:1")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:1"]},"total":0,"results":[]}`, result.Stdout)
}

func TestCSSCommandEmitsJSONWrappedArtifact(t *testing.T) {
	client := fixtureClient(t, `{"nodes":{"1:1":{"document":{"id":"1:1","name":"Frame","type":"FRAME","layoutMode":"HORIZONTAL"}}}}`)

	result := executeCommand(newCSSCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--id", "1:1", "--json")

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout, `"css"`)
	assert.Contains(t, result.Stdout, `display: flex`)
}

func TestCSSCommandCreatesMissingOutputDirectories(t *testing.T) {
	client := fixtureClient(t, `{"nodes":{"1:1":{"document":{"id":"1:1","name":"Frame","type":"FRAME","layoutMode":"HORIZONTAL"}}}}`)
	path := filepath.Join(t.TempDir(), "missing", "styles", "frame.css")

	result := executeCommand(newCSSCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--id", "1:1", "--output", path)

	require.NoError(t, result.Err)
	assert.FileExists(t, path)
}

func TestLayoutCommandEmitsScopedDetail(t *testing.T) {
	client := fixtureClient(t, `{"nodes":{"1:1":{"document":{"id":"1:1","name":"Frame","type":"FRAME"}}}}`)

	result := executeCommand(newLayoutCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--id", "1:1")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:1"]},"result":{"id":"1:1","name":"Frame","type":"FRAME"}}`, result.Stdout)
}

func fixtureClient(t *testing.T, body string) *figma.Client {
	t.Helper()
	return fixtureClientWithStatus(t, http.StatusOK, body)
}

func fixtureClientWithStatus(t *testing.T, status int, body string) *figma.Client {
	t.Helper()
	return &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
}
