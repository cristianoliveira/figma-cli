package cmd

import (
	"net/http"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
)

func TestAssetFilename(t *testing.T) {
	asset := extract.Asset{ID: "I1:2;3:4", Name: "Iconography / Close"}

	assert.Equal(t, "iconography-close_I1-2-3-4", assetFilename(asset))
}

func TestAssetFilenameFallsBackForUnnamedAsset(t *testing.T) {
	asset := extract.Asset{ID: "1:2", Name: "---"}

	assert.Equal(t, "asset_1-2", assetFilename(asset))
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
	err := assetExportResult(cli.AssetExportManifest{Succeeded: 1, Failed: 1}, false)

	var exitErr *cli.ExitCodeError
	assert.ErrorAs(t, err, &exitErr)
	assert.Equal(t, 1, exitErr.Code)
}

func TestAssetExportResultAllowsExplicitPartialExport(t *testing.T) {
	err := assetExportResult(cli.AssetExportManifest{Succeeded: 1, Failed: 1}, true)

	assert.NoError(t, err)
}
