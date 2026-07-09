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
	Short: "Find Figma layers by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerName, _ := cmd.Flags().GetString("name")
		nodeID, _ := cmd.Flags().GetString("id")
		if layerName == "" {
			return fmt.Errorf("--name is required")
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
		matches := extract.FindLayersByName(doc, layerName)
		if err := cli.NewPrinter(cmd).JSON(map[string]any{"name": layerName, "matches": matches}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	findCmd.Flags().String("name", "", "exact layer name to find")
	findCmd.Flags().String("id", "", "node ID to search within; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(findCmd)
}
