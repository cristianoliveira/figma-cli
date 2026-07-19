package cli

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFlagUsageErrorSuggestsNearestFlagCompactly(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	command := &cobra.Command{Use: "inspect"}
	root.AddCommand(command)
	command.Flags().String("node", "", "node ID")
	command.Flags().String("format", "json", "output format")

	err := NewFlagUsageError(command, errors.New("unknown flag: --nide"))

	require.Error(t, err)
	assert.Equal(t, 2, ExitCode(err))
	assert.ErrorContains(t, err, "Did you mean `--node`?")
	assert.ErrorContains(t, err, "Run `tool inspect --help` for valid flags.")
	assert.NotContains(t, err.Error(), "output format")
}

func TestNewFlagUsageErrorOmitsUnrelatedSuggestion(t *testing.T) {
	command := &cobra.Command{Use: "tool"}
	command.Flags().String("threshold", "", "threshold")

	err := NewFlagUsageError(command, errors.New("unknown flag: --bogus"))

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "Did you mean")
	assert.ErrorContains(t, err, "Run `tool --help` for valid flags.")
}

func TestNewFlagUsageErrorAcceptsNil(t *testing.T) {
	assert.NoError(t, NewFlagUsageError(&cobra.Command{Use: "tool"}, nil))
}

func TestEditDistanceHandlesUnicodeRunes(t *testing.T) {
	assert.Equal(t, 1, editDistance("nóde", "node"))
	assert.Equal(t, 0, editDistance("node", "node"))
}
