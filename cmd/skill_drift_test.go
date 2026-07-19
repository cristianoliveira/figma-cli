package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFigmaSkillExactAXIFactsMatchCobra(t *testing.T) {
	reference := readSkillFile(t, "..", "skills", "figma-cli", "references", "commands.md")
	skill := readSkillFile(t, "..", "skills", "figma-cli", "SKILL.md")
	layout := findCommand(t, rootCmd, "layout")

	depth := layout.Flags().Lookup("depth")
	require.NotNil(t, depth, "skill drift: figma layout --depth is missing; run: update Cobra or skills/figma-cli/references/commands.md")
	assert.Equal(t, "4", depth.DefValue)
	require.NotNil(t, layout.Flags().Lookup("full"), "skill drift: figma layout --full is missing")

	for _, commandName := range []string{"colors", "comments", "components", "find", "frames", "inspect", "texts"} {
		command := findCommand(t, rootCmd, commandName)
		limit := command.Flags().Lookup("limit")
		require.NotNilf(t, limit, "skill drift: figma %s --limit is missing", commandName)
		assert.Equalf(t, "100", limit.DefValue, "skill drift: figma %s --limit default changed", commandName)
		require.NotNilf(t, command.Flags().Lookup("full"), "skill drift: figma %s --full is missing", commandName)
	}

	assert.Contains(t, skill, "figma layout --depth 4", "skill drift: update skills/figma-cli/SKILL.md")
	assert.Contains(t, skill, "--limit", "skill drift: update skills/figma-cli/SKILL.md")
	assert.Contains(t, reference, "figma layout --depth 4", "skill drift: update skills/figma-cli/references/commands.md")
	assert.Contains(t, reference, "--full", "skill drift: update skills/figma-cli/references/commands.md")
	assert.Contains(t, reference, "--limit", "skill drift: update skills/figma-cli/references/commands.md")
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
