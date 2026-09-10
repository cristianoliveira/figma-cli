package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

type changesOutput struct {
	Scope     output.Scope               `json:"scope"`
	From      string                     `json:"from"`
	To        string                     `json:"to"`
	Total     int                        `json:"total"`
	Truncated int                        `json:"truncated"`
	Changes   []extract.StructuralChange `json:"changes"`
}

// newChangesCommand constructs `figma changes` using the explicit loadClient
// dependency.
func newChangesCommand(deps Deps) *cobra.Command {
	command := &cobra.Command{
		Use:   "changes [figma-url-or-file-id] --from version-id --to version-id",
		Short: "Diff frontend-relevant structure between Figma versions",
		Example: `  figma changes --from <version-id> --to <version-id> <url>
  figma changes --from <version-id> --to <version-id> --id 42:1 <file-key>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromVersion, _ := cmd.Flags().GetString("from")
			toVersion, _ := cmd.Flags().GetString("to")
			explicitNodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			quiet, _ := cmd.Flags().GetBool("quiet")
			terse, _ := cmd.Flags().GetBool("terse")
			limit, _ := cmd.Flags().GetInt("limit")
			if fromVersion == "" || toVersion == "" {
				return cli.NewUsageError(fmt.Errorf("--from and --to are required"))
			}
			if limit <= 0 {
				return cli.NewUsageError(fmt.Errorf("--limit must be greater than zero"))
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			nodeIDs := figma.ResolveNodeIDs(input, explicitNodeID)
			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			fromDocument, err := figma.FetchDocument(client, input.FileID, nodeIDs, fromVersion, "")
			if err != nil {
				return err
			}
			toDocument, err := figma.FetchDocument(client, input.FileID, nodeIDs, toVersion, "")
			if err != nil {
				return err
			}
			changes := extract.DiffDocuments(fromDocument, toDocument)
			if quiet {
				if len(changes) > 0 {
					return nil
				}
				return &cli.ExitCodeError{Code: 1}
			}
			prepared, total, truncated := prepareStructuralChanges(changes, terse, limit)
			result := changesOutput{
				Scope:     output.Scope{FileKey: input.FileID, NodeIDs: append([]string{}, nodeIDs...)},
				From:      fromVersion,
				To:        toVersion,
				Total:     total,
				Truncated: truncated,
				Changes:   prepared,
			}
			return cli.NewPrinter(cmd).Structured(result)
		},
	}
	addNodeIDFlag(command, "node ID to compare; defaults to URL node-id")
	command.Flags().String("from", "", "source Figma version ID")
	command.Flags().String("to", "", "target Figma version ID")
	command.Flags().Bool("quiet", false, "suppress output; exit 0 if changes exist, 1 if none")
	command.Flags().Bool("terse", false, "omit property details and return changed nodes only")
	command.Flags().Int("limit", 1000, "maximum number of changed nodes to emit")
	return command
}

func prepareStructuralChanges(changes []extract.StructuralChange, terse bool, limit int) ([]extract.StructuralChange, int, int) {
	total := len(changes)
	count := min(total, limit)
	prepared := make([]extract.StructuralChange, count)
	copy(prepared, changes[:count])
	if terse {
		for index := range prepared {
			prepared[index].Changes = nil
		}
	}
	return prepared, total, total - count
}
