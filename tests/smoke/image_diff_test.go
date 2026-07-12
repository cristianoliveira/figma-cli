package smoke

import (
	"encoding/json"
	"fmt"
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
	assert.Equal(t, &diff.Bounds{X: 0, Y: 0, Width: 4, Height: 3}, comparison.Bounds)
	assert.Equal(t, []int{0, 2}, comparison.ChangedRows)
	require.Len(t, comparison.Regions, 2)
	assert.NotEmpty(t, comparison.Regions[0].Classification)
}

func TestPixelPerfectProbeCLI(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	output, err := exec.Command(binary, "probe", filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"), "--at", "0,0").CombinedOutput()

	require.NoError(t, err, string(output))
	assert.Equal(t, "x,y,ref,act,delta,input_ref,input_act\n0,0,#FFFFFF,#000000,255,,\n", string(output))
}

func TestPixelPerfectScanCLI(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	output, err := exec.Command(binary, "scan", filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"), "--y", "0").CombinedOutput()

	require.NoError(t, err, string(output))
	assert.Contains(t, string(output), "image,axis,index,start,end,length,hex,input_axis,input_index")
	assert.Contains(t, string(output), "ref,x,0,0,3,4,#FFFFFF,,")
	assert.Contains(t, string(output), "act,x,0,0,0,1,#000000,,")
}

func TestPixelPerfectScanCLIErrorContracts(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	output, err := exec.Command(binary, "scan", filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"), "--x", "0", "--y", "0").CombinedOutput()

	require.Error(t, err)
	assert.Contains(t, string(output), "provide exactly one of --x or --y")
}

func TestPixelPerfectProbeCLIErrorContracts(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	output, err := exec.Command(binary, "probe", filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "unequal-dimensions.png"), "--at", "0,0").CombinedOutput()

	require.Error(t, err)
	assert.Contains(t, string(output), "image dimensions differ")
}

func TestPixelPerfectStandaloneCLIDefaultMask(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	workDir := t.TempDir()
	reference := filepath.Join(workDir, "reference.png")
	actual := filepath.Join(workDir, "actual.png")
	require.NoError(t, copyFile(filepath.Join(fixtures, "reference.png"), reference))
	require.NoError(t, copyFile(filepath.Join(fixtures, "two-regions.png"), actual))
	output, err := exec.Command(binary, reference, actual, "--threshold", "8").CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 2, comparison.ChangedPixels)
	assert.Equal(t, filepath.Join(workDir, "actual.diff.png"), comparison.Mask)
	_, statErr := os.Stat(comparison.Mask)
	require.NoError(t, statErr)
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
				assert.Greater(t, comparison.Regions[0].PerceptualRMSE, 0.0)
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

func TestPixelPerfectAppliesExplicitCropFixtures(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "crop-reference.png"), filepath.Join(fixtures, "crop-actual.png"),
		"--reference-crop", "1,0,2,2",
		"--actual-crop", "0,0,2,2",
		"--output", mask,
	).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 2, comparison.Width)
	assert.Equal(t, 2, comparison.Height)
	assert.Equal(t, 0, comparison.ChangedPixels)
	assert.Empty(t, comparison.Regions)
	require.NotNil(t, comparison.Inputs)
	assert.Equal(t, diff.ImageInput{Width: 4, Height: 2, Crop: &diff.Bounds{X: 1, Y: 0, Width: 2, Height: 2}}, comparison.Inputs.Reference)
	assert.Equal(t, diff.ImageInput{Width: 2, Height: 2, Crop: &diff.Bounds{X: 0, Y: 0, Width: 2, Height: 2}}, comparison.Inputs.Actual)
	assertPNGDimensions(t, mask, 2, 2)
}

func TestPixelPerfectExplicitCropReportsOriginalInputBounds(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "crop-reference.png"), filepath.Join(fixtures, "crop-actual-changed.png"),
		"--reference-crop", "1,0,2,2",
		"--actual-crop", "0,0,2,2",
		"--output", mask,
	).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	require.Len(t, comparison.Regions, 1)
	assert.Equal(t, diff.Bounds{X: 1, Y: 1, Width: 1, Height: 1}, comparison.Regions[0].Bounds)
	assert.Equal(t, &diff.InputBounds{
		Reference: diff.Bounds{X: 2, Y: 1, Width: 1, Height: 1},
		Actual:    diff.Bounds{X: 1, Y: 1, Width: 1, Height: 1},
	}, comparison.Regions[0].InputBounds)
}

func TestPixelPerfectExplicitCropFixtureErrors(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	reference := filepath.Join(fixtures, "crop-reference.png")
	actual := filepath.Join(fixtures, "crop-actual.png")
	tests := []struct {
		name, expected string
		flags          []string
	}{
		{name: "out of bounds", flags: []string{"--actual-crop", "1,0,2,2"}, expected: "invalid --actual-crop: crop 1,0,2,2 is outside image bounds 2x2"},
		{name: "dimension mismatch", flags: []string{"--reference-crop", "1,0,2,2", "--actual-crop", "0,0,1,2"}, expected: "cropped image dimensions differ: reference is 2x2, actual is 1x2"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := []string{reference, actual, "--output", filepath.Join(t.TempDir(), "mask.png")}
			args = append(args, test.flags...)
			output, err := exec.Command(binary, args...).CombinedOutput()

			require.Error(t, err)
			assert.Contains(t, string(output), test.expected)
		})
	}
}

func TestPixelPerfectOmitsOffsetWhenMaskExcludesAllPixels(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"),
		"--output", mask,
		"--mask", filepath.Join(fixtures, "all-excluded-mask.png"),
		"--suggest-offset", "2",
	).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Zero(t, comparison.ComparedPixels)
	assert.Nil(t, comparison.SuggestedOffset)
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

func TestPixelPerfectHelpDocumentsStandaloneContract(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	output, err := exec.Command(binary, "--help").CombinedOutput()

	require.NoError(t, err, string(output))
	help := string(output)
	assert.Contains(t, help, "pixel-perfect <reference.png> <actual.png>")
	for _, flag := range []string{"--output", "--overlay", "--region", "--reference-crop", "--actual-crop", "--ignore-region", "--mask", "--threshold", "--perceptual-threshold", "--suggest-offset", "--max-rmse", "--max-changed-ratio", "--max-perceptual-changed-ratio"} {
		assert.Contains(t, help, flag)
	}
}

func TestPixelPerfectRejectsRegionsOutsideImage(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	reference := filepath.Join(fixtures, "reference.png")
	actual := filepath.Join(fixtures, "two-regions.png")
	for _, region := range []string{"0,0,0,0", "-1,0,2,2", "100,100,2,2", "3,2,5,5"} {
		t.Run(region, func(t *testing.T) {
			mask := filepath.Join(t.TempDir(), "mask.png")
			output, err := exec.Command(binary, reference, actual, "--output", mask, "--region", region).CombinedOutput()

			require.Error(t, err)
			assert.Contains(t, string(output), "is outside image bounds 4x3")
			_, statErr := os.Stat(mask)
			assert.ErrorIs(t, statErr, os.ErrNotExist)
		})
	}
}

func TestPixelPerfectRejectsArtifactPathCollisions(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	reference := filepath.Join(fixtures, "reference.png")
	actual := filepath.Join(fixtures, "two-regions.png")
	tests := []struct {
		name, expected string
		args           func(*testing.T) []string
	}{
		{name: "mask overwrites reference", expected: "--output must not overwrite an input image", args: func(*testing.T) []string { return []string{reference, actual, "--output", reference} }},
		{name: "mask overwrites actual", expected: "--output must not overwrite an input image", args: func(*testing.T) []string { return []string{reference, actual, "--output", actual} }},
		{name: "overlay overwrites mask", expected: "--overlay must differ from --output", args: func(t *testing.T) []string {
			artifact := filepath.Join(t.TempDir(), "artifact.png")
			return []string{reference, actual, "--output", artifact, "--overlay", artifact}
		}},
		{name: "overlay overwrites reference", expected: "--overlay must not overwrite an input image", args: func(t *testing.T) []string {
			return []string{reference, actual, "--output", filepath.Join(t.TempDir(), "mask.png"), "--overlay", reference}
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

func TestPixelPerfectRejectsInvalidFlagValues(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	reference := filepath.Join(fixtures, "reference.png")
	actual := filepath.Join(fixtures, "two-regions.png")
	tests := []struct {
		name, expected string
		flags          []string
	}{
		{name: "threshold overflow", flags: []string{"--threshold", "256"}, expected: "value out of range"},
		{name: "malformed region", flags: []string{"--region", "0,0,4"}, expected: "--region must be x,y,width,height"},
		{name: "non numeric region", flags: []string{"--region", "0,zero,4,3"}, expected: "--region must contain integers"},
		{name: "malformed ignored region", flags: []string{"--ignore-region", "0,0,4"}, expected: "invalid --ignore-region"},
		{name: "empty ignored region", flags: []string{"--ignore-region", "0,0,0,1"}, expected: "invalid --ignore-region: width and height must be positive"},
		{name: "negative ignored region size", flags: []string{"--ignore-region", "0,0,1,-1"}, expected: "invalid --ignore-region: width and height must be positive"},
		{name: "negative offset radius", flags: []string{"--suggest-offset", "-1"}, expected: "--suggest-offset must be non-negative"},
		{name: "negative region gap", flags: []string{"--region-gap", "-1"}, expected: "--region-gap must be non-negative"},
		{name: "zero minimum region pixels", flags: []string{"--min-region-pixels", "0"}, expected: "--min-region-pixels must be positive"},
		{name: "negative perceptual threshold", flags: []string{"--perceptual-threshold", "-0.1"}, expected: "--perceptual-threshold must be a finite non-negative number"},
		{name: "NaN perceptual threshold", flags: []string{"--perceptual-threshold", "NaN"}, expected: "--perceptual-threshold must be a finite non-negative number"},
		{name: "NaN RMSE limit", flags: []string{"--max-rmse", "NaN"}, expected: "--max-rmse must be -1 or a finite non-negative number"},
		{name: "invalid negative RMSE limit", flags: []string{"--max-rmse", "-2"}, expected: "--max-rmse must be -1 or a finite non-negative number"},
		{name: "changed ratio above one", flags: []string{"--max-changed-ratio", "1.1"}, expected: "--max-changed-ratio must be -1 or between 0 and 1"},
		{name: "invalid negative changed ratio", flags: []string{"--max-changed-ratio", "-2"}, expected: "--max-changed-ratio must be -1 or between 0 and 1"},
		{name: "perceptual changed ratio above one", flags: []string{"--max-perceptual-changed-ratio", "1.1"}, expected: "--max-perceptual-changed-ratio must be -1 or between 0 and 1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := []string{reference, actual, "--output", filepath.Join(t.TempDir(), "mask.png")}
			args = append(args, test.flags...)
			output, err := exec.Command(binary, args...).CombinedOutput()
			require.Error(t, err)
			assert.Contains(t, string(output), test.expected)
		})
	}
}

func TestPixelPerfectCombinesRegionAndRepeatedIgnores(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"),
		"--output", mask,
		"--region", "0,0,4,2",
		"--ignore-region", "0,0,1,1",
		"--ignore-region", "1,0,1,1",
	).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 0, comparison.ChangedPixels)
	assert.Equal(t, 6, comparison.ComparedPixels)
	assert.Equal(t, &diff.Bounds{X: 0, Y: 0, Width: 4, Height: 2}, comparison.ComparedRegion)
	assertPNGDimensions(t, mask, 4, 2)
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

func TestPixelPerfectAcceptsPerceptualThresholdAboveBlackWhiteDistance(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	mask := filepath.Join(t.TempDir(), "mask.png")
	output, err := exec.Command(binary,
		filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, "two-regions.png"),
		"--output", mask, "--perceptual-threshold", "1.1",
	).CombinedOutput()

	require.NoError(t, err, string(output))
	var comparison diff.ImageComparison
	require.NoError(t, json.Unmarshal(output, &comparison))
	assert.Equal(t, 1.1, comparison.PerceptualThreshold)
	assert.Zero(t, comparison.PerceptualChangedPixels)
	assert.Equal(t, 2, comparison.ChangedPixels)
}

func TestPixelPerfectValidationGateBoundaries(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "image-diff")
	reference := filepath.Join(fixtures, "reference.png")
	actual := filepath.Join(fixtures, "two-regions.png")
	baselineMask := filepath.Join(t.TempDir(), "baseline.png")
	baselineOutput, err := exec.Command(binary, reference, actual, "--output", baselineMask).CombinedOutput()
	require.NoError(t, err, string(baselineOutput))
	var baseline diff.ImageComparison
	require.NoError(t, json.Unmarshal(baselineOutput, &baseline))

	tests := []struct {
		name, flag string
		limit      float64
		passes     bool
		expected   string
	}{
		{name: "RMSE accepts exact boundary", flag: "--max-rmse", limit: baseline.RMSE, passes: true},
		{name: "RMSE rejects below boundary", flag: "--max-rmse", limit: baseline.RMSE / 2, expected: "RMSE"},
		{name: "changed ratio accepts exact boundary", flag: "--max-changed-ratio", limit: baseline.ChangedRatio, passes: true},
		{name: "changed ratio rejects below boundary", flag: "--max-changed-ratio", limit: baseline.ChangedRatio / 2, expected: "changed ratio"},
		{name: "perceptual ratio accepts exact boundary", flag: "--max-perceptual-changed-ratio", limit: baseline.PerceptualChangedRatio, passes: true},
		{name: "perceptual ratio rejects below boundary", flag: "--max-perceptual-changed-ratio", limit: baseline.PerceptualChangedRatio / 2, expected: "perceptual changed ratio"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mask := filepath.Join(t.TempDir(), "mask.png")
			output, err := exec.Command(binary, reference, actual, "--output", mask, test.flag, fmt.Sprintf("%.17g", test.limit)).CombinedOutput()
			if test.passes {
				require.NoError(t, err, string(output))
				var comparison diff.ImageComparison
				require.NoError(t, json.Unmarshal(output, &comparison))
				return
			}
			require.Error(t, err)
			assert.Contains(t, string(output), "image diff validation failed")
			assert.Contains(t, string(output), test.expected)
			assertPNGDimensions(t, mask, 4, 3)
		})
	}
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

func copyFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, data, 0o600)
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
