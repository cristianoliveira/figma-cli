package smoke

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFigmaCLINoArgsShowsReadinessWithoutToken(t *testing.T) {
	binary := buildCommand(t, "figma")
	command := exec.Command(binary)
	command.Env = withoutEnvironmentVariable(os.Environ(), "FIGMA_ACCESS_TOKEN")

	output, err := command.CombinedOutput()

	require.NoError(t, err, string(output))
	resolvedBinary, resolveErr := filepath.EvalSymlinks(binary)
	require.NoError(t, resolveErr)
	assert.Contains(t, string(output), "Executable: "+resolvedBinary)
	assert.Contains(t, string(output), "Authentication: missing FIGMA_ACCESS_TOKEN")
	assert.Contains(t, string(output), "figma inspect")
	assert.NotContains(t, string(output), "Available Commands:")
	assert.Less(t, len(output), 400)
}

func TestFigmaCLIExitCodeContracts(t *testing.T) {
	binary := buildCommand(t, "figma")
	tests := []struct {
		name, expected string
		args           []string
		exitCode       int
	}{
		{name: "unknown command", args: []string{"inspec"}, exitCode: 2, expected: "unknown command"},
		{name: "invalid option", args: []string{"assets", "abc", "--id", "1:2", "--kind", "unsupported"}, exitCode: 2, expected: "invalid kind"},
		{name: "missing dependency", args: []string{"meta", "abc"}, exitCode: 1, expected: "Figma authentication is not configured."},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := exec.Command(binary, test.args...)
			command.Env = withoutEnvironmentVariable(os.Environ(), "FIGMA_ACCESS_TOKEN")
			var stdout, stderr bytes.Buffer
			command.Stdout = &stdout
			command.Stderr = &stderr

			err := command.Run()

			var exitError *exec.ExitError
			require.ErrorAs(t, err, &exitError)
			assert.Equal(t, test.exitCode, exitError.ExitCode())
			assert.Empty(t, stderr.String())
			assert.Contains(t, stdout.String(), "error:")
			assert.Contains(t, stdout.String(), "exitCode: "+fmt.Sprint(test.exitCode))
			assert.Contains(t, stdout.String(), test.expected)
		})
	}
}

func TestFigmaCLIErrorSupportsJSONCompatibility(t *testing.T) {
	binary := buildCommand(t, "figma")
	command := exec.Command(binary, "--json", "inspec")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	assert.Equal(t, 2, exitError.ExitCode())
	assert.Empty(t, stderr.String())
	assert.JSONEq(t, `{"error":{"category":"usage","message":"unknown command \"inspec\" for \"figma\"","exitCode":2,"recovery":"Run the command with --help to list valid commands."}}`, stdout.String())
}

func withoutEnvironmentVariable(environment []string, name string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(environment))
	for _, value := range environment {
		if strings.HasPrefix(value, prefix) {
			continue
		}
		filtered = append(filtered, value)
	}
	return filtered
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
