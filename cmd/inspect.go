package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [figma-url-or-file-id]",
	Short: "Show a curated summary of a specific Figma node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		explicitNodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		nodeID, err := inspectNodeID(input, explicitNodeID)
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		documents, err := figma.FetchNodeDocuments(client, input.FileID, []string{nodeID})
		if err != nil {
			return err
		}
		document, ok := documents[0].(map[string]any)
		if !ok {
			return fmt.Errorf("node %s has an invalid document", nodeID)
		}
		node := extract.NodeToInspectOutput(document)
		if err := cli.NewPrinter(cmd).JSON(node); err != nil {
			return err
		}
		return nil
	},
}

func inspectNodeID(input *figma.FileInput, explicitNodeID string) (string, error) {
	nodeIDs := figma.ResolveNodeIDs(input, explicitNodeID)
	if len(nodeIDs) == 0 {
		return "", fmt.Errorf("inspect requires a Figma URL with node-id or --id")
	}
	if len(nodeIDs) != 1 {
		return "", fmt.Errorf("inspect requires exactly one node ID")
	}
	return nodeIDs[0], nil
}

func init() {
	inspectCmd.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	rootCmd.AddCommand(inspectCmd)
}
