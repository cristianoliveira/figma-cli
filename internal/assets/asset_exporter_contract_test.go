package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetExporterEmptyRequestReturnsEmptyManifest(t *testing.T) {
	manifest := AssetExporter{}.Export(t.TempDir(), nil)

	assert.Empty(t, manifest.Items)
	assert.Zero(t, manifest.Succeeded)
	assert.Zero(t, manifest.Failed)
}
