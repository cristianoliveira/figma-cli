package cmd

import "github.com/spf13/cobra"

// newDiffCommand constructs the parent `figma diff` command and attaches
// its text and blame subcommands via the loadClient-aware factories. The
// nested assembly happens here (not in init) so two independently built
// roots cannot drift their nested tree.
func newDiffCommand(deps Deps) *cobra.Command {
	command := &cobra.Command{
		Use:   "diff",
		Short: "Diff Figma file versions",
		Example: `  figma diff text --from <version-id> --to <version-id> <url>
  figma diff blame --to <version-id> --id 42:1 <url>`,
	}
	command.AddCommand(newDiffTextCommand(deps))
	command.AddCommand(newDiffBlameCommand(deps))
	return command
}
