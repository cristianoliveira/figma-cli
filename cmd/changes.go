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
	Scope   output.Scope               `json:"scope"`
	From    string                     `json:"from"`
	To      string                     `json:"to"`
	Changes []extract.StructuralChange `json:"changes"`
}

var changesCmd = newChangesCommand(cli.LoadClient)

func newChangesCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "changes [figma-url-or-file-id] --from version-id --to version-id",
		Short: "Diff frontend-relevant structure between Figma versions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fromVersion, _ := cmd.Flags().GetString("from")
			toVersion, _ := cmd.Flags().GetString("to")
			explicitNodeID, _ := cmd.Flags().GetString("id")
			if fromVersion == "" || toVersion == "" {
				return fmt.Errorf("--from and --to are required")
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			nodeIDs := figma.ResolveNodeIDs(input, explicitNodeID)
			client, err := loadClient()
			if err != nil {
				return err
			}
			fromDocument, err := figma.FetchDocument(client, input.FileID, nodeIDs, fromVersion, "")
			if err != nil {
				return err
			}
			toDocument, err := figma.FetchDocument(client, input.FileID, nodeIDs, toVersion, "")
			if err != nil {
				return err
			}
			result := changesOutput{
				Scope:   output.Scope{FileKey: input.FileID, NodeIDs: append([]string{}, nodeIDs...)},
				From:    fromVersion,
				To:      toVersion,
				Changes: extract.DiffDocuments(fromDocument, toDocument),
			}
			return cli.NewPrinter(cmd).JSON(result)
		},
	}
	command.Flags().String("id", "", "node ID to compare; defaults to URL node-id")
	command.Flags().String("from", "", "source Figma version ID")
	command.Flags().String("to", "", "target Figma version ID")
	return command
}

func init() {
	rootCmd.AddCommand(changesCmd)
}
