package cmd

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetFilename(t *testing.T) {
	asset := extract.Asset{ID: "I1:2;3:4", Name: "Iconography / Close"}

	assert.Equal(t, "iconography-close_I1-2-3-4", assetFilename(asset))
}

func TestAssetFilenameFallsBackForUnnamedAsset(t *testing.T) {
	asset := extract.Asset{ID: "1:2", Name: "---"}

	assert.Equal(t, "asset_1-2", assetFilename(asset))
}

func TestAssetNameFilenameSupportsExplicitPrefixTrimming(t *testing.T) {
	asset := extract.Asset{ID: "1:2", Name: "icon/24/arrow-left / dark"}

	assert.Equal(t, "arrow-left-dark", assetNameFilename(asset, "icon/24/"))
}

func TestAssetsCommandExportsFilteredAssets(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := "svg"
		switch request.URL.Path {
		case "/v1/files/FILE/nodes":
			body = `{"nodes":{"1:2":{"document":{"id":"1:2","name":"Screen","type":"FRAME","children":[{"id":"2:3","name":"Close","type":"VECTOR"},{"id":"4:5","name":"Photo","fills":[{"type":"IMAGE"}]}]}}}}`
		case "/v1/images/FILE":
			assert.Equal(t, "2:3", request.URL.Query().Get("ids"))
			body = `{"images":{"2:3":"https://cdn.example/close.svg"}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}
	outputDirectory := t.TempDir()

	result := executeCommand(
		newAssetsCommand(func() (*figma.Client, error) { return client, nil }, nil),
		"https://www.figma.com/design/FILE/Screen?node-id=1-2", "--output", outputDirectory, "--kind", "icon", "--name", "close",
	)

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout, "close_2-3.svg")
	assert.FileExists(t, filepath.Join(outputDirectory, "close_2-3.svg"))
	assert.Contains(t, result.Stderr, "exported 1 asset(s), 0 failed")
}

func TestAssetsCommandFailsOnPartialDownloadWithoutAllowPartial(t *testing.T) {
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		status := http.StatusOK
		body := "svg"
		switch request.URL.Path {
		case "/v1/files/FILE/nodes":
			body = `{"nodes":{"1:2":{"document":{"id":"1:2","name":"Screen","type":"FRAME","children":[{"id":"2:3","name":"Close","type":"VECTOR"},{"id":"4:5","name":"Photo","fills":[{"type":"IMAGE"}]}]}}}}`
		case "/v1/images/FILE":
			body = `{"images":{"2:3":"https://cdn.example/close.svg","4:5":"https://cdn.example/photo.png"}}`
		case "/photo.png":
			status = http.StatusBadGateway
			body = "unavailable"
		}
		return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	client := &figma.Client{HTTP: &http.Client{Transport: transport}}

	result := executeCommand(
		newAssetsCommand(func() (*figma.Client, error) { return client, nil }, nil),
		"https://www.figma.com/design/FILE/Screen?node-id=1-2", "--output", t.TempDir(),
	)

	var exitErr *cli.ExitCodeError
	assert.ErrorAs(t, result.Err, &exitErr)
	assert.Equal(t, 1, exitErr.Code)
	assert.Contains(t, result.Stderr, "warning: node 4:5")
	assert.Contains(t, result.Stderr, "exported 1 asset(s), 1 failed")
}

func TestAssetsCommandRejectsFilenameModeBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newAssetsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}, http.DefaultClient), "abc", "--id", "1:2", "--filename", "random")

	assert.EqualError(t, result.Err, `invalid filename mode "random": expected name or name-id`)
	assert.False(t, loaded)
}

func TestAssetsCommandRejectsFormatBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newAssetsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}, http.DefaultClient), "abc", "--id", "1:2", "--format", "gif")

	assert.EqualError(t, result.Err, `invalid format "gif": expected png, jpg, svg, or pdf`)
	assert.False(t, loaded)
}

func TestAssetExportResultFailsOnPartialExport(t *testing.T) {
	err := assetExportResult(assets.AssetExportManifest{Succeeded: 1, Failed: 1}, false)

	var exitErr *cli.ExitCodeError
	assert.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.Code)
}

func TestAssetExportResultAllowsExplicitPartialExport(t *testing.T) {
	err := assetExportResult(assets.AssetExportManifest{Succeeded: 1, Failed: 1}, true)

	assert.NoError(t, err)
}
