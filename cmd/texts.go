package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

var textsCmd = &cobra.Command{
	Use:   "texts [file-id-or-url]",
	Short: "Extract text from Figma layers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerName, _ := cmd.Flags().GetString("layer")
		recursive, _ := cmd.Flags().GetBool("recursive")
		if layerName == "" {
			return fmt.Errorf("--layer is required")
		}

		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}

		var doc any
		if len(input.NodeIDs) > 0 {
			// Specific nodes use the /nodes endpoint, which returns one document per node.
			nodesURL, err := figma.BuildNodesURL(input.FileID, input.NodeIDs)
			if err != nil {
				return err
			}
			var resp api.GetFileNodesResponse
			if err := client.Fetch(nodesURL, &resp); err != nil {
				return err
			}
			for _, node := range resp.Nodes {
				d, err := figma.UnmarshalDocument(node.Document)
				if err != nil {
					return err
				}
				doc = d
				break
			}
		} else {
			var err error
			doc, err = figma.FetchDocument(client, input.FileID, nil, "", "")
			if err != nil {
				return err
			}
		}

		matches := extract.FindTextByLayerName(doc, layerName, recursive)
		if err := cli.NewPrinter(cmd).JSON(map[string]any{"layer": layerName, "matches": matches}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	textsCmd.Flags().String("layer", "", "layer name to extract text from")
	textsCmd.Flags().Bool("recursive", false, "include text from all descendant nodes")
	rootCmd.AddCommand(textsCmd)
}
