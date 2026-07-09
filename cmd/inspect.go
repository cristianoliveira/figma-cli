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
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		if nodeID == "" {
			cli.Die(fmt.Errorf("--id is required"))
		}
		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			cli.Die(fmt.Errorf("could not resolve node ID"))
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			cli.Die(err)
		}
		found := extract.FindNodeByID(doc, nodeIDs[0])
		if found == nil {
			cli.Die(fmt.Errorf("node %s not found", nodeIDs[0]))
		}
		node := extract.NodeToInspectOutput(found)
		if err := cli.NewPrinter(cmd).JSON(node); err != nil {
			cli.Die(err)
		}
	},
}

func init() {
	inspectCmd.Flags().String("id", "", "node ID to inspect; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(inspectCmd)
}
