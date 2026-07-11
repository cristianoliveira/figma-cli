package smoke

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPixelPerfectHandlesThousandsOfDisconnectedChanges(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	mask := filepath.Join(dir, "mask.png")
	base := image.NewRGBA(image.Rect(0, 0, 512, 512))
	changed := image.NewRGBA(base.Bounds())
	changedPixels := 0
	for y := 1; y < 512; y += 4 {
		for x := 1; x < 512; x += 4 {
			changed.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
			changedPixels++
		}
	}
	writeSmokePNG(t, reference, base)
	writeSmokePNG(t, actual, changed)

	output, err := exec.Command(binary, reference, actual, "--output", mask).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 512*512, comparison.ComparedPixels)
	assert.Equal(t, changedPixels, comparison.ChangedPixels)
	require.Len(t, comparison.Regions, 20)
	for _, region := range comparison.Regions {
		assert.Equal(t, 1, region.ChangedPixels)
		assert.Equal(t, 1, region.Bounds.Width)
		assert.Equal(t, 1, region.Bounds.Height)
	}
	assertPNGDimensions(t, mask, 512, 512)
}

func TestPixelPerfectHandlesDenseAlphaGradient(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	mask := filepath.Join(dir, "mask.png")
	overlay := filepath.Join(dir, "overlay.png")
	base := image.NewRGBA(image.Rect(0, 0, 1024, 512))
	changed := image.NewRGBA(base.Bounds())
	for y := 0; y < 512; y++ {
		for x := 0; x < 1024; x++ {
			alpha := uint8(x % 256)
			base.SetRGBA(x, y, color.RGBA{R: 40, G: 80, B: 120, A: alpha})
			changed.SetRGBA(x, y, color.RGBA{R: 42, G: 78, B: 125, A: alpha})
		}
	}
	writeSmokePNG(t, reference, base)
	writeSmokePNG(t, actual, changed)

	output, err := exec.Command(binary, reference, actual, "--output", mask, "--overlay", overlay, "--threshold", "1").CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 1024*512, comparison.ComparedPixels)
	assert.Greater(t, comparison.ChangedPixels, 500_000)
	assert.Greater(t, comparison.RGBRMSE, 0.0)
	assert.Equal(t, 0.0, comparison.AlphaRMSE)
	assertPNGDimensions(t, mask, 1024, 512)
	assertPNGDimensions(t, overlay, 1024, 512)
}

func writeSmokePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, img))
	require.NoError(t, file.Close())
}
