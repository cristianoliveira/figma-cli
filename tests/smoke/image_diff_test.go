package smoke

import (
	"encoding/json"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPixelPerfectStandaloneCLI(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary, filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"), "--output", mask).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 2, comparison.ChangedPixels)
	require.Len(t, comparison.Regions, 2)
	assert.NotEmpty(t, comparison.Regions[0].Classification)
}

func TestImageDiffScenarios(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	tests := []struct {
		name                  string
		actual                string
		flags                 []string
		changed               int
		regionCount           int
		expectedCompared      int
		expectsOffset         bool
		expectsColorPair      bool
		expectsClassification bool
	}{
		{name: "identical images", actual: "identical.png"},
		{name: "disconnected changes", actual: "two-regions.png", changed: 2, regionCount: 2},
		{name: "region reports dominant color pair", actual: "two-regions.png", changed: 2, regionCount: 2, expectsColorPair: true},
		{name: "region reports likely mismatch class", actual: "two-regions.png", changed: 2, regionCount: 2, expectsClassification: true},
		{name: "tiny regions can be omitted", actual: "two-regions.png", flags: []string{"--min-region-pixels", "2"}, changed: 2},
		{name: "nearby regions can be grouped", actual: "two-regions.png", flags: []string{"--region-gap", "4"}, changed: 2, regionCount: 1},
		{name: "known dynamic area can be ignored", actual: "two-regions.png", flags: []string{"--ignore-region", "0,0,1,1"}, changed: 1, regionCount: 1},
		{name: "comparison mask selects pixels", actual: "two-regions.png", flags: []string{"--mask", filepath.Join(fixtures, "comparison-mask.png")}, changed: 1, regionCount: 1, expectedCompared: 11},
		{name: "translation can be reported", actual: "two-regions.png", flags: []string{"--suggest-offset", "1"}, changed: 2, regionCount: 2, expectsOffset: true},
		{name: "threshold ignores subtle rendering noise", actual: "subtle-change.png", flags: []string{"--threshold", "5"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			outputDir := t.TempDir()
			mask := filepath.Join(outputDir, "mask.png")
			overlay := filepath.Join(outputDir, "overlay.png")
			args := []string{filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, test.actual), "--output", mask, "--overlay", overlay}
			args = append(args, test.flags...)
			output, err := exec.Command(binary, args...).CombinedOutput()

			require.NoError(t, err, string(output))
			var comparison diff.ImageComparison
			require.NoError(t, json.Unmarshal(output, &comparison))
			assert.Equal(t, test.changed, comparison.ChangedPixels)
			assert.Len(t, comparison.Regions, test.regionCount)
			if test.regionCount > 0 {
				assert.Greater(t, comparison.Regions[0].ChangedRatio, 0.0)
				assert.Greater(t, comparison.Regions[0].RMSE, 0.0)
			}
			if test.expectedCompared > 0 {
				assert.Equal(t, test.expectedCompared, comparison.ComparedPixels)
			}
			if test.expectsOffset {
				assert.NotNil(t, comparison.SuggestedOffset)
			}
			if test.expectsColorPair {
				require.NotEmpty(t, comparison.Regions[0].DominantColorPairs)
				assert.Greater(t, comparison.Regions[0].DominantColorPairs[0].Pixels, 0)
			}
			if test.expectsClassification {
				assert.NotEmpty(t, comparison.Regions[0].Classification)
			}
			_, err = os.Stat(mask)
			require.NoError(t, err)
			_, err = os.Stat(overlay)
			require.NoError(t, err)
		})
	}
}

func TestPixelPerfectBoundaryAndCompositingScenarios(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	tests := []struct {
		name, reference, actual string
		flags                   []string
		changed, compared       int
		assertResult            func(*testing.T, diff.ImageComparison)
	}{
		{name: "difference equal to threshold is ignored", reference: "reference.png", actual: "threshold-equal.png", flags: []string{"--threshold", "5"}, compared: 12},
		{name: "difference above threshold is detected", reference: "reference.png", actual: "threshold-exceeded.png", flags: []string{"--threshold", "5"}, changed: 12, compared: 12},
		{name: "alpha-only changes affect alpha metric", reference: "reference.png", actual: "alpha-only-change.png", changed: 12, compared: 12, assertResult: func(t *testing.T, result diff.ImageComparison) { assert.Greater(t, result.AlphaRMSE, 0.0) }},
		{name: "hidden RGB is ignored for transparent pixels", reference: "all-excluded-mask.png", actual: "transparent-hidden-rgb.png", compared: 12},
		{name: "fully excluded comparison is valid", reference: "reference.png", actual: "two-regions.png", flags: []string{"--mask", filepath.Join(fixtures, "all-excluded-mask.png")}},
		{name: "translation direction is exact", reference: "offset-reference.png", actual: "offset-right-one.png", flags: []string{"--suggest-offset", "2"}, changed: 2, compared: 12, assertResult: func(t *testing.T, result diff.ImageComparison) {
			require.NotNil(t, result.SuggestedOffset)
			assert.Equal(t, -1, result.SuggestedOffset.X)
			assert.Equal(t, 0, result.SuggestedOffset.Y)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mask := filepath.Join(t.TempDir(), "mask.png")
			args := []string{filepath.Join(fixtures, test.reference), filepath.Join(fixtures, test.actual), "--output", mask}
			args = append(args, test.flags...)
			output, err := exec.Command(binary, args...).CombinedOutput()
			require.NoError(t, err, string(output))
			var comparison diff.ImageComparison
			require.NoError(t, json.Unmarshal(output, &comparison))
			assert.Equal(t, test.changed, comparison.ChangedPixels)
			assert.Equal(t, test.compared, comparison.ComparedPixels)
			assertPNGDimensions(t, mask, 4, 3)
			if test.assertResult != nil {
				test.assertResult(t, comparison)
			}
		})
	}
}

func TestPixelPerfectRejectsWrongSizeComparisonMask(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"),
		"--output", mask, "--mask", filepath.Join(fixtures, "wrong-size-mask.png"),
	).CombinedOutput()

	require.Error(t, err)
	assert.Contains(t, string(output), "comparison mask dimensions differ")
	_, statErr := os.Stat(mask)
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestPixelPerfectCLIErrorContracts(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	reference := filepath.Join(fixtures, "reference.png")
	actual := filepath.Join(fixtures, "two-regions.png")
	tests := []struct {
		name, expected string
		args           func(*testing.T) []string
	}{
		{name: "missing arguments", expected: "error: accepts 2 arg(s), received 0", args: func(*testing.T) []string { return nil }},
		{name: "missing required output", expected: "error: --output is required", args: func(*testing.T) []string { return []string{reference, actual} }},
		{name: "unknown flag", expected: "error: unknown flag: --unknown", args: func(*testing.T) []string { return []string{reference, actual, "--unknown"} }},
		{name: "malformed PNG", expected: "error:", args: func(t *testing.T) []string {
			invalid := filepath.Join(t.TempDir(), "invalid.png")
			require.NoError(t, os.WriteFile(invalid, []byte("not a png"), 0o600))
			return []string{reference, invalid, "--output", filepath.Join(t.TempDir(), "mask.png")}
		}},
		{name: "output is directory", expected: "error:", args: func(t *testing.T) []string {
			return []string{reference, actual, "--output", t.TempDir()}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := exec.Command(binary, test.args(t)...).CombinedOutput()
			require.Error(t, err)
			assert.Contains(t, string(output), test.expected)
		})
	}
}

func TestPixelPerfectOutputIsDeterministic(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	args := []string{
		filepath.Join(fixtures, "real-ui-reference.png"),
		filepath.Join(fixtures, "real-ui-implementation.png"),
		"--output", mask, "--region-gap", "8", "--min-region-pixels", "12",
	}

	first, err := exec.Command(binary, args...).CombinedOutput()
	require.NoError(t, err, string(first))
	firstMask, err := os.ReadFile(mask)
	require.NoError(t, err)
	second, err := exec.Command(binary, args...).CombinedOutput()
	require.NoError(t, err, string(second))
	secondMask, err := os.ReadFile(mask)
	require.NoError(t, err)

	assert.Equal(t, string(first), string(second))
	assert.Equal(t, firstMask, secondMask)
}

func TestPixelPerfectRealUIScreenshot(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	overlay := filepath.Join(t.TempDir(), "overlay.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "real-ui-reference.png"),
		filepath.Join(fixtures, "real-ui-implementation.png"),
		"--output", mask,
		"--overlay", overlay,
		"--region-gap", "8",
		"--min-region-pixels", "12",
	).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 575, comparison.Width)
	assert.Equal(t, 477, comparison.Height)
	assert.Equal(t, 575*477, comparison.ComparedPixels)
	assert.Greater(t, comparison.ChangedPixels, 1_000)
	assert.Greater(t, comparison.RMSE, 0.0)
	assert.Greater(t, comparison.EdgeRMSE, 0.0)
	assert.NotEmpty(t, comparison.Regions)
	assert.LessOrEqual(t, len(comparison.Regions), 20)
	assertPNGDimensions(t, mask, 575, 477)
	assertPNGDimensions(t, overlay, 575, 477)
}

func TestPixelPerfectRealUIValidationGate(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "real-ui-reference.png"),
		filepath.Join(fixtures, "real-ui-implementation.png"),
		"--output", mask,
		"--max-changed-ratio", "0.001",
	).CombinedOutput()

	require.Error(t, err)
	assert.Contains(t, string(output), "error: image diff validation failed: changed ratio")
	assertPNGDimensions(t, mask, 575, 477)
}

func TestPixelPerfectRejectsRealUIWithUnequalDimensions(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "real-ui-reference.png"),
		filepath.Join(fixtures, "unequal-dimensions.png"),
		"--output", mask,
	).CombinedOutput()

	require.Error(t, err)
	assert.Contains(t, string(output), "dimensions")
	_, statErr := os.Stat(mask)
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func assertPNGDimensions(t *testing.T, path string, width, height int) {
	t.Helper()
	file, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, file.Close()) })
	config, err := png.DecodeConfig(file)
	require.NoError(t, err)
	assert.Equal(t, width, config.Width)
	assert.Equal(t, height, config.Height)
}

func buildCommand(t *testing.T, name string) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), name)
	command := exec.Command("go", "build", "-o", binary, "./cmd/"+name)
	command.Dir = filepath.Join("..", "..")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	return binary
}
