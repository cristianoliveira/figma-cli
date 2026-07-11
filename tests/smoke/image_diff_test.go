package smoke

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageDiffScenarios(t *testing.T) {
	binary := buildCLI(t)
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
			args := []string{"diff", "image", filepath.Join(fixtures, "reference.png"), filepath.Join(fixtures, test.actual), "--output", mask, "--overlay", overlay}
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

func buildCLI(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "figma")
	command := exec.Command("go", "build", "-o", binary, "./cmd/figma")
	command.Dir = filepath.Join("..", "..")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	return binary
}
