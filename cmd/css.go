package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var cssCmd = newCSSCommand(cli.LoadClient)

func newCSSCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "css [figma-url-or-file-id]",
		Short: "Generate CSS rules from a Figma element's layout and styles",
		Example: `  figma css "<url>?node-id=42-1"
  figma css --id 42:1 --recursive <file-key>
  figma css --id 42:1 --output component.css <file-key>`,
		Long: `Generate CSS from a Figma node tree.

Emits CSS for the selected node. With --recursive, walks its subtree and emits
one rule per node that contributes a meaningful property (autolayout, fills,
borders, radius, text). Layout maps to
flexbox: layoutMode -> display:flex, itemSpacing -> gap, paddings -> shorthand,
sizing modes -> width/height. Output is deterministic (sorted properties).

This is the public-API equivalent of Figma Dev Mode's CSS panel — scriptable,
batchable, and CI-safe. Semantic token names (--Base-Primary) are only available
via Variables (Enterprise); use 'figma tokens' for the color palette.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return err
			}
			outputPath, _ := cmd.Flags().GetString("output")
			recursive, _ := cmd.Flags().GetBool("recursive")
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			nodeIDs, err := figma.ResolveRequiredNodeIDs(input, nodeID, "css")
			if err != nil {
				return err
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
			if err != nil {
				return err
			}
			var rules []extract.CSSRule
			for _, document := range documents {
				rules = append(rules, extract.ExtractCSSRules(document, recursive)...)
			}
			out := extract.FormatCSSRules(rules)
			if outputPath != "" {
				if err := output.WriteFile(outputPath, []byte(out), 0o644); err != nil {
					return err
				}
				if err := cli.NewPrinter(cmd).File(outputPath, map[string]any{"format": "css", "bytes": len(out)}); err != nil {
					return err
				}
				return nil
			}
			if err := cli.NewPrinter(cmd).Text("css", out); err != nil {
				return err
			}
			return nil
		},
	}
	addNodeIDFlag(command, "node ID to inspect; defaults to URL node-id")
	command.Flags().String("output", "", "write CSS to a file instead of stdout")
	command.Flags().Bool("recursive", false, "include CSS rules from all descendant nodes")
	return command
}

func init() {
	rootCmd.AddCommand(cssCmd)
}
