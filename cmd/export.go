package cmd

import (
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [figma-url-with-node-id]",
	Short: "Export a Figma node asset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		outputPath, _ := cmd.Flags().GetString("output")
		nodeID, _ := cmd.Flags().GetString("id")
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		resolvedNodeID, err := figma.ResolveSingleNodeID(input, nodeID, "export")
		if err != nil {
			return err
		}
		if outputPath == "" {
			outputPath = cli.DefaultExportOutputPath(input.FileID, resolvedNodeID, format)
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		apiURL, err := figma.BuildExportURL(input.FileID, []string{resolvedNodeID}, format)
		if err != nil {
			return err
		}
		assetURL, err := figma.FetchExportURL(client, apiURL, resolvedNodeID)
		if err != nil {
			return err
		}
		if err := cli.DownloadFile(http.DefaultClient, outputPath, assetURL); err != nil {
			return err
		}
		if err := cli.NewPrinter(cmd).File(outputPath, map[string]any{"format": format, "node": resolvedNodeID}); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	exportCmd.Flags().String("format", "png", "export format: png, jpg, svg, or pdf")
	exportCmd.Flags().String("id", "", "node ID to export; defaults to URL node-id")
	exportCmd.Flags().StringP("output", "o", "", "output file path; defaults to <file-key>_<node-id>.<format>")
	rootCmd.AddCommand(exportCmd)
}
