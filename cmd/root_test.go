package cmd

import (
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootNoArgsShowsCompactReadinessAndNextSteps(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "")
	root := newRootCommandWithExecutable(func() (string, error) { return "~/bin/figma", nil }, &cobra.Command{Use: "inspect"})
	root.SilenceErrors = true
	root.SilenceUsage = true
	var stdout strings.Builder
	root.SetOut(&stdout)
	root.SetArgs(nil)

	require.NoError(t, root.Execute())
	result := stdout.String()
	assert.Contains(t, result, "figma")
	assert.Contains(t, result, "Executable: ~/bin/figma")
	assert.Contains(t, result, "Authentication: missing FIGMA_ACCESS_TOKEN")
	assert.Contains(t, result, "figma me")
	assert.Contains(t, result, "figma inspect")
	assert.Contains(t, result, "figma --help")
	assert.NotContains(t, result, "Available Commands:")
	assert.Less(t, len(result), 400)
}

func TestStructuredOutputDefaultsToTOONAndRetainsJSONCompatibility(t *testing.T) {
	newStructuredCommand := func() *cobra.Command {
		return &cobra.Command{
			Use: "show",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return cli.NewPrinter(cmd).Structured(map[string]any{"results": []map[string]any{{"id": 1, "name": "Ada"}}})
			},
		}
	}

	toonRoot := newRootCommand(newStructuredCommand())
	var toonOutput strings.Builder
	toonRoot.SetOut(&toonOutput)
	toonRoot.SetArgs([]string{"show"})
	require.NoError(t, toonRoot.Execute())
	assert.Equal(t, "results[1]{id,name}:\n  1,Ada\n", toonOutput.String())

	jsonResult := executeCommand(newStructuredCommand())
	require.NoError(t, jsonResult.Err)
	assert.JSONEq(t, `{"results":[{"id":1,"name":"Ada"}]}`, jsonResult.Stdout)
}

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
		newMetaCommand(DepsForLoadClient(nil)),
		newInspectCommand(DepsForLoadClient(nil)),
		newFindCommand(DepsForLoadClient(nil)),
		newLayoutCommand(DepsForLoadClient(nil)),
		newExportCommandWithClient(DepsForLoadClient(nil), nil),
		newFramesCommand(DepsForLoadClient(nil)),
		newComponentsCommand(DepsForLoadClient(nil)),
		newTextsCommand(DepsForLoadClient(nil)),
		newColorsCommand(DepsForLoadClient(nil)),
		newAssetsCommandWithClient(DepsForLoadClient(nil), nil),
		newCSSCommand(DepsForLoadClient(nil)),
		newTokensCommand(DepsForLoadClient(nil)),
	}

	for _, command := range commands {
		t.Run(command.Name(), func(t *testing.T) {
			assert.NotEmpty(t, command.Example)
			assert.Contains(t, command.Example, "figma "+command.Name())
		})
	}
}

func TestWorkspaceDiscoveryCommandsProvideLocalExamples(t *testing.T) {
	commands := []*cobra.Command{newMeCommand(DepsForLoadClient(nil)), newProjectsCommand(DepsForLoadClient(nil)), newFilesCommand(DepsForLoadClient(nil)), newVersionsCommand(DepsForLoadClient(nil)), newCommentsCommand(DepsForLoadClient(nil))}

	for _, command := range commands {
		t.Run(command.Name(), func(t *testing.T) {
			assert.NotEmpty(t, command.Example)
			assert.Contains(t, command.Example, "figma "+command.Name())
		})
	}
}

func TestChangeAnalysisCommandsProvideLocalExamples(t *testing.T) {
	deps := DepsForLoadClient(nil)
	commands := map[string]*cobra.Command{
		"figma changes":        newChangesCommand(deps),
		"figma diff":           newDiffCommand(deps),
		"figma diff text":      newDiffTextCommand(deps),
		"figma diff blame":     newDiffBlameCommand(deps),
		"figma layout compare": newLayoutCompareCommand(deps),
	}

	for prefix, command := range commands {
		t.Run(command.CommandPath(), func(t *testing.T) {
			assert.NotEmpty(t, command.Example)
			assert.Contains(t, command.Example, prefix)
		})
	}
}

func TestUnknownCommandUsesUsageExitCode(t *testing.T) {
	root := newRootCommand(&cobra.Command{Use: "inspect"})
	root.SetArgs([]string{"inspec"})

	err := root.Execute()

	require.Error(t, err)
	assert.ErrorContains(t, err, `unknown command "inspec"`)
	assert.Equal(t, 2, cli.ExitCode(err))
}

func TestCommandOwnedInputErrorsUseUsageExitCode(t *testing.T) {
	tests := []struct {
		name    string
		command *cobra.Command
		args    []string
	}{
		{name: "assets kind", command: newAssetsCommandWithClient(DepsForLoadClient(nil), nil), args: []string{"abc", "--id", "1:2", "--kind", "unsupported"}},
		{name: "changes versions", command: newChangesCommand(DepsForLoadClient(nil)), args: []string{"abc"}},
		{name: "comments state", command: newCommentsCommand(DepsForLoadClient(nil)), args: []string{"abc", "--state", "unsupported"}},
		{name: "components kind", command: newComponentsCommand(DepsForLoadClient(nil)), args: []string{"abc", "--kind", "unsupported"}},
		{name: "find filters", command: newFindCommand(DepsForLoadClient(nil)), args: []string{"abc"}},
		{name: "frames scope", command: newFramesCommand(DepsForLoadClient(nil)), args: []string{"abc"}},
		{name: "inspect format", command: newInspectCommand(DepsForLoadClient(nil)), args: []string{"abc", "--id", "1:2", "--format", "unsupported"}},
		{name: "layout comparison", command: newLayoutCompareCommand(DepsForLoadClient(nil)), args: []string{"abc", "--id", "1:2"}},
		{name: "tokens team", command: newTokensCommand(DepsForLoadClient(nil)), args: []string{"abc", "--team", "123"}},
		{name: "tokens format", command: newTokensCommand(DepsForLoadClient(nil)), args: []string{"abc", "--format", "unsupported"}},
		{name: "tokens source", command: newTokensCommand(DepsForLoadClient(nil)), args: []string{"abc", "--source", "unsupported"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := executeCommand(test.command, test.args...)

			require.Error(t, result.Err)
			assert.Equal(t, 2, cli.ExitCode(result.Err))
		})
	}
}

func TestUnknownFlagErrorSuggestsNearestLocalOptionWithoutDumpingHelp(t *testing.T) {
	command := newInspectCommand(DepsForLoadClient(func() (*figma.Client, error) { return nil, nil }))
	result := executeCommand(command, "abc", "--nide", "1:2")

	require.Error(t, result.Err)
	message := result.Err.Error()
	assert.Contains(t, message, "unknown flag: --nide")
	assert.Contains(t, message, "Did you mean `--node`?")
	assert.Contains(t, message, "Run `figma inspect --help` for valid flags.")
	assert.NotContains(t, message, "Available flags")
	assert.Less(t, len(message), 180)
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
