package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var textsCmd = &cobra.Command{
	Use:   "texts [file-id-or-url]",
	Short: "Extract text from Figma layers",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		layerName, _ := cmd.Flags().GetString("layer")
		recursive, _ := cmd.Flags().GetBool("recursive")
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}

		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		var doc any
		if len(nodeIDs) > 0 {
			documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
			if err != nil {
				return err
			}
			doc = documents[0]
		} else {
			doc, err = figma.FetchDocument(client, input.FileID, nil, "", "")
			if err != nil {
				return err
			}
		}
		input.NodeIDs = nodeIDs

		result, err := textResult(doc, input, layerName, recursive)
		if err != nil {
			return err
		}
		output := map[string]any{"layer": layerName, "matches": result}
		if layerName == "" {
			output = map[string]any{"nodeId": nodeIDs[0], "texts": result}
		}
		if err := cli.NewPrinter(cmd).JSON(output); err != nil {
			return err
		}
		return nil
	},
}

func textResult(doc any, input *figma.FileInput, layerName string, recursive bool) (any, error) {
	if len(input.NodeIDs) > 0 && layerName == "" {
		return extract.OrderedTextForFrame(doc), nil
	}
	if layerName == "" {
		return nil, fmt.Errorf("--layer or a node ID is required")
	}
	return extract.FindTextByLayerName(doc, layerName, recursive), nil
}

func init() {
	textsCmd.Flags().String("layer", "", "layer name to extract text from")
	textsCmd.Flags().String("id", "", "node ID to extract descendant text from; defaults to URL node-id")
	textsCmd.Flags().Bool("recursive", false, "include text from all descendant nodes")
	rootCmd.AddCommand(textsCmd)
}
