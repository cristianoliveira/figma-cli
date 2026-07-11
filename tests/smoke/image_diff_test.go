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
		name        string
		actual      string
		flags       []string
		changed     int
		regionCount int
	}{
		{name: "identical images", actual: "identical.png"},
		{name: "disconnected changes", actual: "two-regions.png", changed: 2, regionCount: 2},
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
