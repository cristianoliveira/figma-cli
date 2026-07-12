package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

// hasTextChanges reports whether a text diff contains any added, removed,
// or changed text nodes. Figma renders empty diff slices as JSON null, so a
// length check is the reliable signal.
func hasTextChanges(d extract.TextOutput) bool {
	return len(d.Added) > 0 || len(d.Removed) > 0 || len(d.Changed) > 0
}

var diffTextCmd = &cobra.Command{
	Use:   "text [file-id-or-url] --from version-id --to version-id",
	Short: "Diff text nodes between two Figma file versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromVersion, _ := cmd.Flags().GetString("from")
		toVersion, _ := cmd.Flags().GetString("to")
		explicitNodeID, err := explicitNodeIDFlag(cmd)
		if err != nil {
			return err
		}
		if fromVersion == "" || toVersion == "" {
			return fmt.Errorf("--from and --to are required")
		}

		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		nodeIDs := figma.ResolveNodeIDs(input, explicitNodeID)
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		client = client.WithContext(cmd.Context())

		fromDoc, err := figma.FetchDocument(client, input.FileID, nodeIDs, fromVersion, "")
		if err != nil {
			return err
		}
		toDoc, err := figma.FetchDocument(client, input.FileID, nodeIDs, toVersion, "")
		if err != nil {
			return err
		}

		textDiff := extract.DiffText(
			extract.ExtractTextNodes(fromDoc),
			extract.ExtractTextNodes(toDoc),
		)
		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			if hasTextChanges(textDiff) {
				return nil
			}
			return &cli.ExitCodeError{Code: 1}
		}
		if err := cli.NewPrinter(cmd).JSON(textDiff); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	addNodeIDFlag(diffTextCmd, "node ID to compare; defaults to URL node-id")
	diffTextCmd.Flags().String("from", "", "source Figma version ID")
	diffTextCmd.Flags().String("to", "", "target Figma version ID")
	diffTextCmd.Flags().Bool("quiet", false, "suppress output; exit 0 if changes exist, 1 if none (grep-style)")
	diffCmd.AddCommand(diffTextCmd)
	rootCmd.AddCommand(diffCmd)
}
