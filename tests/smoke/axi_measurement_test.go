package smoke

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAXIMeasurementHarnessCoversRepresentativeDecisions(t *testing.T) {
	command := exec.Command("bash", "scripts/measure-axi.sh")
	command.Dir = filepath.Join("..", "..")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	result := string(output)
	for _, scenario := range []string{
		"figma-discovery",
		"figma-invalid-toon",
		"figma-invalid-json",
		"pixel-identical-toon",
		"pixel-identical-json",
		"pixel-failed-gate-toon",
		"pixel-probe-truncated-csv",
		"pixel-probe-truncated-json",
		"pixel-scan-csv",
		"pixel-scan-json",
	} {
		assert.Contains(t, result, scenario)
	}
}
