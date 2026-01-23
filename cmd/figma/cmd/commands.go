package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(textCmd)
	rootCmd.AddCommand(exportCmd)
}

var parseCmd = &cobra.Command{
	Use:   "parse <figma-url>",
	Short: "Parse a Figma URL and extract metadata",
	Long: `Parse a Figma URL and extract file metadata, node information, and design structure.

The command accepts any valid Figma URL (design, file, or node) and returns structured
JSON output with metadata about the file and the specified node.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.PrintErrln("Command 'parse' is not yet implemented.")
		cmd.Println("This command will parse Figma URL:", args[0])
		return nil
	},
}

var nodesCmd = &cobra.Command{
	Use:   "nodes <figma-url>",
	Short: "Fetch node hierarchy for a design",
	Long: `Fetch the node hierarchy for a Figma design, showing parent/child relationships
and layer structure.

Use the --hierarchy flag to get the full tree of nodes.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.PrintErrln("Command 'nodes' is not yet implemented.")
		cmd.Println("This command will fetch nodes for Figma URL:", args[0])
		return nil
	},
}

var textCmd = &cobra.Command{
	Use:   "text <figma-url>",
	Short: "Extract text layers from a design",
	Long: `Extract all text layers from a Figma design, including text content, styles,
and bounding box information.

The command returns a list of text layers with their properties.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.PrintErrln("Command 'text' is not yet implemented.")
		cmd.Println("This command will extract text layers from Figma URL:", args[0])
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export <figma-url>",
	Short: "Export assets from a frame",
	Long: `Export assets (images, SVGs, etc.) from a Figma frame or node.

Specify output format, scale, and directory using flags.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.PrintErrln("Command 'export' is not yet implemented.")
		cmd.Println("This command will export assets from Figma URL:", args[0])
		return nil
	},
}
