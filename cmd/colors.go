package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var colorsCmd = &cobra.Command{
	Use:   "colors [figma-url-or-file-id]",
	Short: "Extract the color palette from a Figma node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			cli.Die(fmt.Errorf("colors requires --id or a Figma URL with node-id"))
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			cli.Die(err)
		}
		palette := extract.CollectColors(doc)
		output, err := json.MarshalIndent(palette, "", "  ")
		if err != nil {
			cli.Die(err)
		}
		fmt.Println(string(output))
	},
}

func init() {
	colorsCmd.Flags().String("id", "", "node ID to extract colors from; accepts 20089:685897 or 20089-685897")
	rootCmd.AddCommand(colorsCmd)
}
