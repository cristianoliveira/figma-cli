package cmd

import (
	"encoding/json"
	"image"
	"image/png"
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
		newExportCommand(func() (*figma.Client, error) { return client, nil }, nil),
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
		newExportCommand(func() (*figma.Client, error) { return client, nil }, nil),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "png", "--scale", "2", "--output", outputPath, "--json",
	)

	require.NoError(t, result.Err)
	assert.Contains(t, exportQuery, "scale=2")
	assert.JSONEq(t, `{"path":"`+outputPath+`","format":"png","node":"42:1","scale":2}`, result.Stdout)
}

func TestExportCommandRejectsScaleForVectorFormats(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}, nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--scale", "2")

	assert.EqualError(t, result.Err, "--scale is only supported for png and jpg exports")
	assert.False(t, loaded)
}

func TestExportCommandRejectsInvalidScale(t *testing.T) {
	loaded := false
	result := executeCommand(newExportCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}, nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "png", "--scale", "0")

	assert.EqualError(t, result.Err, "--scale must be a finite number between 0.01 and 4")
	assert.False(t, loaded)
}

func TestExportCommandWritesMetadataSidecar(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := `<svg width="336" height="182" viewBox="0 0 336 182"><path d="M8 6H320C320 6 320 29.8986 320 51C320 60 328 63.935 328 75C328 86.065 320 88.9604 320 99C320 113.466 320 172 320 172H8L8 6Z" /></svg>`
		switch {
		case strings.Contains(request.URL.Path, "/v1/images/"):
			body = `{"images":{"42:1":"https://cdn.example/image"}}`
		case strings.Contains(request.URL.Path, "/v1/files/abc/nodes"):
			body = `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Rectangle Copy 13","type":"VECTOR","absoluteBoundingBox":{"x":136,"y":562,"width":320,"height":166},"effects":[{"type":"DROP_SHADOW","visible":true,"radius":8,"offset":{"x":0,"y":2}}]}}}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	httpClient := &http.Client{Transport: transport}
	client := &figma.Client{HTTP: httpClient}
	dir := t.TempDir()
	outputPath := dir + "/rectangle.svg"
	metadataPath := dir + "/rectangle.export.json"

	result := executeCommand(
		newExportCommand(func() (*figma.Client, error) { return client, nil }, nil),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--output", outputPath, "--metadata", metadataPath, "--json",
	)

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"path":"`+outputPath+`","format":"svg","node":"42:1","scale":1,"metadata":"`+metadataPath+`"}`, result.Stdout)
	metadata, err := os.ReadFile(metadataPath)
	require.NoError(t, err)
	assert.JSONEq(t, `{"version":1,"nodeId":"42:1","format":"svg","scale":1,"nodeBounds":{"x":136,"y":562,"width":320,"height":166},"exportBounds":{"width":336,"height":182},"dimensionDelta":{"width":16,"height":16},"logicalCrop":{"x":8,"y":6,"width":320,"height":166},"exportPadding":{"left":8,"top":6,"right":8,"bottom":10},"paddingEvidence":["DROP_SHADOW radius=8 offsetX=0 offsetY=2"],"output":"`+outputPath+`"}`, string(metadata))
}

func TestMeasureExportBoundsReadsPNGDimensions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "export.png")
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 6, 4))))
	require.NoError(t, file.Close())

	bounds, err := measureExportBounds(path, "png")

	require.NoError(t, err)
	assert.Equal(t, exportSize{Width: 6, Height: 4}, bounds)
}

func TestExportPaddingEvidenceOmitsSpeculationWhenEffectsAreAbsent(t *testing.T) {
	assert.Empty(t, exportPaddingEvidence(nil))
}

func TestExportCommandReportsPartialArtifactWhenMetadataWriteFails(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "image"
		if strings.Contains(request.URL.Path, "/v1/images/") {
			body = `{"images":{"42:1":"https://cdn.example/image"}}`
		}
		if strings.Contains(request.URL.Path, "/v1/files/abc/nodes") {
			body = `{"nodes":{"42:1":{"document":{"id":"42:1","absoluteBoundingBox":{"width":1,"height":1}}}}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}
	outputPath := t.TempDir() + "/button.png"

	result := executeCommand(
		newExportCommand(func() (*figma.Client, error) { return client, nil }, nil),
		"https://www.figma.com/design/abc/Name?node-id=42-1", "--output", outputPath, "--metadata", t.TempDir(),
	)

	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "exported "+outputPath+" but failed to write metadata")
	_, err := os.Stat(outputPath)
	require.NoError(t, err)
}

func TestExportCommandRejectsMetadataOutputCollision(t *testing.T) {
	loaded := false
	path := t.TempDir() + "/artifact.svg"
	result := executeCommand(newExportCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}, nil), "https://www.figma.com/design/abc/Name?node-id=42-1", "--format", "svg", "--output", path, "--metadata", path)

	assert.EqualError(t, result.Err, "--metadata must differ from --output")
	assert.False(t, loaded)
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
