package cmd

import (
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/cristianoliveira/figma-cli/internal/assets"
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
			scale, _ := cmd.Flags().GetFloat64("scale")
			if err := validateExportScale(format, scale); err != nil {
				return err
			}
			outputPath, _ := cmd.Flags().GetString("output")
			metadataPath, _ := cmd.Flags().GetString("metadata")
			if metadataPath != "" && sameExportPath(metadataPath, outputPath) {
				return fmt.Errorf("--metadata must differ from --output")
			}
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			resolvedNodeID, err := figma.ResolveSingleNodeID(input, nodeID, "export")
			if err != nil {
				return err
			}
			if outputPath == "" {
				outputPath = assets.DefaultExportOutputPath(input.FileID, resolvedNodeID, format)
			}
			if metadataPath != "" && sameExportPath(metadataPath, outputPath) {
				return fmt.Errorf("--metadata must differ from --output")
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
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
			metadata := map[string]any{"format": format, "node": resolvedNodeID, "scale": scale}
			if metadataPath != "" {
				if err := writeExportMetadata(client, input.FileID, resolvedNodeID, format, scale, outputPath, metadataPath); err != nil {
					return fmt.Errorf("exported %s but failed to write metadata: %w", outputPath, err)
				}
				metadata["metadata"] = metadataPath
			}
			if err := cli.NewPrinter(cmd).File(outputPath, metadata); err != nil {
				return err
			}
			return nil
		},
	}
	command.Flags().String("format", "png", "export format: png, jpg, svg, or pdf")
	addNodeIDFlag(command, "node ID to export; defaults to URL node-id")
	command.Flags().StringP("output", "o", "", "output file path; defaults to <file-key>_<node-id>.<format>")
	command.Flags().Float64("scale", 1, "raster export scale for png/jpg (0.01-4)")
	command.Flags().String("metadata", "", "write export metadata sidecar JSON to this path")
	return command
}

type exportMetadata struct {
	Version         int            `json:"version"`
	NodeID          string         `json:"nodeId"`
	Format          string         `json:"format"`
	Scale           float64        `json:"scale"`
	NodeBounds      exportBounds   `json:"nodeBounds"`
	ExportBounds    exportSize     `json:"exportBounds"`
	DimensionDelta  exportSize     `json:"dimensionDelta"`
	LogicalCrop     *exportBounds  `json:"logicalCrop,omitempty"`
	ExportPadding   *exportPadding `json:"exportPadding,omitempty"`
	PaddingEvidence []string       `json:"paddingEvidence,omitempty"`
	Output          string         `json:"output"`
}

type exportBounds struct {
	X      float64 `json:"x,omitempty"`
	Y      float64 `json:"y,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

type exportSize struct {
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

type exportPadding struct {
	Left   float64 `json:"left"`
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
}

func validateExportScale(format string, scale float64) error {
	if math.IsNaN(scale) || math.IsInf(scale, 0) || scale < 0.01 || scale > 4 {
		return fmt.Errorf("--scale must be a finite number between 0.01 and 4")
	}
	if scale != 1 && format != "png" && format != "jpg" {
		return fmt.Errorf("--scale is only supported for png and jpg exports")
	}
	return nil
}

func writeExportMetadata(client *figma.Client, fileID, nodeID, format string, scale float64, outputPath, metadataPath string) error {
	details, err := figma.FetchNodeDetails(client, fileID, []string{nodeID})
	if err != nil {
		return err
	}
	node, _ := details.Documents[0].(map[string]any)
	nodeBounds := exportBoundsFromValue(node["absoluteBoundingBox"])
	exportBounds, err := measureExportBounds(outputPath, format)
	if err != nil {
		return err
	}
	logicalCrop, padding := measureLogicalCrop(outputPath, format, nodeBounds, exportBounds, scale)
	if logicalCrop == nil && format != "svg" {
		logicalCrop, padding = measureLogicalCropFromSVGExport(client, fileID, nodeID, nodeBounds, exportBounds, scale)
	}
	metadata := exportMetadata{
		Version:         1,
		NodeID:          nodeID,
		Format:          format,
		Scale:           scale,
		NodeBounds:      nodeBounds,
		ExportBounds:    exportBounds,
		DimensionDelta:  exportSize{Width: exportBounds.Width - nodeBounds.Width, Height: exportBounds.Height - nodeBounds.Height},
		LogicalCrop:     logicalCrop,
		ExportPadding:   padding,
		PaddingEvidence: exportPaddingEvidence(node["effects"]),
		Output:          outputPath,
	}
	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metadataPath, append(encoded, '\n'), 0o600)
}

func exportBoundsFromValue(value any) exportBounds {
	object, _ := value.(map[string]any)
	return exportBounds{
		X:      numberFromAny(object["x"]),
		Y:      numberFromAny(object["y"]),
		Width:  numberFromAny(object["width"]),
		Height: numberFromAny(object["height"]),
	}
}

func measureExportBounds(path, format string) (exportSize, error) {
	if format == "png" {
		file, err := os.Open(path)
		if err != nil {
			return exportSize{}, err
		}
		config, err := png.DecodeConfig(file)
		closeErr := file.Close()
		if err != nil {
			return exportSize{}, err
		}
		if closeErr != nil {
			return exportSize{}, closeErr
		}
		return exportSize{Width: float64(config.Width), Height: float64(config.Height)}, nil
	}
	if format == "svg" {
		data, err := os.ReadFile(path)
		if err != nil {
			return exportSize{}, err
		}
		return svgSize(data)
	}
	return exportSize{}, nil
}

func svgSize(data []byte) (exportSize, error) {
	width, err := svgAttributeNumber(data, "width")
	if err != nil {
		return exportSize{}, err
	}
	height, err := svgAttributeNumber(data, "height")
	if err != nil {
		return exportSize{}, err
	}
	return exportSize{Width: width, Height: height}, nil
}

func svgAttributeNumber(data []byte, name string) (float64, error) {
	pattern := regexp.MustCompile(name + `="([0-9.]+)"`)
	matches := pattern.FindSubmatch(data)
	if len(matches) != 2 {
		return 0, fmt.Errorf("svg %s attribute not found", name)
	}
	return strconv.ParseFloat(string(matches[1]), 64)
}

func measureLogicalCropFromSVGExport(client *figma.Client, fileID, nodeID string, nodeBounds exportBounds, measuredBounds exportSize, scale float64) (*exportBounds, *exportPadding) {
	apiURL, err := figma.BuildExportURL(fileID, []string{nodeID}, "svg", 1)
	if err != nil {
		return nil, nil
	}
	assetURL, err := figma.FetchExportURL(client, apiURL, nodeID)
	if err != nil {
		return nil, nil
	}
	tempFile, err := os.CreateTemp("", "figma-export-metadata-*.svg")
	if err != nil {
		return nil, nil
	}
	path := tempFile.Name()
	_ = tempFile.Close()
	defer func() { _ = os.Remove(path) }()
	if err := assets.DownloadFile(client.HTTP, path, assetURL); err != nil {
		return nil, nil
	}
	return measureLogicalCrop(path, "svg", nodeBounds, measuredBounds, scale)
}

func measureLogicalCrop(path, format string, nodeBounds exportBounds, measuredBounds exportSize, scale float64) (*exportBounds, *exportPadding) {
	if format != "svg" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil
	}
	left, top, ok := svgPathStart(data)
	if !ok {
		return nil, nil
	}
	left *= scale
	top *= scale
	crop := &exportBounds{X: left, Y: top, Width: nodeBounds.Width * scale, Height: nodeBounds.Height * scale}
	padding := &exportPadding{
		Left:   left,
		Top:    top,
		Right:  measuredBounds.Width - left - crop.Width,
		Bottom: measuredBounds.Height - top - crop.Height,
	}
	return crop, padding
}

func svgPathStart(data []byte) (float64, float64, bool) {
	pathPattern := regexp.MustCompile(`<path[^>]+d="M\s*(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)`)
	matches := pathPattern.FindSubmatch(data)
	if len(matches) != 3 {
		return 0, 0, false
	}
	x, xErr := strconv.ParseFloat(string(matches[1]), 64)
	y, yErr := strconv.ParseFloat(string(matches[2]), 64)
	if xErr != nil || yErr != nil {
		return 0, 0, false
	}
	return x, y, true
}

func exportPaddingEvidence(value any) []string {
	effects, _ := value.([]any)
	evidence := make([]string, 0, len(effects))
	for _, item := range effects {
		effect, _ := item.(map[string]any)
		if visible, ok := effect["visible"].(bool); ok && !visible {
			continue
		}
		typeName, _ := effect["type"].(string)
		if typeName == "" {
			continue
		}
		offset, _ := effect["offset"].(map[string]any)
		evidence = append(evidence, fmt.Sprintf("%s radius=%g offsetX=%g offsetY=%g", typeName, numberFromAny(effect["radius"]), numberFromAny(offset["x"]), numberFromAny(offset["y"])))
	}
	return evidence
}

func numberFromAny(value any) float64 {
	number, _ := value.(float64)
	return number
}

func sameExportPath(first, second string) bool {
	if first == "" || second == "" {
		return false
	}
	return filepath.Clean(first) == filepath.Clean(second)
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
