package cmd

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffImageCommandProducesMaskAndJSONMetrics(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	mask := filepath.Join(dir, "mask.png")
	writeTestPNG(t, reference, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	writeTestPNG(t, actual, image.NewRGBA(image.Rect(0, 0, 2, 2)))

	result := executeCommand(newDiffImageCommand(diff.CompareImagesInRegion), reference, actual, "--output", mask)

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"width":2,"height":2,"changedPixels":0,"changedRatio":0,"rmse":0,"mask":"`+mask+`"}`, result.Stdout)
}

func TestParseImageRegion(t *testing.T) {
	region, err := parseImageRegion("10, 20,300,400")

	require.NoError(t, err)
	assert.Equal(t, &diff.Bounds{X: 10, Y: 20, Width: 300, Height: 400}, region)
}

func TestDiffImageCommandFailsValidationThreshold(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	mask := filepath.Join(dir, "mask.png")
	writeTestPNG(t, reference, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	changed := image.NewRGBA(image.Rect(0, 0, 2, 2))
	changed.Set(0, 0, image.White)
	writeTestPNG(t, actual, changed)

	result := executeCommand(newDiffImageCommand(diff.CompareImagesInRegion), reference, actual, "--output", mask, "--max-changed-ratio", "0.1")

	assert.EqualError(t, result.Err, "image diff validation failed: changed ratio 0.250000 exceeds maximum 0.100000")
}

func writeTestPNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, img))
	require.NoError(t, file.Close())
}
