package cmd

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var cssCmd = &cobra.Command{
	Use:   "css [figma-url-or-file-id]",
	Short: "Generate CSS rules from a Figma element's layout and styles",
	Long: `Generate CSS from a Figma node tree.

Walks the node subtree and emits one CSS rule per node that contributes a
meaningful property (autolayout, fills, borders, radius, text). Layout maps to
flexbox: layoutMode -> display:flex, itemSpacing -> gap, paddings -> shorthand,
sizing modes -> width/height. Output is deterministic (sorted properties).

This is the public-API equivalent of Figma Dev Mode's CSS panel — scriptable,
batchable, and CI-safe. Semantic token names (--Base-Primary) are only available
via Variables (Enterprise); use 'figma tokens' for the color palette.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("id")
		outputPath, _ := cmd.Flags().GetString("output")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			cli.Die(fmt.Errorf("css requires --id or a Figma URL with node-id"))
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}
		doc, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
		if err != nil {
			cli.Die(err)
		}
		rules := extract.ExtractCSSRules(doc)
		out := extract.FormatCSSRules(rules)
		if outputPath != "" {
			if err := os.WriteFile(outputPath, []byte(out), 0o644); err != nil {
				cli.Die(err)
			}
			return
		}
		fmt.Print(out)
	},
}

func init() {
	cssCmd.Flags().String("id", "", "node ID to inspect; accepts 20089:685897 or 20089-685897")
	cssCmd.Flags().String("output", "", "write CSS to a file instead of stdout")
	rootCmd.AddCommand(cssCmd)
}
