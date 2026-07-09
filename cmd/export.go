package cmd

import (
	"fmt"
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [figma-url-with-node-id]",
	Short: "Export a Figma node asset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		outputPath, _ := cmd.Flags().GetString("output")
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		nodeIDs := figma.ResolveNodeIDs(input, nodeID)
		if len(nodeIDs) == 0 {
			cli.Die(fmt.Errorf("export requires --id or a Figma URL with node-id"))
		}
		if outputPath == "" {
			outputPath = cli.DefaultExportOutputPath(input.FileID, nodeIDs[0], format)
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}
		apiURL, err := figma.BuildExportURL(input.FileID, []string{nodeIDs[0]}, format)
		if err != nil {
			cli.Die(err)
		}
		assetURL, err := figma.FetchExportURL(client, apiURL, nodeIDs[0])
		if err != nil {
			cli.Die(err)
		}
		if err := cli.DownloadFile(http.DefaultClient, outputPath, assetURL); err != nil {
			cli.Die(err)
		}
		if err := cli.NewPrinter(cmd).File(outputPath, map[string]any{"format": format, "node": nodeIDs[0]}); err != nil {
			cli.Die(err)
		}
	},
}

func init() {
	exportCmd.Flags().String("format", "png", "export format: png, jpg, svg, or pdf")
	exportCmd.Flags().String("id", "", "node ID to export; accepts 20089:685897 or 20089-685897")
	exportCmd.Flags().StringP("output", "o", "", "output file path; defaults to <file-key>_<node-id>.<format>")
	rootCmd.AddCommand(exportCmd)
}
