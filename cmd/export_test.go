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
		newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) { return client, nil }), nil),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--output", outputPath, "--json",
	)

	require.NoError(t, result.Err)
	expectedJSON, err := json.Marshal(map[string]any{
		"path":   outputPath,
		"format": "svg",
		"node":   "42:1",
		"scale":  1.0,
	})
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedJSON), result.Stdout)
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "image", string(content))
}

func TestExportCommandRequestsRasterScale(t *testing.T) {
	var exportQuery string
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "image"
		if strings.Contains(request.URL.Path, "/v1/images/") {
			exportQuery = request.URL.RawQuery
			body = `{"images":{"42:1":"https://cdn.example/image"}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}
	outputPath := t.TempDir() + "/button.png"

	result := executeCommand(
		newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) { return client, nil }), nil),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "png", "--scale", "2", "--output", outputPath, "--json",
	)

	require.NoError(t, result.Err)
	assert.Contains(t, exportQuery, "scale=2")
	assert.JSONEq(t, `{"path":"`+outputPath+`","format":"png","node":"42:1","scale":2}`, result.Stdout)
}

func TestExportCommandDerivesRasterScaleFromWidth(t *testing.T) {
	var exportQuery string
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "image"
		switch {
		case strings.Contains(request.URL.Path, "/v1/files/"):
			body = `{"nodes":{"42:1":{"document":{"id":"42:1","absoluteBoundingBox":{"width":390,"height":200}}}}}`
		case strings.Contains(request.URL.Path, "/v1/images/"):
			exportQuery = request.URL.RawQuery
			body = `{"images":{"42:1":"https://cdn.example/image"}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}
	outputPath := t.TempDir() + "/button.png"

	result := executeCommand(
		newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) { return client, nil }), nil),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "png", "--width", "780", "--output", outputPath, "--json",
	)

	require.NoError(t, result.Err)
	assert.Contains(t, exportQuery, "scale=2")
	assert.JSONEq(t, `{"path":"`+outputPath+`","format":"png","node":"42:1","scale":2,"requestedWidth":780}`, result.Stdout)
}

func TestExportCommandRejectsWidthWithScale(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "png", "--width", "780", "--scale", "2")

	assert.EqualError(t, result.Err, "--width cannot be combined with --scale")
	assert.False(t, loaded)
}

func TestExportCommandRejectsWidthForVectorFormats(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--width", "780")

	assert.EqualError(t, result.Err, "--width is only supported for png and jpg exports")
	assert.False(t, loaded)
}

func TestExportCommandRejectsScaleForVectorFormats(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--scale", "2")

	assert.EqualError(t, result.Err, "--scale is only supported for png and jpg exports")
	assert.False(t, loaded)
}

func TestExportCommandRejectsInvalidScale(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "png", "--scale", "0")

	assert.EqualError(t, result.Err, "--scale must be a finite number between 0.01 and 4")
	assert.False(t, loaded)
}

func TestExportCommandRejectsFormatBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommandWithClient(DepsForLoadClient(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), http.DefaultClient), "abc", "--id", "1:2", "--format", "gif")

	assert.EqualError(t, result.Err, `invalid format "gif": expected png, jpg, svg, or pdf`)
	assert.False(t, loaded)
}
