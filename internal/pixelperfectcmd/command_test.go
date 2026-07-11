package pixelperfectcmd

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
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

	result := executeCommand(newCommand(diff.CompareImagesWithIgnoredRegions), reference, actual, "--output", mask)

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"width":2,"height":2,"changedPixels":0,"comparedPixels":4,"changedRatio":0,"rmse":0,"rgbRmse":0,"luminanceRmse":0,"alphaRmse":0,"edgeRmse":0,"mask":"`+mask+`"}`, result.Stdout)
}

func TestGroupImageRegionsMergesNearbyClusters(t *testing.T) {
	regions := []diff.Region{
		{Bounds: diff.Bounds{X: 0, Y: 0, Width: 2, Height: 2}, ChangedPixels: 3},
		{Bounds: diff.Bounds{X: 4, Y: 1, Width: 2, Height: 2}, ChangedPixels: 4},
		{Bounds: diff.Bounds{X: 20, Y: 20, Width: 1, Height: 1}, ChangedPixels: 1},
	}

	grouped := groupImageRegions(regions, 2)

	require.Len(t, grouped, 2)
	assert.Equal(t, diff.Region{Bounds: diff.Bounds{X: 0, Y: 0, Width: 6, Height: 3}, ChangedPixels: 7}, grouped[0])
}

func TestFilterImageRegionsRemovesTinyClusters(t *testing.T) {
	regions := []diff.Region{{ChangedPixels: 2}, {ChangedPixels: 20}, {ChangedPixels: 5}}

	filtered := filterImageRegions(regions, 5)

	assert.Equal(t, []diff.Region{{ChangedPixels: 20}, {ChangedPixels: 5}}, filtered)
}

func TestParseImageRegion(t *testing.T) {
	region, err := parseImageRegion("10, 20,300,400")

	require.NoError(t, err)
	assert.Equal(t, &diff.Bounds{X: 10, Y: 20, Width: 300, Height: 400}, region)
}

func TestDiffImageCommandRejectsEmptyIgnoredRegions(t *testing.T) {
	for _, region := range []string{"0,0,0,1", "0,0,1,0", "0,0,-1,1", "0,0,1,-1"} {
		t.Run(region, func(t *testing.T) {
			result := executeCommand(newCommand(diff.CompareImagesWithIgnoredRegions), "reference.png", "actual.png", "--output", "mask.png", "--ignore-region", region)

			assert.EqualError(t, result.Err, "invalid --ignore-region: width and height must be positive")
		})
	}
}

func TestDiffImageCommandRejectsInvalidAnalysisLimits(t *testing.T) {
	tests := []struct {
		name, flag, value, expected string
	}{
		{name: "negative offset radius", flag: "--suggest-offset", value: "-1", expected: "--suggest-offset must be non-negative"},
		{name: "negative region gap", flag: "--region-gap", value: "-1", expected: "--region-gap must be non-negative"},
		{name: "zero minimum region pixels", flag: "--min-region-pixels", value: "0", expected: "--min-region-pixels must be positive"},
		{name: "NaN RMSE limit", flag: "--max-rmse", value: "NaN", expected: "--max-rmse must be -1 or a finite non-negative number"},
		{name: "invalid negative RMSE limit", flag: "--max-rmse", value: "-2", expected: "--max-rmse must be -1 or a finite non-negative number"},
		{name: "changed ratio above one", flag: "--max-changed-ratio", value: "1.1", expected: "--max-changed-ratio must be -1 or between 0 and 1"},
		{name: "invalid negative changed ratio", flag: "--max-changed-ratio", value: "-2", expected: "--max-changed-ratio must be -1 or between 0 and 1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := executeCommand(newCommand(diff.CompareImagesWithIgnoredRegions), "reference.png", "actual.png", "--output", "mask.png", test.flag, test.value)

			assert.EqualError(t, result.Err, test.expected)
		})
	}
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

	result := executeCommand(newCommand(diff.CompareImagesWithIgnoredRegions), reference, actual, "--output", mask, "--max-changed-ratio", "0.1")

	assert.EqualError(t, result.Err, "image diff validation failed: changed ratio 0.250000 exceeds maximum 0.100000")
}

func writeTestPNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, img))
	require.NoError(t, file.Close())
}

func executeCommand(command *cobra.Command, args ...string) commandResult {
	var stdout, stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs(args)
	err := command.Execute()
	return commandResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
}

type commandResult struct {
	Stdout, Stderr string
	Err            error
}
