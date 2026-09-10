package cmd

import (
	"fmt"
	"math"
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var exportCmd = newExportCommand(cli.LoadClient, nil)

const (
	exportFormatPNG = "png"
	exportFormatJPG = "jpg"
	exportFormatSVG = "svg"
)

func newExportCommand(loadClient func() (*figma.Client, error), downloadClient *http.Client) *cobra.Command {
	command := &cobra.Command{
		Use:   "export [figma-url-with-node-id]",
		Short: "Export a Figma node asset",
		Example: `  figma export "<url>?node-id=42-1"
  figma export --id 42:1 --format svg <file-key>
  figma export --id 42:1 --width 1200 --output screen.png <file-key>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			if err := figma.ValidateExportFormat(format); err != nil {
				return cli.NewUsageError(err)
			}
			scale, _ := cmd.Flags().GetFloat64("scale")
			requestedWidth, _ := cmd.Flags().GetFloat64("width")
			if cmd.Flags().Changed("width") && cmd.Flags().Changed("scale") {
				return cli.NewUsageError(fmt.Errorf("--width cannot be combined with --scale"))
			}
			if err := validateExportWidth(format, requestedWidth); err != nil {
				return cli.NewUsageError(err)
			}
			if requestedWidth == 0 {
				if err := validateExportScale(format, scale); err != nil {
					return cli.NewUsageError(err)
				}
			}
			outputPath, _ := cmd.Flags().GetString("output")
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			resolvedNodeID, err := figma.ResolveSingleNodeID(input, nodeID, "export")
			if err != nil {
				return cli.NewUsageError(err)
			}
			if outputPath == "" {
				outputPath = assets.DefaultExportOutputPath(input.FileID, resolvedNodeID, format)
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			if requestedWidth > 0 {
				computedScale, err := exportScaleForWidth(client, input.FileID, resolvedNodeID, requestedWidth)
				if err != nil {
					return err
				}
				scale = computedScale
				if err := validateExportScale(format, scale); err != nil {
					return err
				}
			}
			apiURL, err := figma.BuildExportURL(input.FileID, []string{resolvedNodeID}, format, scale)
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
			if err := assets.DownloadFile(exportDownloadClient, outputPath, assetURL); err != nil {
				return err
			}
			result := map[string]any{"format": format, "node": resolvedNodeID, "scale": scale}
			if requestedWidth > 0 {
				result["requestedWidth"] = requestedWidth
			}
			if err := cli.NewPrinter(cmd).File(outputPath, result); err != nil {
				return err
			}
			return nil
		},
	}
	command.Flags().String("format", "png", "export format: png, jpg, svg, or pdf")
	addNodeIDFlag(command, "node ID to export; defaults to URL node-id")
	command.Flags().StringP("output", "o", "", "output file path; defaults to <file-key>_<node-id>.<format>")
	command.Flags().Float64("scale", 1, "raster export scale for png/jpg (0.01-4)")
	command.Flags().Float64("width", 0, "target raster export width in pixels; derives scale (png/jpg)")
	return command
}

func validateExportWidth(format string, width float64) error {
	if width == 0 {
		return nil
	}
	if math.IsNaN(width) || math.IsInf(width, 0) || width <= 0 {
		return fmt.Errorf("--width must be a finite positive number")
	}
	if format != exportFormatPNG && format != exportFormatJPG {
		return fmt.Errorf("--width is only supported for png and jpg exports")
	}
	return nil
}

func validateExportScale(format string, scale float64) error {
	if math.IsNaN(scale) || math.IsInf(scale, 0) || scale < 0.01 || scale > 4 {
		return fmt.Errorf("--scale must be a finite number between 0.01 and 4")
	}
	if scale != 1 && format != exportFormatPNG && format != exportFormatJPG {
		return fmt.Errorf("--scale is only supported for png and jpg exports")
	}
	return nil
}

func exportScaleForWidth(client *figma.Client, fileID, nodeID string, requestedWidth float64) (float64, error) {
	details, err := figma.FetchNodeDetails(client, fileID, []string{nodeID})
	if err != nil {
		return 0, err
	}
	if len(details.Documents) == 0 {
		return 0, fmt.Errorf("node %s was not returned by Figma", nodeID)
	}
	node, _ := details.Documents[0].(map[string]any)
	bounds, _ := node["absoluteBoundingBox"].(map[string]any)
	width := numberFromAny(bounds["width"])
	if width <= 0 {
		return 0, fmt.Errorf("node %s has no usable width for --width", nodeID)
	}
	return requestedWidth / width, nil
}

func numberFromAny(value any) float64 {
	number, _ := value.(float64)
	return number
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
