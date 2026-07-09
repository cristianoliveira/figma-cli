package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var componentsCmd = &cobra.Command{
	Use:   "components [figma-url-or-file-id]",
	Short: "List descendant nodes within a Figma element as JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		nameFilter, _ := cmd.Flags().GetString("name")
		raw, _ := cmd.Flags().GetBool("raw")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			cli.Die(fmt.Errorf("components requires --id or a Figma URL with node-id"))
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			cli.Die(err)
		}
		var outputValue any = extract.ExtractComponents(doc)
		if raw {
			outputValue = extract.ExtractRawComponents(doc)
		}
		if nameFilter != "" {
			outputValue = extract.FilterByName(outputValue, nameFilter)
		}
		output, err := json.MarshalIndent(outputValue, "", "  ")
		if err != nil {
			cli.Die(err)
		}
		fmt.Println(string(output))
	},
}

func init() {
	componentsCmd.Flags().String("id", "", "node ID to inspect; accepts 20089:685897 or 20089-685897")
	componentsCmd.Flags().String("name", "", "filter nodes by name (case-insensitive substring match)")
	componentsCmd.Flags().Bool("raw", false, "output raw Figma node JSON for jq power users")
	rootCmd.AddCommand(componentsCmd)
}
