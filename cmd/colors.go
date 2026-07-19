package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var colorsCmd = newColorsCommand(cli.LoadClient)

func newColorsCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "colors [figma-url-or-file-id]",
		Short: "Extract the color palette from a Figma node",
		Example: `  figma colors "<url>?node-id=42-1"
  figma colors --id 42:1 <file-key>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			resultLimit, err := readResultLimit(cmd)
			if err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			nodeIDs, err := figma.ResolveRequiredNodeIDs(input, nodeID, "colors")
			if err != nil {
				return cli.NewUsageError(err)
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
			if err != nil {
				return err
			}
			palette := extract.CollectColors(doc)
			palette, total := limitResults(resultLimit, palette)
			result := newLimitedQuery(cmd, output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}, nil, total, palette)
			if err := cli.NewPrinter(cmd).Structured(result); err != nil {
				return err
			}
			return nil
		},
	}
	addNodeIDFlag(command, "node ID to extract colors from; defaults to URL node-id")
	addResultLimitFlags(command)
	return command
}

func init() {
	rootCmd.AddCommand(colorsCmd)
}
