package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplicitNodeIDFlagAcceptsNodeAlias(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	addNodeIDFlag(command, "node ID")
	require.NoError(t, command.Flags().Set("node", "0:147"))

	got, err := explicitNodeIDFlag(command)

	require.NoError(t, err)
	assert.Equal(t, "0:147", got)
}

func TestExplicitNodeIDFlagRejectsConflictingAliases(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	addNodeIDFlag(command, "node ID")
	require.NoError(t, command.Flags().Set("id", "0:147"))
	require.NoError(t, command.Flags().Set("node", "0:148"))

	_, err := explicitNodeIDFlag(command)

	assert.EqualError(t, err, "--id and --node must match when both are provided")
}

func TestExplicitNodeIDsFlagAcceptsRepeatedNodeAlias(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	addNodeIDsFlag(command, "node IDs")
	require.NoError(t, command.Flags().Set("node", "0:147"))
	require.NoError(t, command.Flags().Set("node", "0:148"))

	got, err := explicitNodeIDsFlag(command)

	require.NoError(t, err)
	assert.Equal(t, []string{"0:147", "0:148"}, got)
}

func TestExplicitNodeIDsFlagRejectsConflictingRepeatedAliases(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	addNodeIDsFlag(command, "node IDs")
	require.NoError(t, command.Flags().Set("id", "0:147"))
	require.NoError(t, command.Flags().Set("node", "0:148"))

	_, err := explicitNodeIDsFlag(command)

	assert.EqualError(t, err, "--id and --node must match when both are provided")
}
