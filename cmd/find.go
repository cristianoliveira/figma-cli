package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var findCmd = newFindCommand(cli.LoadClient)

func newFindCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "find [figma-url-or-file-id]",
		Short: "Find Figma layers by name and/or type",
		Example: `  figma find --name "Button" <url>
  figma find --type COMPONENT --id 42:1 <file-key>
  figma find --name "Card" --type INSTANCE <url>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layerName, _ := cmd.Flags().GetString("name")
			nodeType, _ := cmd.Flags().GetString("type")
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			if layerName == "" && nodeType == "" {
				return cli.NewUsageError(fmt.Errorf("at least one of --name or --type is required"))
			}
			resultLimit, err := readResultLimit(cmd)
			if err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			nodeIDs := figma.ResolveNodeIDs(input, nodeID)
			doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
			if err != nil {
				return err
			}
			matches := extract.Search(doc, extract.SearchCriteria{Name: layerName, Type: nodeType})
			matches, total := limitResults(resultLimit, matches)
			query := map[string]any{"name": layerName, "type": nodeType}
			result := newLimitedQuery(cmd, output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}, query, total, matches)
			if err := cli.NewPrinter(cmd).Structured(result); err != nil {
				return err
			}
			return nil
		},
	}
	command.Flags().String("name", "", "substring of layer name to find (case-insensitive)")
	command.Flags().String("type", "", "node type to find, e.g. FRAME, COMPONENT, INSTANCE, SECTION (case-insensitive)")
	addNodeIDFlag(command, "node ID to search within; defaults to URL node-id")
	addResultLimitFlags(command)
	return command
}

func init() {
	rootCmd.AddCommand(findCmd)
}
