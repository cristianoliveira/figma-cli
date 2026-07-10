package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find [figma-url-or-file-id]",
	Short: "Find Figma layers by name and/or type",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerName, _ := cmd.Flags().GetString("name")
		nodeType, _ := cmd.Flags().GetString("type")
		nodeID, _ := cmd.Flags().GetString("id")
		if layerName == "" && nodeType == "" {
			return fmt.Errorf("at least one of --name or --type is required")
		}
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			return err
		}
		matches := extract.Search(doc, extract.SearchCriteria{Name: layerName, Type: nodeType})
		if err := cli.NewPrinter(cmd).JSON(map[string]any{
			"name":    layerName,
			"type":    nodeType,
			"matches": matches,
		}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	findCmd.Flags().String("name", "", "substring of layer name to find (case-insensitive)")
	findCmd.Flags().String("type", "", "node type to find, e.g. FRAME, COMPONENT, INSTANCE, SECTION (case-insensitive)")
	findCmd.Flags().String("id", "", "node ID to search within; defaults to URL node-id")
	rootCmd.AddCommand(findCmd)
}
