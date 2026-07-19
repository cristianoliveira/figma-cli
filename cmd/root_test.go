package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootHelpIsCompactAndPointsToDecisionRelevantCommands(t *testing.T) {
	result := executeCommand(newRootCommand(
		&cobra.Command{Use: "inspect"},
		&cobra.Command{Use: "export"},
	), "--help")

	require.NoError(t, result.Err)
	assert.Less(t, len(result.Stdout), 1800)
	assert.Contains(t, result.Stdout, "figma inspect")
	assert.Contains(t, result.Stdout, "figma export")
	assert.Contains(t, result.Stdout, "figma <command> --help")
	assert.NotContains(t, result.Stdout, "What is this file about?")
}

func TestPrimaryExplorationCommandsProvideLocalExamples(t *testing.T) {
	commands := []*cobra.Command{
		newMetaCommand(nil),
		newInspectCommand(nil),
		newFindCommand(nil),
		newLayoutCommand(nil),
		newExportCommand(nil, nil),
		newFramesCommand(nil),
		newComponentsCommand(nil),
		newTextsCommand(nil),
		newColorsCommand(nil),
		newAssetsCommand(nil, nil),
		newCSSCommand(nil),
		newTokensCommand(nil),
	}

	for _, command := range commands {
		t.Run(command.Name(), func(t *testing.T) {
			assert.NotEmpty(t, command.Example)
			assert.Contains(t, command.Example, "figma "+command.Name())
		})
	}
}

func TestWorkspaceDiscoveryCommandsProvideLocalExamples(t *testing.T) {
	commands := []*cobra.Command{meCmd, projectsCmd, filesCmd, versionsCmd, newCommentsCommand(nil)}

	for _, command := range commands {
		t.Run(command.Name(), func(t *testing.T) {
			assert.NotEmpty(t, command.Example)
			assert.Contains(t, command.Example, "figma "+command.Name())
		})
	}
}

func TestUnknownFlagErrorIncludesAvailableOptions(t *testing.T) {
	command := newInspectCommand(func() (*figma.Client, error) { return nil, nil })
	result := executeCommand(command, "abc", "--bogus")

	require.Error(t, result.Err)
	message := result.Err.Error()
	assert.Contains(t, message, "unknown flag: --bogus")
	assert.Contains(t, message, "Available flags for \"figma inspect\"")
	assert.Contains(t, message, "--id string")
	assert.Contains(t, message, "--node string")
	assert.Contains(t, message, "Global flags:")
	assert.Contains(t, message, "--json")
	assert.Contains(t, message, "Run `figma inspect --help` for details.")
}

func TestNewRootCommandDoesNotLeakPersistentFlags(t *testing.T) {
	newProbe := func(values *[]bool) *cobra.Command {
		return &cobra.Command{
			Use:  "probe",
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				asJSON, err := cmd.Flags().GetBool("json")
				*values = append(*values, asJSON)
				return err
			},
		}
	}

	firstValues := []bool{}
	first := newRootCommand(newProbe(&firstValues))
	first.SetArgs([]string{"probe", "--json"})
	require.NoError(t, first.Execute())

	secondValues := []bool{}
	second := newRootCommand(newProbe(&secondValues))
	second.SetArgs([]string{"probe"})
	require.NoError(t, second.Execute())

	assert.Equal(t, []bool{true}, firstValues)
	assert.Equal(t, []bool{false}, secondValues)
}
