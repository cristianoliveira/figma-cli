package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadResultLimitRejectsInvalidAndAmbiguousOptions(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{name: "non-positive", args: []string{"--limit", "0"}, expected: "--limit must be greater than zero"},
		{name: "full and limit", args: []string{"--full", "--limit", "5"}, expected: "--full cannot be combined with --limit"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := &cobra.Command{Use: "query", RunE: func(cmd *cobra.Command, _ []string) error {
				_, err := readResultLimit(cmd)
				return err
			}}
			addResultLimitFlags(command)
			command.SetArgs(test.args)

			err := command.Execute()

			require.Error(t, err)
			assert.ErrorContains(t, err, test.expected)
			assert.Equal(t, 2, cli.ExitCode(err))
		})
	}
}

func TestLimitResultsPreservesTotalAndFullEscapeHatch(t *testing.T) {
	values := []int{1, 2, 3}

	limited, total := limitResults(resultLimit{maximum: 2}, values)
	full, fullTotal := limitResults(resultLimit{maximum: 2, full: true}, values)

	assert.Equal(t, []int{1, 2}, limited)
	assert.Equal(t, 3, total)
	assert.Equal(t, values, full)
	assert.Equal(t, 3, fullTotal)
}
