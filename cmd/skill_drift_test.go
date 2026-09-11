package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFigmaSkillUsesProgressiveCommandDiscovery(t *testing.T) {
	skill := readSkillFile(t, "..", "skills", "figma-cli", "SKILL.md")

	assert.Contains(t, skill, "figma --help")
	assert.Contains(t, skill, "figma <command> --help")
	assert.NotContains(t, skill, "## Workflow")
	assert.NotContains(t, skill, "Verify authentication")

	skillRoot := NewRootCommand(DefaultDeps())
	for _, commandName := range []string{"inspect", "layout", "frames", "texts", "assets", "comments", "versions"} {
		findCommand(t, skillRoot, commandName)
		assert.Containsf(t, skill, "`figma "+commandName+"`", "skill routing omits figma %s", commandName)
	}
}

func TestFigmaSkillExamplesMatchOutputAndNodeIDConventions(t *testing.T) {
	reference := readSkillFile(t, "..", "skills", "figma-cli", "references", "commands.md")
	skill := readSkillFile(t, "..", "skills", "figma-cli", "SKILL.md")
	guidance := skill + "\n" + reference

	assert.Regexp(t, regexp.MustCompile(`figma --json inspect [^\n]+ \| jq`), guidance)
	assert.Regexp(t, regexp.MustCompile(`figma --json find [^\n]+ \| jq`), guidance)
	assert.NotRegexp(t, regexp.MustCompile(`node-id=[0-9]+:[0-9]+`), guidance)
	assert.NotContains(t, reference, "→ JSON")
}

func TestFigmaSkillKeepsAuthenticationCheckOnFailurePath(t *testing.T) {
	skill := readSkillFile(t, "..", "skills", "figma-cli", "SKILL.md")

	assert.Contains(t, skill, "authentication failure")
	assert.Contains(t, skill, "`figma me`")
}

func TestFigmaSkillHomeExamplesRemainDiscoverable(t *testing.T) {
	skill := readSkillFile(t, "..", "skills", "figma-cli", "SKILL.md")
	root := newRootCommandWithExecutable(func() (string, error) { return "~/bin/figma", nil })
	var home strings.Builder
	root.SetOut(&home)
	root.SetArgs(nil)
	require.NoError(t, root.Execute())

	for _, example := range []string{"figma me", "figma inspect"} {
		assert.Containsf(t, home.String(), example, "home view missing %q", example)
		assert.Containsf(t, skill, example, "skill drift: home example %q is missing", example)
	}
}

func findCommand(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	t.Helper()
	for _, command := range root.Commands() {
		if command.Name() == name {
			return command
		}
	}
	t.Fatalf("skill drift: figma %s command is missing", name)
	return nil
}

func readSkillFile(t *testing.T, path ...string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(path...))
	require.NoError(t, err)
	return string(contents)
}
