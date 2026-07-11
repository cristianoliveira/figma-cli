package smoke

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"

	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPixelPerfectAgainstUpstreamComparisonCorpus(t *testing.T) {
	binary := buildCommand(t, "pixel-perfect")
	fixtures := filepath.Join("fixtures", "upstream")
	tests := []struct {
		name, project, reference, actual string
		width, height                    int
		changed, regions                 int
		assertMetrics                    func(*testing.T, diff.ImageComparison)
	}{
		{
			name: "pixelmatch realistic antialias and transparency", project: "pixelmatch", reference: "1a.png", actual: "1b.png",
			width: 512, height: 256, changed: 12_933, regions: 14,
			assertMetrics: func(t *testing.T, result diff.ImageComparison) {
				assert.Greater(t, result.RGBRMSE, 0.0)
				assert.Greater(t, result.EdgeRMSE, 0.0)
			},
		},
		{
			name: "looks-same antialias-only raster variation", project: "looks-same", reference: "antialiasing-ref.png", actual: "antialiasing-actual.png",
			width: 12, height: 15, changed: 5, regions: 5,
			assertMetrics: func(t *testing.T, result diff.ImageComparison) {
				assert.Greater(t, result.EdgeRMSE, result.RGBRMSE)
			},
		},
		{
			name: "odiff equivalent extreme alpha encodings", project: "odiff", reference: "extreme-alpha.png", actual: "extreme-alpha-1.png",
			width: 450, height: 450,
			assertMetrics: func(t *testing.T, result diff.ImageComparison) {
				assert.Zero(t, result.RMSE)
				assert.Zero(t, result.AlphaRMSE)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mask := filepath.Join(t.TempDir(), "mask.png")
			output, err := exec.Command(binary,
				filepath.Join(fixtures, test.project, test.reference),
				filepath.Join(fixtures, test.project, test.actual),
				"--output", mask,
			).CombinedOutput()

			require.NoError(t, err, string(output))
			var comparison diff.ImageComparison
			require.NoError(t, json.Unmarshal(output, &comparison))
			assert.Equal(t, test.width, comparison.Width)
			assert.Equal(t, test.height, comparison.Height)
			assert.Equal(t, test.width*test.height, comparison.ComparedPixels)
			assert.Equal(t, test.changed, comparison.ChangedPixels)
			assert.Len(t, comparison.Regions, test.regions)
			assertPNGDimensions(t, mask, test.width, test.height)
			test.assertMetrics(t, comparison)
		})
	}
}
