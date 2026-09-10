package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/spf13/cobra"
)

// NewRootCommand is the single composition site for the production command
// tree. Each leaf is constructed by its factory using the explicit Deps, so
// no package-level command instance or init() registration leaks across
// tests. Two independent invocations produce two independent trees with no
// shared flag state.
//
// The factory is the only place that knows the shape of the tree; leaf
// command files expose factories but never bind them to package globals.
// Production callers go through cmd/figma/main.go.
func NewRootCommand(deps Deps) *cobra.Command {
	return buildRoot(deps, nil)
}

// newRootCommand is the test-only helper used by existing tests. It builds
// the persistence flags (--json) of the production root but only attaches
// the caller-supplied test commands. Tests must not inherit the production
// tree or they would collide on command names and the production
// collaborators would run instead of the test stubs.
//
// Production code MUST use NewRootCommand; newRootCommand is unexported so
// only the cmd package can reach it.
func newRootCommand(extraChildren ...*cobra.Command) *cobra.Command {
	deps := DefaultDeps()
	root := &cobra.Command{
		Use:   "figma",
		Short: "Explore and inspect Figma designs from the command line",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			executable, err := deps.ResolveExec()
			if err != nil {
				return err
			}
			authentication := "configured"
			if deps.GetEnv("FIGMA_ACCESS_TOKEN") == "" {
				authentication = "missing FIGMA_ACCESS_TOKEN"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), `figma — agent-facing Figma inspection CLI
Executable: %s
Authentication: %s
Next:
  figma me
  figma inspect "<figma-url>?node-id=<node-id>"
  figma --help
`, executable, authentication)
			return err
		},
		Long: `Query Figma files using a file key or full Figma URL.
Results are structured for scripts and agents. Set FIGMA_ACCESS_TOKEN before
commands that access Figma. Run figma <command> --help for local options.`,
		Example: `  figma meta <url>
  figma inspect "<url>?node-id=42-1"
  figma export --id <node-id> --format png <url>`,
	}
	root.PersistentFlags().Bool("json", false,
		"emit compatibility JSON instead of default TOON (wraps text/file results)")
	root.SetFlagErrorFunc(cli.NewFlagUsageError)
	for _, child := range extraChildren {
		root.AddCommand(child)
	}
	return root
}

// newRootCommandWithExecutable is a test helper that injects a custom
// executable resolver while keeping the rest of the production tree
// intact. It preserves the existing test surface that asserts the readiness
// text uses the configured executable path.
func newRootCommandWithExecutable(resolveExecutable func() (string, error), extraChildren ...*cobra.Command) *cobra.Command {
	deps := DefaultDeps()
	deps.ResolveExec = resolveExecutable
	return buildRoot(deps, extraChildren)
}

// buildRoot is the shared internal assembly used by both the production
// factory and the test helper. extraChildren, when non-empty, is appended to
// the production tree so tests can run a known-good root plus a probe
// command.
func buildRoot(deps Deps, extraChildren []*cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:   "figma",
		Short: "Explore and inspect Figma designs from the command line",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			executable, err := deps.ResolveExec()
			if err != nil {
				return err
			}
			authentication := "configured"
			if deps.GetEnv("FIGMA_ACCESS_TOKEN") == "" {
				authentication = "missing FIGMA_ACCESS_TOKEN"
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), `figma — agent-facing Figma inspection CLI
Executable: %s
Authentication: %s
Next:
  figma me
  figma inspect "<figma-url>?node-id=<node-id>"
  figma --help
`, executable, authentication)
			return err
		},
		Long: `Query Figma files using a file key or full Figma URL.
Results are structured for scripts and agents. Set FIGMA_ACCESS_TOKEN before
commands that access Figma. Run figma <command> --help for local options.`,
		Example: `  figma meta <url>
  figma inspect "<url>?node-id=42-1"
  figma export --id <node-id> --format png <url>`,
	}
	root.PersistentFlags().Bool("json", false,
		"emit compatibility JSON instead of default TOON (wraps text/file results)")
	root.SetFlagErrorFunc(cli.NewFlagUsageError)
	root.AddCommand(
		newMeCommand(),
		newProjectsCommand(deps),
		newFilesCommand(deps),
		newVersionsCommand(deps),
		newMetaCommand(deps),
		newInspectCommand(deps),
		newFindCommand(deps),
		newLayoutCommand(deps),
		newExportCommand(deps),
		newFramesCommand(deps),
		newComponentsCommand(deps),
		newTextsCommand(deps),
		newColorsCommand(deps),
		newAssetsCommand(deps),
		newCSSCommand(deps),
		newTokensCommand(deps),
		newCommentsCommand(deps),
		newChangesCommand(deps),
		newDiffCommand(deps),
	)
	for _, child := range extraChildren {
		root.AddCommand(child)
	}
	return root
}

// DefaultDeps returns a Deps struct wired with production collaborators.
// Tests that need custom dependencies should construct their own Deps value
// directly.
func DefaultDeps() Deps {
	return NewProductionDeps()
}
