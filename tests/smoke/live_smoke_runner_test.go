package smoke

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	workingDirectory, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Clean(filepath.Join(workingDirectory, "..", ".."))
}

func TestLiveSmokeRunnerRejectsMissingConfigurationBeforeInvokingBinary(t *testing.T) {
	script := filepath.Join(projectRoot(t), "scripts", "live-smoke.sh")
	command := exec.Command("bash", script)
	command.Env = []string{"PATH=" + t.TempDir(), "FIGMA_ACCESS_TOKEN=secret-that-must-not-appear"}

	output, err := command.CombinedOutput()

	require.Error(t, err)
	assert.Equal(t, 2, command.ProcessState.ExitCode())
	assert.Contains(t, string(output), "set FIGMA_SMOKE_FILE_URL")
	assert.NotContains(t, string(output), "secret-that-must-not-appear")
}

func TestLiveSmokeRunnerRejectsMissingTokenBeforeLookingUpBinary(t *testing.T) {
	script := filepath.Join(projectRoot(t), "scripts", "live-smoke.sh")
	command := exec.Command("bash", script)
	command.Env = []string{
		"PATH=" + t.TempDir(),
		"FIGMA_SMOKE_FILE_URL=https://www.figma.com/design/example",
		"FIGMA_SMOKE_NODE_URL=https://www.figma.com/design/example?node-id=1-1",
	}

	output, err := command.CombinedOutput()

	require.Error(t, err)
	assert.Equal(t, 2, command.ProcessState.ExitCode())
	assert.Contains(t, string(output), "set FIGMA_ACCESS_TOKEN")
	assert.NotContains(t, string(output), "figma binary not found")
}
