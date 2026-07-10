package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
