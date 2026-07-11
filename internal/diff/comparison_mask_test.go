package diff

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIgnoredRegionsFromMaskConvertsExcludedPixelsToRuns(t *testing.T) {
	dir := t.TempDir()
	referencePath := filepath.Join(dir, "reference.png")
	maskPath := filepath.Join(dir, "mask.png")
	writeTestPNG(t, referencePath, image.NewRGBA(image.Rect(0, 0, 4, 2)))
	mask := image.NewRGBA(image.Rect(0, 0, 4, 2))
	mask.Set(1, 0, color.White)
	mask.Set(2, 0, color.White)
	mask.Set(2, 1, color.White)
	writeTestPNG(t, maskPath, mask)

	regions, err := IgnoredRegionsFromMask(maskPath, referencePath)

	require.NoError(t, err)
	assert.Equal(t, []Bounds{
		{X: 0, Y: 0, Width: 1, Height: 1},
		{X: 3, Y: 0, Width: 1, Height: 1},
		{X: 0, Y: 1, Width: 2, Height: 1},
		{X: 3, Y: 1, Width: 1, Height: 1},
	}, regions)
}
