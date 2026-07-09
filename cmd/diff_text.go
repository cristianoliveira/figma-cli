package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var diffTextCmd = &cobra.Command{
	Use:   "text [file-id-or-url] --from version-id --to version-id",
	Short: "Diff text nodes between two Figma file versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fromVersion, _ := cmd.Flags().GetString("from")
		toVersion, _ := cmd.Flags().GetString("to")
		if fromVersion == "" || toVersion == "" {
			return fmt.Errorf("--from and --to are required")
		}

		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}

		fromDoc, err := figma.FetchDocument(client, input.FileID, input.NodeIDs, fromVersion, "")
		if err != nil {
			return err
		}
		toDoc, err := figma.FetchDocument(client, input.FileID, input.NodeIDs, toVersion, "")
		if err != nil {
			return err
		}

		textDiff := extract.DiffText(
			extract.ExtractTextNodes(fromDoc),
			extract.ExtractTextNodes(toDoc),
		)
		if err := cli.NewPrinter(cmd).JSON(textDiff); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	diffTextCmd.Flags().String("from", "", "source Figma version ID")
	diffTextCmd.Flags().String("to", "", "target Figma version ID")
	diffCmd.AddCommand(diffTextCmd)
	rootCmd.AddCommand(diffCmd)
}
