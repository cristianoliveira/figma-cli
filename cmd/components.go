package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeID, _ := cmd.Flags().GetString("id")
		nameFilter, _ := cmd.Flags().GetString("name")
		raw, _ := cmd.Flags().GetBool("raw")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			return fmt.Errorf("components requires --id or a Figma URL with node-id")
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			return err
		}
		var outputValue any = extract.ExtractComponents(doc)
		if raw {
			outputValue = extract.ExtractRawComponents(doc)
		}
		if nameFilter != "" {
			outputValue = extract.FilterByName(outputValue, nameFilter)
		}
		if err := cli.NewPrinter(cmd).JSON(outputValue); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	componentsCmd.Flags().String("id", "", "node ID to inspect; accepts 20089:685897 or 20089-685897")
	componentsCmd.Flags().String("name", "", "filter nodes by name (case-insensitive substring match)")
	componentsCmd.Flags().Bool("raw", false, "output raw Figma node JSON for jq power users")
	rootCmd.AddCommand(componentsCmd)
}
