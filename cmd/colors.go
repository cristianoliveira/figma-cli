package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var colorsCmd = &cobra.Command{
	Use:   "colors [figma-url-or-file-id]",
	Short: "Extract the color palette from a Figma node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		nodeIDs, err := figma.ResolveRequiredNodeIDs(input, nodeID, "colors")
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			return err
		}
		palette := extract.CollectColors(doc)
		if err := cli.NewPrinter(cmd).JSON(palette); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	colorsCmd.Flags().String("id", "", "node ID to extract colors from; defaults to URL node-id")
	rootCmd.AddCommand(colorsCmd)
}
