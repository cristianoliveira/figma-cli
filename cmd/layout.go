package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var layoutCmd = &cobra.Command{
	Use:   "layout [figma-url-or-file-id]",
	Short: "Show an ordered frame tree with layout and copy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		resolvedNodeID, err := figma.ResolveSingleNodeID(input, nodeID, "layout")
		if err != nil {
			return err
		}

		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		documents, err := figma.FetchNodeDocuments(client, input.FileID, []string{resolvedNodeID})
		if err != nil {
			return err
		}
		return cli.NewPrinter(cmd).JSON(extract.ExtractLayout(documents[0]))
	},
}

func init() {
	layoutCmd.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	rootCmd.AddCommand(layoutCmd)
}
