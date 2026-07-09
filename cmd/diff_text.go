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
	Run: func(cmd *cobra.Command, args []string) {
		fromVersion, _ := cmd.Flags().GetString("from")
		toVersion, _ := cmd.Flags().GetString("to")
		if fromVersion == "" || toVersion == "" {
			cli.Die(fmt.Errorf("--from and --to are required"))
		}

		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}

		fromDoc, err := figma.FetchDocument(client, input.FileID, input.NodeIDs, fromVersion, "")
		if err != nil {
			cli.Die(err)
		}
		toDoc, err := figma.FetchDocument(client, input.FileID, input.NodeIDs, toVersion, "")
		if err != nil {
			cli.Die(err)
		}

		textDiff := extract.DiffText(
			extract.ExtractTextNodes(fromDoc),
			extract.ExtractTextNodes(toDoc),
		)
		if err := cli.NewPrinter(cmd).JSON(textDiff); err != nil {
			cli.Die(err)
		}
	},
}

func init() {
	diffTextCmd.Flags().String("from", "", "source Figma version ID")
	diffTextCmd.Flags().String("to", "", "target Figma version ID")
	diffCmd.AddCommand(diffTextCmd)
	rootCmd.AddCommand(diffCmd)
}
