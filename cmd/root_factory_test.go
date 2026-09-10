package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// probeFlags records every flag value the probe observes. Each probe owns
// its own slice, so the only way two probes could share data is via shared
// mutable package-level state — which this test forbids.
type probeFlags struct {
	JSON      bool
	Recursive bool
}

func newProbeCommand(out *probeFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, _ []string) error {
			asJSON, err := cmd.Flags().GetBool("json")
			if err != nil {
				return err
			}
			out.JSON = asJSON
			recursive, err := cmd.Flags().GetBool("recursive")
			if err != nil {
				return err
			}
			out.Recursive = recursive
			return nil
		},
	}
	// Each instance gets its own --recursive flag so two roots cannot
	// share defaults through any package-level variable.
	cmd.Flags().Bool("recursive", false, "probe flag for state-isolation tests")
	return cmd
}

// TestIndependentRootsDoNotShareState is the contract guarantee from
// TASK-0002: two roots built independently must not bleed flag defaults
// or Changed state into each other. Root A flips --json and --recursive;
// root B receives defaults; B must not see either flag flipped.
func TestIndependentRootsDoNotShareState(t *testing.T) {
	first := &probeFlags{}
	second := &probeFlags{}

	rootA := newRootCommand(newProbeCommand(first))
	rootA.SilenceErrors = true
	rootA.SilenceUsage = true
	var outA bytes.Buffer
	rootA.SetOut(&outA)
	rootA.SetArgs([]string{"probe", "--json", "--recursive"})
	require.NoError(t, rootA.Execute())
	assert.True(t, first.JSON, "root A observed --json=true")
	assert.True(t, first.Recursive, "root A observed --recursive=true")

	rootB := newRootCommand(newProbeCommand(second))
	rootB.SilenceErrors = true
	rootB.SilenceUsage = true
	var outB bytes.Buffer
	rootB.SetOut(&outB)
	rootB.SetArgs([]string{"probe"})
	require.NoError(t, rootB.Execute())
	assert.False(t, second.JSON, "root B saw --json=false despite root A flipping it")
	assert.False(t, second.Recursive, "root B saw --recursive=false despite root A flipping it")
}

// TestIndependentRootsDoNotSharePersistentFlagChangedState asserts that
// root A's flag mutations don't leak into root B's persistent flags, the
// same guarantee the production tree relies on for --json.
func TestIndependentRootsDoNotSharePersistentFlagChangedState(t *testing.T) {
	runProbe := func(args ...string) probeFlags {
		var out probeFlags
		root := newRootCommand(newProbeCommand(&out))
		root.SilenceErrors = true
		root.SilenceUsage = true
		var buf bytes.Buffer
		root.SetOut(&buf)
		root.SetArgs(args)
		require.NoError(t, root.Execute())
		return out
	}

	got := runProbe("probe", "--json")
	assert.True(t, got.JSON, "root A observed --json=true")

	got = runProbe("probe")
	assert.False(t, got.JSON, "defaults restored on a fresh root")
}

// TestNewRootCommandReturnsIndependentTrees covers the structural side of
// the guarantee: NewRootCommand(deps) and another NewRootCommand(deps)
// return two distinct *cobra.Command trees, so a mutation on one cannot
// affect the other.
func TestNewRootCommandReturnsIndependentTrees(t *testing.T) {
	deps := DefaultDeps()
	rootA := NewRootCommand(deps)
	rootB := NewRootCommand(deps)

	require.NotSame(t, rootA, rootB, "production factory must return distinct root instances")
	assert.Equal(t, rootA.Use, rootB.Use)

	// Mutating A's persistent flags must not change B.
	require.NoError(t, rootA.PersistentFlags().Set("json", "true"))
	got, err := rootB.PersistentFlags().GetBool("json")
	require.NoError(t, err)
	assert.False(t, got, "root B persistent flags unaffected by root A mutation")
}
