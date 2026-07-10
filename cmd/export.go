package cmd

import (
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var exportCmd = newExportCommand(cli.LoadClient, nil)

func newExportCommand(loadClient func() (*figma.Client, error), downloadClient *http.Client) *cobra.Command {
	command := &cobra.Command{
		Use:   "export [figma-url-with-node-id]",
		Short: "Export a Figma node asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			if err := figma.ValidateExportFormat(format); err != nil {
				return err
			}
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
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			apiURL, err := figma.BuildExportURL(input.FileID, []string{resolvedNodeID}, format)
			if err != nil {
				return err
			}
			assetURL, err := figma.FetchExportURL(client, apiURL, resolvedNodeID)
			if err != nil {
				return err
			}
			exportDownloadClient := downloadClient
			if exportDownloadClient == nil {
				exportDownloadClient = client.HTTP
			}
			if err := cli.DownloadFile(exportDownloadClient, outputPath, assetURL); err != nil {
				return err
			}
			if err := cli.NewPrinter(cmd).File(outputPath, map[string]any{"format": format, "node": resolvedNodeID}); err != nil {
				return err
			}
			return nil
		},
	}
	command.Flags().String("format", "png", "export format: png, jpg, svg, or pdf")
	command.Flags().String("id", "", "node ID to export; defaults to URL node-id")
	command.Flags().StringP("output", "o", "", "output file path; defaults to <file-key>_<node-id>.<format>")
	return command
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
