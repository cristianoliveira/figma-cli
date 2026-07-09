package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
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
