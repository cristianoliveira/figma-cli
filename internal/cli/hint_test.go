package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullHintPreservesScopeAndQuotesShellValues(t *testing.T) {
	root := &cobra.Command{Use: "tool"}
	command := &cobra.Command{Use: "find <file>"}
	command.Flags().StringArray("id", nil, "")
	command.Flags().String("name", "", "")
	command.Flags().Int("limit", 100, "")
	command.Flags().Bool("full", false, "")
	root.AddCommand(command)
	root.SetArgs([]string{"find", "design's file", "--id", "1:2", "--id", "3:4", "--name", "Primary button", "--limit", "1"})
	command.RunE = func(cmd *cobra.Command, args []string) error {
		assert.Equal(t, `tool find 'design'"'"'s file' --id 1:2 --id 3:4 --name 'Primary button' --full`, FullHint(cmd, args))
		return nil
	}

	require.NoError(t, root.Execute())
}
