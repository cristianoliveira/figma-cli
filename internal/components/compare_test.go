package components

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareMatchesBaseFigmaNamesAgainstCodeComponents(t *testing.T) {
	comparison := Compare(
		[]FigmaComponent{
			{Name: "Button / Primary / Large", NodeID: "1:1"},
			{Name: "Avatar Group", NodeID: "1:2"},
		},
		[]CodeComponent{
			{Name: "Button", Path: "src/components/Button"},
			{Name: "AvatarGroup", Path: "src/components/AvatarGroup"},
			{Name: "LegacyButton", Path: "src/components/LegacyButton"},
		},
	)

	assert.Equal(t, []Match{
		{Figma: FigmaComponent{Name: "Button / Primary / Large", NodeID: "1:1"}, Code: CodeComponent{Name: "Button", Path: "src/components/Button"}},
		{Figma: FigmaComponent{Name: "Avatar Group", NodeID: "1:2"}, Code: CodeComponent{Name: "AvatarGroup", Path: "src/components/AvatarGroup"}},
	}, comparison.Matched)
	assert.Empty(t, comparison.Missing)
	assert.Equal(t, []CodeComponent{{Name: "LegacyButton", Path: "src/components/LegacyButton"}}, comparison.Extra)
}

func TestDiscoverCodeComponentsUsesDirectoriesAndRootSourceFiles(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, writeTestFile(filepath.Join(root, "Button", "Button.tsx")))
	require.NoError(t, writeTestFile(filepath.Join(root, "Card", "index.tsx")))
	require.NoError(t, writeTestFile(filepath.Join(root, "AvatarGroup.tsx")))
	require.NoError(t, writeTestFile(filepath.Join(root, "README.md")))

	got, err := DiscoverCodeComponents(root)

	require.NoError(t, err)
	assert.Equal(t, []CodeComponent{
		{Name: "AvatarGroup", Path: filepath.Join(root, "AvatarGroup.tsx")},
		{Name: "Button", Path: filepath.Join(root, "Button")},
		{Name: "Card", Path: filepath.Join(root, "Card")},
	}, got)
}

func writeTestFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, nil, 0o644)
}

func TestDiscoverCodeComponentsRejectsMissingDirectory(t *testing.T) {
	_, err := DiscoverCodeComponents(filepath.Join(t.TempDir(), "missing"))

	assert.ErrorContains(t, err, "reading codebase")
}
