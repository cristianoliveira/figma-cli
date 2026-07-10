package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var layoutCmd = newLayoutCommand(cli.LoadClient)

func newLayoutCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "layout [figma-url-or-file-id]",
		Short: "Show an ordered frame tree with layout and copy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, _ := cmd.Flags().GetString("id")
			measureSpacing, _ := cmd.Flags().GetBool("measure-spacing")
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			resolvedNodeID, err := figma.ResolveSingleNodeID(input, nodeID, "layout")
			if err != nil {
				return err
			}

			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			documents, err := figma.FetchNodeDocuments(client, input.FileID, []string{resolvedNodeID})
			if err != nil {
				return err
			}
			layout := extract.ExtractLayout(documents[0], extract.LayoutOptions{MeasureSpacing: measureSpacing})
			result := output.Detail[extract.LayoutNode]{
				Scope:  output.Scope{FileKey: input.FileID, NodeIDs: []string{resolvedNodeID}},
				Result: layout,
			}
			return cli.NewPrinter(cmd).JSON(result)
		},
	}
	command.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	command.Flags().Bool("measure-spacing", false, "measure geometric gaps between adjacent layout children")
	command.AddCommand(newLayoutCompareCommand(loadClient))
	return command
}

func init() {
	rootCmd.AddCommand(layoutCmd)
}
