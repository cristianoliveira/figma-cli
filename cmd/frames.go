package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var framesCmd = newFramesCommand(cli.LoadClient)

func newFramesCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "frames [figma-url-or-file-id]",
		Short: "List screen-level frames in a page or section",
		Example: `  figma frames "<url>?node-id=1-2"
  figma frames --id 1:2 <file-key>`,
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
			nodeIDs := figma.ResolveNodeIDs(input, nodeID)
			if len(nodeIDs) == 0 {
				return cli.NewUsageError(fmt.Errorf("frames requires a page or section node ID from the URL or --id"))
			}

			client, err := loadClient()
			if err != nil {
				return err
			}
			document, err := figma.FetchDocument(client.WithContext(cmd.Context()), input.FileID, nodeIDs, "", "")
			if err != nil {
				return err
			}
			frames := extract.DiscoverFrames(document)
			frames, total := limitResults(resultLimit, frames)
			result := output.NewLimitedQuery(output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}, nil, total, frames)
			return cli.NewPrinter(cmd).Structured(result)
		},
	}
	addNodeIDFlag(command, "page or section node ID; defaults to URL node-id")
	addResultLimitFlags(command)
	return command
}

func init() {
	rootCmd.AddCommand(framesCmd)
}
