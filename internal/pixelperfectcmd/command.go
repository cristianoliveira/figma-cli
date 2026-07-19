package pixelperfectcmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/annotations"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/imagecontext"
	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
	outputpkg "github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/cristianoliveira/figma-cli/internal/pixelperfectreport"
	"github.com/spf13/cobra"
)

type imageComparer func(referencePath, actualPath, maskPath string, threshold uint8, perceptualThreshold float64, region *diff.Bounds, ignored []diff.Bounds) (diff.ImageComparison, error)

type preparedImageInputs struct {
	referencePath string
	actualPath    string
	metadata      *diff.ImageInputs
	cleanup       func()
}

type exportMetadata struct {
	Version      int                   `json:"version"`
	NodeBounds   exportMetadataSize    `json:"nodeBounds"`
	ExportBounds exportMetadataSize    `json:"exportBounds"`
	LogicalCrop  *exportMetadataBounds `json:"logicalCrop"`
}

type exportMetadataBounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type exportMetadataSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type scanOutput struct {
	Axis      string            `json:"axis"`
	Index     int               `json:"index"`
	Length    int               `json:"length"`
	Reference []scanRun         `json:"reference"`
	Actual    []scanRun         `json:"actual"`
	Inputs    *diff.ImageInputs `json:"inputs,omitempty"`
	InputLine *scanInputLine    `json:"inputLine,omitempty"`
}

type scanInputLine struct {
	Reference scanLinePosition `json:"reference"`
	Actual    scanLinePosition `json:"actual"`
}

type scanLinePosition struct {
	Axis  string `json:"axis"`
	Index int    `json:"index"`
}

type scanRun struct {
	Start  int      `json:"start"`
	End    int      `json:"end"`
	Length int      `json:"length"`
	RGBA   [4]uint8 `json:"rgba"`
	Hex    string   `json:"hex"`
}

type probeOutput struct {
	Points []probePointOutput `json:"points"`
	Inputs *diff.ImageInputs  `json:"inputs,omitempty"`
}

type probePointOutput struct {
	Point      probePoint       `json:"point"`
	Reference  probeColor       `json:"reference"`
	Actual     probeColor       `json:"actual"`
	Delta      probeDelta       `json:"delta"`
	InputPoint *probeInputPoint `json:"inputPoint,omitempty"`
}

type probeInputPoint struct {
	Reference probePoint `json:"reference"`
	Actual    probePoint `json:"actual"`
}

type probePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type probeColor struct {
	RGBA [4]uint8 `json:"rgba"`
	Hex  string   `json:"hex"`
}

type probeDelta struct {
	R int `json:"r"`
	G int `json:"g"`
	B int `json:"b"`
	A int `json:"a"`
}

func newCommand(compare imageComparer) *cobra.Command {
	command := &cobra.Command{
		Use:   "image <reference.png> <actual.png>",
		Short: "Compare equal-sized PNGs and write a changed-pixel mask",
		Args:  requireImagePair("compare", "pixel-perfect reference.png actual.png"),
		RunE: func(cmd *cobra.Command, args []string) error {
			configuration, err := applyComparisonProfile(cmd)
			if err != nil {
				return err
			}
			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = defaultMaskPath(args[1])
			}
			threshold, _ := cmd.Flags().GetUint8("threshold")
			overlay, _ := cmd.Flags().GetString("overlay")
			report, _ := cmd.Flags().GetString("report")
			if output != "" && (samePath(output, args[0]) || samePath(output, args[1])) {
				return fmt.Errorf("--output must not overwrite an input image")
			}
			if overlay != "" && (samePath(overlay, args[0]) || samePath(overlay, args[1])) {
				return fmt.Errorf("--overlay must not overwrite an input image")
			}
			if overlay != "" && output != "" && samePath(overlay, output) {
				return fmt.Errorf("--overlay must differ from --output")
			}
			if report != "" && (samePath(report, args[0]) || samePath(report, args[1]) || (output != "" && samePath(report, output)) || samePath(report, overlay)) {
				return fmt.Errorf("--report must not overwrite an input, mask, or overlay")
			}
			offsetRadius, _ := cmd.Flags().GetInt("suggest-offset")
			if offsetRadius < 0 {
				return fmt.Errorf("--suggest-offset must be non-negative")
			}
			movementRadius, _ := cmd.Flags().GetInt("suggest-movement")
			if movementRadius < 0 {
				return fmt.Errorf("--suggest-movement must be non-negative")
			}
			regionGap, _ := cmd.Flags().GetInt("region-gap")
			if regionGap < 0 {
				return fmt.Errorf("--region-gap must be non-negative")
			}
			minRegionPixels, _ := cmd.Flags().GetInt("min-region-pixels")
			if minRegionPixels < 1 {
				return fmt.Errorf("--min-region-pixels must be positive")
			}
			perceptualThreshold, _ := cmd.Flags().GetFloat64("perceptual-threshold")
			if perceptualThreshold < 0 || math.IsNaN(perceptualThreshold) || math.IsInf(perceptualThreshold, 0) {
				return fmt.Errorf("--perceptual-threshold must be a finite non-negative number")
			}
			maxRMSE, _ := cmd.Flags().GetFloat64("max-rmse")
			if maxRMSE != -1 && (maxRMSE < 0 || math.IsNaN(maxRMSE) || math.IsInf(maxRMSE, 0)) {
				return fmt.Errorf("--max-rmse must be -1 or a finite non-negative number")
			}
			maxChangedRatio, _ := cmd.Flags().GetFloat64("max-changed-ratio")
			if maxChangedRatio != -1 && (maxChangedRatio < 0 || maxChangedRatio > 1 || math.IsNaN(maxChangedRatio) || math.IsInf(maxChangedRatio, 0)) {
				return fmt.Errorf("--max-changed-ratio must be -1 or between 0 and 1")
			}
			maxPerceptualChangedRatio, _ := cmd.Flags().GetFloat64("max-perceptual-changed-ratio")
			if maxPerceptualChangedRatio != -1 && (maxPerceptualChangedRatio < 0 || maxPerceptualChangedRatio > 1 || math.IsNaN(maxPerceptualChangedRatio) || math.IsInf(maxPerceptualChangedRatio, 0)) {
				return fmt.Errorf("--max-perceptual-changed-ratio must be -1 or between 0 and 1")
			}
			region, err := parseImageRegion(cmd.Flags().Lookup("region").Value.String())
			if err != nil {
				return err
			}
			ignoredValues, _ := cmd.Flags().GetStringArray("ignore-region")
			ignored := make([]diff.Bounds, 0, len(ignoredValues))
			for _, value := range ignoredValues {
				ignoredRegion, parseErr := parseImageRegion(value)
				if parseErr != nil {
					return fmt.Errorf("invalid --ignore-region: %w", parseErr)
				}
				if ignoredRegion.Width < 1 || ignoredRegion.Height < 1 {
					return fmt.Errorf("invalid --ignore-region: width and height must be positive")
				}
				ignored = append(ignored, *ignoredRegion)
			}
			referenceCrop, err := parseOptionalCrop(cmd, "reference-crop")
			if err != nil {
				return err
			}
			referenceMetadataPath, _ := cmd.Flags().GetString("reference-metadata")
			referenceMetadata, err := loadExportMetadata(referenceMetadataPath)
			if err != nil {
				return err
			}
			if referenceCrop != nil && referenceMetadata != nil {
				return fmt.Errorf("--reference-crop and --reference-metadata cannot be used together")
			}
			actualCrop, err := parseOptionalCrop(cmd, "actual-crop")
			if err != nil {
				return err
			}
			inputs, err := prepareImageInputs(args[0], args[1], referenceCrop, actualCrop, referenceMetadata)
			if err != nil {
				return err
			}
			defer inputs.cleanup()
			comparisonMask, _ := cmd.Flags().GetString("mask")
			if comparisonMask != "" {
				maskedRegions, maskErr := diff.IgnoredRegionsFromMask(comparisonMask, inputs.referencePath)
				if maskErr != nil {
					return maskErr
				}
				ignored = append(ignored, maskedRegions...)
			}
			var decoded *diff.DecodedImages
			var result diff.ImageComparison
			if compare == nil {
				decoded, err = diff.LoadDecodedImages(inputs.referencePath, inputs.actualPath)
				if err != nil {
					return err
				}
				result, err = decoded.Compare(output, threshold, perceptualThreshold, region, ignored)
			} else {
				result, err = compare(inputs.referencePath, inputs.actualPath, output, threshold, perceptualThreshold, region, ignored)
			}
			if err != nil {
				return err
			}
			result.Inputs = inputs.metadata
			annotationsPath, _ := cmd.Flags().GetString("annotations")
			var annotationDocument *annotations.Document
			if annotationsPath != "" {
				annotationDocument, err = annotations.Load(annotationsPath)
				if err != nil {
					return err
				}
				imageWidth, imageHeight, dimensionsErr := diff.PNGDimensions(inputs.referencePath)
				if dimensionsErr != nil {
					return dimensionsErr
				}
				if err := annotationDocument.ValidateDimensions(imageWidth, imageHeight); err != nil {
					return err
				}
			}
			if overlay != "" {
				if decoded == nil {
					decoded, err = diff.LoadDecodedImages(inputs.referencePath, inputs.actualPath)
					if err != nil {
						return err
					}
				}
				if err := decoded.WriteOverlay(overlay, region, ignored); err != nil {
					return err
				}
				result.Overlay = overlay
			}
			if offsetRadius > 0 {
				if decoded == nil {
					decoded, err = diff.LoadDecodedImages(inputs.referencePath, inputs.actualPath)
					if err != nil {
						return err
					}
				}
				suggestedOffset := decoded.SuggestOffset(offsetRadius, region, ignored)
				if !math.IsInf(suggestedOffset.RMSE, 0) && !math.IsNaN(suggestedOffset.RMSE) {
					result.SuggestedOffset = &suggestedOffset
				}
			}
			result.Regions = groupImageRegions(result.Regions, regionGap)
			result.Regions = filterImageRegions(result.Regions, minRegionPixels)
			if len(result.Regions) > 20 {
				result.Regions = result.Regions[:20]
			}
			regionMetrics := make([]diff.RegionMetrics, len(result.Regions))
			if len(result.Regions) > 0 {
				regionBounds := make([]diff.Bounds, len(result.Regions))
				for index := range result.Regions {
					regionBounds[index] = result.Regions[index].Bounds
				}
				if decoded == nil {
					decoded, err = diff.LoadDecodedImages(inputs.referencePath, inputs.actualPath)
					if err != nil {
						return err
					}
				}
				regionMetrics, err = decoded.MeasureRegions(regionBounds, threshold, perceptualThreshold, ignored)
				if err != nil {
					return err
				}
			}
			for index := range result.Regions {
				result.Regions[index].InputBounds = inputBounds(result.Regions[index].Bounds, inputs.metadata)
				if annotationDocument != nil {
					bounds := result.Regions[index].Bounds
					result.Regions[index].Annotations = annotationDocument.Intersections(annotations.Bounds{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: bounds.Height})
				}
				metrics := regionMetrics[index]
				result.Regions[index].ChangedPixels = metrics.ChangedPixels
				result.Regions[index].ChangedRatio = metrics.ChangedRatio
				result.Regions[index].RMSE = metrics.RMSE
				result.Regions[index].EdgeRMSE = metrics.EdgeRMSE
				result.Regions[index].PerceptualRMSE = metrics.PerceptualRMSE
				result.Regions[index].PerceptualChangedPixels = metrics.PerceptualChangedPixels
				result.Regions[index].PerceptualChangedRatio = metrics.PerceptualChangedRatio
				result.Regions[index].AntialiasedPixels = metrics.AntialiasedPixels
				result.Regions[index].DominantColorPairs = metrics.DominantColorPairs
				result.Regions[index].Classification = diff.ClassifyImageRegion(metrics)
			}
			if movementRadius > 0 && len(result.Regions) > 0 {
				regionBounds := make([]diff.Bounds, len(result.Regions))
				for index := range result.Regions {
					regionBounds[index] = result.Regions[index].Bounds
				}
				result.MovedRegions = decoded.SuggestRegionMovements(regionBounds, movementRadius, ignored)
			}
			if maxRMSE >= 0 && result.RMSE > maxRMSE {
				return fmt.Errorf("image diff validation failed: RMSE %.6f exceeds maximum %.6f", result.RMSE, maxRMSE)
			}
			if maxChangedRatio >= 0 && result.ChangedRatio > maxChangedRatio {
				return fmt.Errorf("image diff validation failed: changed ratio %.6f exceeds maximum %.6f", result.ChangedRatio, maxChangedRatio)
			}
			if maxPerceptualChangedRatio >= 0 && result.PerceptualChangedRatio > maxPerceptualChangedRatio {
				return fmt.Errorf("image diff validation failed: perceptual changed ratio %.6f exceeds maximum %.6f", result.PerceptualChangedRatio, maxPerceptualChangedRatio)
			}
			if report != "" {
				if err := pixelperfectreport.Write(report, pixelperfectreport.Input{
					ReferencePath:       inputs.referencePath,
					ActualPath:          inputs.actualPath,
					MaskPath:            output,
					OverlayPath:         overlay,
					Threshold:           threshold,
					PerceptualThreshold: perceptualThreshold,
					ComparedRegion:      region,
					Result:              result,
				}); err != nil {
					return err
				}
			}
			outputResult := outputEnvelope{ImageComparison: result, Configuration: configuration}
			visualContextEnabled, _ := cmd.Flags().GetBool("visual-context")
			if visualContextEnabled {
				provider, _ := cmd.Flags().GetString("visual-context-provider")
				model, _ := cmd.Flags().GetString("visual-context-model")
				visualContextPrompt, _ := cmd.Flags().GetString("visual-context-prompt")
				config, configErr := imagecontext.LoadProviderConfig(provider, model)
				if errors.Is(configErr, imagecontext.ErrNotConfigured) {
					outputResult.VisualContext = &imagecontext.Result{Provider: provider, Advisory: true, Disclaimer: fmt.Sprintf("Visual context unavailable: configure %s credentials in the Pi Spectacles config or environment.", provider)}
					return writeJSON(cmd, outputResult)
				}
				if configErr != nil {
					return configErr
				}
				regions := make([]imagecontext.Region, len(result.Regions))
				for index, region := range result.Regions {
					regions[index] = imagecontext.Region{ID: fmt.Sprintf("r%d", index+1), Bounds: imagecontext.Bounds{X: region.Bounds.X, Y: region.Bounds.Y, Width: region.Bounds.Width, Height: region.Bounds.Height}}
				}
				client, clientErr := imagecontext.NewClient(provider, config)
				if clientErr != nil {
					return clientErr
				}
				input := imagecontext.Input{ReferencePath: inputs.referencePath, ActualPath: inputs.actualPath, Regions: regions, Prompt: visualContextPrompt}
				visualContext, explainErr := client.Describe(context.Background(), input)
				if explainErr != nil {
					return explainErr
				}
				outputResult.VisualContext = &visualContext
			}
			return writeJSON(cmd, outputResult)
		},
	}
	command.Flags().String("profile", "", "load comparison options from a versioned JSON profile; explicit flags override profile values")
	command.Flags().String("annotations", "", "enrich mismatch regions from a generic coordinate annotation JSON file")
	command.Flags().StringP("output", "o", "", "path for transparent PNG difference mask; defaults to <actual>.diff.png")
	command.Flags().Uint8("threshold", 0, "ignore per-channel differences at or below this value (0-255)")
	command.Flags().Float64("perceptual-threshold", diff.DefaultPerceptualThreshold, "OKLab HyAB distance above which a pixel is perceptually changed (non-negative)")
	command.Flags().String("region", "", "compare only x,y,width,height")
	command.Flags().String("reference-crop", "", "crop reference before comparing: x,y,width,height")
	command.Flags().String("reference-metadata", "", "apply logical crop from figma export metadata JSON")
	command.Flags().String("actual-crop", "", "crop actual before comparing: x,y,width,height")
	command.Flags().StringArray("ignore-region", nil, "exclude x,y,width,height; repeat for multiple areas")
	command.Flags().String("mask", "", "full-size PNG selecting compared pixels (visible non-black includes)")
	command.Flags().String("overlay", "", "path for directional overlay (reference red, actual green)")
	command.Flags().String("report", "", "write a self-contained HTML report to this path")
	command.Flags().Int("suggest-offset", 0, "report best whole-image translation within this pixel radius without applying it")
	command.Flags().Int("suggest-movement", 0, "report advisory per-region translations within this pixel radius without applying them")
	command.Flags().Int("region-gap", 0, "group mismatch regions separated by at most this many pixels")
	command.Flags().Int("min-region-pixels", 1, "omit disconnected regions smaller than this many changed pixels")
	command.Flags().Float64("max-rmse", -1, "fail when normalized RMSE exceeds this value")
	command.Flags().Float64("max-changed-ratio", -1, "fail when changed-pixel ratio exceeds this value")
	command.Flags().Float64("max-perceptual-changed-ratio", -1, "fail when perceptual changed-pixel ratio exceeds this value")
	command.Flags().Bool("visual-context", false, "add advisory visual descriptions using the configured multimodal model")
	command.Flags().String("visual-context-provider", "openrouter", "visual context provider: openrouter or openai")
	command.Flags().String("visual-context-model", "", "override the visual context model")
	command.Flags().String("visual-context-prompt", "", "extra advisory focus for visual context analysis")
	command.AddCommand(newProbeCommand())
	command.AddCommand(newScanCommand())
	return command
}

func newProbeCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "probe <reference.png> <actual.png>",
		Short: "Inspect colors at one pixel in two equal-sized PNGs",
		Args:  requireImagePair("probe", "pixel-perfect probe reference.png actual.png --at 12,24"),
		Example: `  pixel-perfect probe reference.png actual.png --at 12,24
  pixel-perfect probe reference.png actual.png --from 0,20 --to 100,20
  pixel-perfect probe reference.png actual.png --at 12,24 --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			points, err := probePointsFromFlags(cmd)
			if err != nil {
				return err
			}
			inputs, err := prepareCommandImageInputs(cmd, args[0], args[1])
			if err != nil {
				return err
			}
			defer inputs.cleanup()
			output, err := probeImages(inputs.referencePath, inputs.actualPath, points)
			if err != nil {
				return err
			}
			output.Inputs = inputs.metadata
			for index := range output.Points {
				output.Points[index].InputPoint = inputPoint(output.Points[index].Point, inputs.metadata)
			}
			format, err := tabularFormat(cmd)
			if err != nil {
				return err
			}
			if format == outputpkg.FormatJSON {
				return writeJSON(cmd, output)
			}
			return writeProbeCSV(cmd, output)
		},
	}
	command.Flags().StringArray("at", nil, "pixel coordinate to inspect: x,y in comparison/cropped coordinates; repeat for multiple points")
	command.Flags().String("from", "", "inclusive line start: x,y in comparison/cropped coordinates")
	command.Flags().String("to", "", "inclusive line end: x,y in comparison/cropped coordinates")
	command.Flags().Int("step", 1, "sample every Nth point along --from/--to line")
	command.Flags().Int("radius", 0, "include square pixel neighborhood around every selected point")
	addTabularFormatFlag(command)
	addInputPreparationFlags(command)
	return command
}

func newScanCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "scan <reference.png> <actual.png>",
		Short: "Inspect compact color runs along one row or column in two PNGs",
		Args:  requireImagePair("scan", "pixel-perfect scan reference.png actual.png --row 24"),
		Example: `  pixel-perfect scan reference.png actual.png --row 24
  pixel-perfect scan reference.png actual.png --column 12 --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := prepareCommandImageInputs(cmd, args[0], args[1])
			if err != nil {
				return err
			}
			defer inputs.cleanup()
			xChanged := cmd.Flags().Changed("x") || cmd.Flags().Changed("column")
			yChanged := cmd.Flags().Changed("y") || cmd.Flags().Changed("row")
			if xChanged == yChanged {
				return fmt.Errorf("provide exactly one of --x/--column or --y/--row")
			}
			axis := "y"
			index, _ := cmd.Flags().GetInt("x")
			if cmd.Flags().Changed("column") {
				index, _ = cmd.Flags().GetInt("column")
			}
			if yChanged {
				axis = "x"
				index, _ = cmd.Flags().GetInt("y")
				if cmd.Flags().Changed("row") {
					index, _ = cmd.Flags().GetInt("row")
				}
			}
			if index < 0 {
				return fmt.Errorf("--%s must be non-negative", map[string]string{"x": "y", "y": "x"}[axis])
			}
			output, err := scanImages(inputs.referencePath, inputs.actualPath, axis, index)
			if err != nil {
				return err
			}
			output.Inputs = inputs.metadata
			output.InputLine = inputLine(axis, index, inputs.metadata)
			format, err := tabularFormat(cmd)
			if err != nil {
				return err
			}
			if format == outputpkg.FormatJSON {
				return writeJSON(cmd, output)
			}
			return writeScanCSV(cmd, output)
		},
	}
	command.Flags().Int("x", 0, "scan vertical column at x in comparison/cropped coordinates")
	command.Flags().Int("y", 0, "scan horizontal row at y in comparison/cropped coordinates")
	command.Flags().Int("column", 0, "alias for --x")
	command.Flags().Int("row", 0, "alias for --y")
	addTabularFormatFlag(command)
	addInputPreparationFlags(command)
	return command
}

func addTabularFormatFlag(command *cobra.Command) {
	command.Flags().String("format", string(outputpkg.FormatCSV), "output format: csv or json")
}

func tabularFormat(command *cobra.Command) (outputpkg.Format, error) {
	value, _ := command.Flags().GetString("format")
	return outputpkg.ParseFormat(value, outputpkg.FormatCSV, outputpkg.FormatJSON)
}

func inputPoint(point probePoint, inputs *diff.ImageInputs) *probeInputPoint {
	if inputs == nil {
		return nil
	}
	return &probeInputPoint{
		Reference: pointWithCropOrigin(point, inputs.Reference.Crop),
		Actual:    pointWithCropOrigin(point, inputs.Actual.Crop),
	}
}

func pointWithCropOrigin(point probePoint, crop *diff.Bounds) probePoint {
	if crop == nil {
		return point
	}
	return probePoint{X: point.X + crop.X, Y: point.Y + crop.Y}
}

func inputLine(axis string, index int, inputs *diff.ImageInputs) *scanInputLine {
	if inputs == nil {
		return nil
	}
	return &scanInputLine{
		Reference: lineWithCropOrigin(axis, index, inputs.Reference.Crop),
		Actual:    lineWithCropOrigin(axis, index, inputs.Actual.Crop),
	}
}

func lineWithCropOrigin(axis string, index int, crop *diff.Bounds) scanLinePosition {
	position := scanLinePosition{Axis: axis, Index: index}
	if crop == nil {
		return position
	}
	if axis == "x" {
		position.Index += crop.Y
		return position
	}
	position.Index += crop.X
	return position
}

func probePointsFromFlags(command *cobra.Command) ([]probePoint, error) {
	step, _ := command.Flags().GetInt("step")
	if step < 1 {
		return nil, fmt.Errorf("--step must be positive")
	}
	radius, _ := command.Flags().GetInt("radius")
	if radius < 0 {
		return nil, fmt.Errorf("--radius must be non-negative")
	}
	values, _ := command.Flags().GetStringArray("at")
	points, err := parseProbePoints(values)
	if err != nil {
		return nil, err
	}
	fromValue, _ := command.Flags().GetString("from")
	toValue, _ := command.Flags().GetString("to")
	if (fromValue == "") != (toValue == "") {
		return nil, fmt.Errorf("--from and --to must be provided together")
	}
	if fromValue != "" {
		from, err := parseProbePoint(fromValue)
		if err != nil {
			return nil, fmt.Errorf("invalid --from: %w", err)
		}
		to, err := parseProbePoint(toValue)
		if err != nil {
			return nil, fmt.Errorf("invalid --to: %w", err)
		}
		points = append(points, steppedProbeLine(from, to, step)...)
	}
	if len(points) == 0 {
		return nil, fmt.Errorf("provide at least one --at point or --from/--to line")
	}
	return expandProbeRadius(points, radius), nil
}

func parseProbePoints(values []string) ([]probePoint, error) {
	points := make([]probePoint, 0, len(values))
	for _, value := range values {
		point, err := parseProbePoint(value)
		if err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	return points, nil
}

func steppedProbeLine(from, to probePoint, step int) []probePoint {
	line := rasterProbeLine(from, to)
	points := make([]probePoint, 0, (len(line)+step-1)/step+1)
	for index := 0; index < len(line); index += step {
		points = append(points, line[index])
	}
	if points[len(points)-1] != to {
		points = append(points, to)
	}
	return points
}

func rasterProbeLine(from, to probePoint) []probePoint {
	x, y := from.X, from.Y
	dx, dy := absInt(to.X-from.X), -absInt(to.Y-from.Y)
	stepX, stepY := -1, -1
	if x < to.X {
		stepX = 1
	}
	if y < to.Y {
		stepY = 1
	}
	err := dx + dy
	points := make([]probePoint, 0, max(dx, -dy)+1)
	for {
		points = append(points, probePoint{X: x, Y: y})
		if x == to.X && y == to.Y {
			return points
		}
		twiceError := 2 * err
		if twiceError >= dy {
			err += dy
			x += stepX
		}
		if twiceError <= dx {
			err += dx
			y += stepY
		}
	}
}

func expandProbeRadius(points []probePoint, radius int) []probePoint {
	seen := make(map[probePoint]struct{})
	expanded := make([]probePoint, 0, len(points)*(radius*2+1)*(radius*2+1))
	for _, point := range points {
		for y := max(0, point.Y-radius); y <= point.Y+radius; y++ {
			for x := max(0, point.X-radius); x <= point.X+radius; x++ {
				candidate := probePoint{X: x, Y: y}
				if _, exists := seen[candidate]; exists {
					continue
				}
				seen[candidate] = struct{}{}
				expanded = append(expanded, candidate)
			}
		}
	}
	return expanded
}

func parseProbePoint(value string) (probePoint, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return probePoint{}, fmt.Errorf("--at must be x,y")
	}
	x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return probePoint{}, fmt.Errorf("--at x must be an integer")
	}
	y, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return probePoint{}, fmt.Errorf("--at y must be an integer")
	}
	if x < 0 || y < 0 {
		return probePoint{}, fmt.Errorf("--at coordinates must be non-negative")
	}
	return probePoint{X: x, Y: y}, nil
}

func scanImages(referencePath, actualPath string, axis string, index int) (scanOutput, error) {
	referenceWidth, referenceHeight, err := diff.PNGDimensions(referencePath)
	if err != nil {
		return scanOutput{}, fmt.Errorf("decode reference: %w", err)
	}
	actualWidth, actualHeight, err := diff.PNGDimensions(actualPath)
	if err != nil {
		return scanOutput{}, fmt.Errorf("decode actual: %w", err)
	}
	if referenceWidth != actualWidth || referenceHeight != actualHeight {
		return scanOutput{}, fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", referenceWidth, referenceHeight, actualWidth, actualHeight)
	}
	length := referenceWidth
	if axis == "y" {
		length = referenceHeight
	}
	if index >= map[string]int{"x": referenceHeight, "y": referenceWidth}[axis] {
		flag := map[string]string{"x": "--y", "y": "--x"}[axis]
		limit := map[string]int{"x": referenceHeight, "y": referenceWidth}[axis]
		return scanOutput{}, fmt.Errorf("%s index %d is outside image bounds %dx%d (valid 0-%d)", flag, index, referenceWidth, referenceHeight, limit-1)
	}
	referenceRuns, err := scanPNGRuns(referencePath, axis, index, length)
	if err != nil {
		return scanOutput{}, fmt.Errorf("decode reference: %w", err)
	}
	actualRuns, err := scanPNGRuns(actualPath, axis, index, length)
	if err != nil {
		return scanOutput{}, fmt.Errorf("decode actual: %w", err)
	}
	return scanOutput{Axis: axis, Index: index, Length: length, Reference: referenceRuns, Actual: actualRuns}, nil
}

func scanPNGRuns(path string, axis string, index int, length int) ([]scanRun, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	image, err := png.Decode(file)
	if err != nil {
		return nil, err
	}
	runs := make([]scanRun, 0)
	for position := 0; position < length; position++ {
		point := probePoint{X: position, Y: index}
		if axis == "y" {
			point = probePoint{X: index, Y: position}
		}
		color := colorFromImage(image, point)
		if len(runs) > 0 && runs[len(runs)-1].RGBA == color.RGBA {
			runs[len(runs)-1].End = position
			runs[len(runs)-1].Length++
			continue
		}
		runs = append(runs, scanRun{Start: position, End: position, Length: 1, RGBA: color.RGBA, Hex: color.Hex})
	}
	return runs, nil
}

func probeImages(referencePath, actualPath string, points []probePoint) (probeOutput, error) {
	referenceWidth, referenceHeight, err := diff.PNGDimensions(referencePath)
	if err != nil {
		return probeOutput{}, fmt.Errorf("decode reference: %w", err)
	}
	actualWidth, actualHeight, err := diff.PNGDimensions(actualPath)
	if err != nil {
		return probeOutput{}, fmt.Errorf("decode actual: %w", err)
	}
	if referenceWidth != actualWidth || referenceHeight != actualHeight {
		return probeOutput{}, fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", referenceWidth, referenceHeight, actualWidth, actualHeight)
	}
	output := probeOutput{Points: make([]probePointOutput, 0, len(points))}
	for _, point := range points {
		if point.X >= referenceWidth || point.Y >= referenceHeight {
			return probeOutput{}, fmt.Errorf("--at point %d,%d is outside image bounds %dx%d", point.X, point.Y, referenceWidth, referenceHeight)
		}
		referenceColor, err := probePNGColor(referencePath, point)
		if err != nil {
			return probeOutput{}, fmt.Errorf("decode reference: %w", err)
		}
		actualColor, err := probePNGColor(actualPath, point)
		if err != nil {
			return probeOutput{}, fmt.Errorf("decode actual: %w", err)
		}
		output.Points = append(output.Points, probePointOutput{
			Point:     point,
			Reference: referenceColor,
			Actual:    actualColor,
			Delta: probeDelta{
				R: int(referenceColor.RGBA[0]) - int(actualColor.RGBA[0]),
				G: int(referenceColor.RGBA[1]) - int(actualColor.RGBA[1]),
				B: int(referenceColor.RGBA[2]) - int(actualColor.RGBA[2]),
				A: int(referenceColor.RGBA[3]) - int(actualColor.RGBA[3]),
			},
		})
	}
	return output, nil
}

func probePNGColor(path string, point probePoint) (probeColor, error) {
	file, err := os.Open(path)
	if err != nil {
		return probeColor{}, err
	}
	defer func() { _ = file.Close() }()
	image, err := png.Decode(file)
	if err != nil {
		return probeColor{}, err
	}
	return colorFromImage(image, point), nil
}

func colorFromImage(image image.Image, point probePoint) probeColor {
	r, g, b, a := image.At(point.X, point.Y).RGBA()
	color := probeColor{RGBA: [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}}
	color.Hex = fmt.Sprintf("#%02X%02X%02X", color.RGBA[0], color.RGBA[1], color.RGBA[2])
	return color
}

func defaultMaskPath(actualPath string) string {
	extension := filepath.Ext(actualPath)
	if extension == "" {
		return actualPath + ".diff.png"
	}
	return strings.TrimSuffix(actualPath, extension) + ".diff.png"
}

func inputBounds(bounds diff.Bounds, inputs *diff.ImageInputs) *diff.InputBounds {
	if inputs == nil {
		return nil
	}
	return &diff.InputBounds{
		Reference: boundsWithCropOrigin(bounds, inputs.Reference.Crop),
		Actual:    boundsWithCropOrigin(bounds, inputs.Actual.Crop),
	}
}

func boundsWithCropOrigin(bounds diff.Bounds, crop *diff.Bounds) diff.Bounds {
	if crop == nil {
		return bounds
	}
	bounds.X += crop.X
	bounds.Y += crop.Y
	return bounds
}

func parseOptionalCrop(cmd *cobra.Command, flagName string) (*diff.Bounds, error) {
	value, _ := cmd.Flags().GetString(flagName)
	crop, err := parseImageRegion(value)
	if err != nil {
		return nil, fmt.Errorf("invalid --%s: %w", flagName, err)
	}
	if crop != nil && (crop.Width < 1 || crop.Height < 1) {
		return nil, fmt.Errorf("invalid --%s: width and height must be positive", flagName)
	}
	return crop, nil
}

func addInputPreparationFlags(command *cobra.Command) {
	command.Flags().String("reference-crop", "", "crop reference before operation: x,y,width,height")
	command.Flags().String("reference-metadata", "", "apply logical crop from figma export metadata JSON")
	command.Flags().String("actual-crop", "", "crop actual before operation: x,y,width,height")
}

func prepareCommandImageInputs(cmd *cobra.Command, referencePath, actualPath string) (preparedImageInputs, error) {
	referenceCrop, err := parseOptionalCrop(cmd, "reference-crop")
	if err != nil {
		return preparedImageInputs{}, err
	}
	referenceMetadataPath, _ := cmd.Flags().GetString("reference-metadata")
	referenceMetadata, err := loadExportMetadata(referenceMetadataPath)
	if err != nil {
		return preparedImageInputs{}, err
	}
	if referenceCrop != nil && referenceMetadata != nil {
		return preparedImageInputs{}, fmt.Errorf("--reference-crop and --reference-metadata cannot be used together")
	}
	actualCrop, err := parseOptionalCrop(cmd, "actual-crop")
	if err != nil {
		return preparedImageInputs{}, err
	}
	return prepareImageInputs(referencePath, actualPath, referenceCrop, actualCrop, referenceMetadata)
}

func prepareImageInputs(referencePath, actualPath string, referenceCrop, actualCrop *diff.Bounds, referenceMetadata *exportMetadata) (preparedImageInputs, error) {
	referenceWidth, referenceHeight, err := diff.PNGDimensions(referencePath)
	if err != nil {
		return preparedImageInputs{}, fmt.Errorf("decode reference: %w", err)
	}
	actualWidth, actualHeight, err := diff.PNGDimensions(actualPath)
	if err != nil {
		return preparedImageInputs{}, fmt.Errorf("decode actual: %w", err)
	}
	if referenceMetadata != nil {
		if int(referenceMetadata.ExportBounds.Width) != referenceWidth || int(referenceMetadata.ExportBounds.Height) != referenceHeight {
			return preparedImageInputs{}, fmt.Errorf("--reference-metadata export bounds %gx%g do not match reference image %dx%d", referenceMetadata.ExportBounds.Width, referenceMetadata.ExportBounds.Height, referenceWidth, referenceHeight)
		}
		referenceCrop = cropFromExportMetadata(*referenceMetadata)
	}
	if referenceCrop == nil && actualCrop == nil {
		return preparedImageInputs{referencePath: referencePath, actualPath: actualPath, cleanup: func() {}}, nil
	}
	if err := validateCrop(referenceCrop, referenceWidth, referenceHeight); err != nil {
		return preparedImageInputs{}, fmt.Errorf("invalid --reference-crop: %w", err)
	}
	if err := validateCrop(actualCrop, actualWidth, actualHeight); err != nil {
		return preparedImageInputs{}, fmt.Errorf("invalid --actual-crop: %w", err)
	}
	metadata := &diff.ImageInputs{
		Reference: diff.ImageInput{Width: referenceWidth, Height: referenceHeight, Crop: referenceCrop},
		Actual:    diff.ImageInput{Width: actualWidth, Height: actualHeight, Crop: actualCrop},
	}
	referenceCompareWidth, referenceCompareHeight := croppedDimensions(referenceWidth, referenceHeight, referenceCrop)
	actualCompareWidth, actualCompareHeight := croppedDimensions(actualWidth, actualHeight, actualCrop)
	if referenceCompareWidth != actualCompareWidth || referenceCompareHeight != actualCompareHeight {
		return preparedImageInputs{}, fmt.Errorf("cropped image dimensions differ: reference is %dx%d, actual is %dx%d", referenceCompareWidth, referenceCompareHeight, actualCompareWidth, actualCompareHeight)
	}
	tempDir, err := os.MkdirTemp("", "pixel-perfect-crops-*")
	if err != nil {
		return preparedImageInputs{}, err
	}
	cleanup := func() { _ = os.RemoveAll(tempDir) }
	prepared := preparedImageInputs{referencePath: referencePath, actualPath: actualPath, metadata: metadata, cleanup: cleanup}
	if referenceCrop != nil {
		prepared.referencePath = filepath.Join(tempDir, "reference.png")
		if err := diff.WriteCroppedPNG(referencePath, prepared.referencePath, *referenceCrop); err != nil {
			cleanup()
			return preparedImageInputs{}, fmt.Errorf("invalid --reference-crop: %w", err)
		}
	}
	if actualCrop != nil {
		prepared.actualPath = filepath.Join(tempDir, "actual.png")
		if err := diff.WriteCroppedPNG(actualPath, prepared.actualPath, *actualCrop); err != nil {
			cleanup()
			return preparedImageInputs{}, fmt.Errorf("invalid --actual-crop: %w", err)
		}
	}
	return prepared, nil
}

func loadExportMetadata(path string) (*exportMetadata, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var metadata exportMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	if metadata.Version != 1 {
		return nil, fmt.Errorf("unsupported --reference-metadata version %d", metadata.Version)
	}
	return &metadata, nil
}

func cropFromExportMetadata(metadata exportMetadata) *diff.Bounds {
	if metadata.LogicalCrop != nil {
		crop := &diff.Bounds{
			X:      int(math.Round(metadata.LogicalCrop.X)),
			Y:      int(math.Round(metadata.LogicalCrop.Y)),
			Width:  int(math.Round(metadata.LogicalCrop.Width)),
			Height: int(math.Round(metadata.LogicalCrop.Height)),
		}
		if crop.X >= 0 && crop.Y >= 0 && crop.X+crop.Width <= int(math.Round(metadata.ExportBounds.Width)) && crop.Y+crop.Height <= int(math.Round(metadata.ExportBounds.Height)) {
			return crop
		}
	}
	return &diff.Bounds{
		X:      int(math.Floor((metadata.ExportBounds.Width - metadata.NodeBounds.Width) / 2)),
		Y:      int(math.Floor((metadata.ExportBounds.Height - metadata.NodeBounds.Height) / 2)),
		Width:  int(math.Round(metadata.NodeBounds.Width)),
		Height: int(math.Round(metadata.NodeBounds.Height)),
	}
}

func croppedDimensions(width, height int, crop *diff.Bounds) (int, int) {
	if crop == nil {
		return width, height
	}
	return crop.Width, crop.Height
}

func validateCrop(crop *diff.Bounds, width, height int) error {
	if crop == nil {
		return nil
	}
	if crop.X < 0 || crop.Y < 0 || crop.Width <= 0 || crop.Height <= 0 || crop.X+crop.Width > width || crop.Y+crop.Height > height {
		return fmt.Errorf("crop %d,%d,%d,%d is outside image bounds %dx%d", crop.X, crop.Y, crop.Width, crop.Height, width, height)
	}
	return nil
}

func groupImageRegions(regions []diff.Region, gap int) []diff.Region {
	grouped := append([]diff.Region(nil), regions...)
	for merged := true; merged; {
		merged = false
		for i := 0; i < len(grouped) && !merged; i++ {
			for j := i + 1; j < len(grouped); j++ {
				if !regionsWithinGap(grouped[i].Bounds, grouped[j].Bounds, gap) {
					continue
				}
				grouped[i] = mergeImageRegions(grouped[i], grouped[j])
				grouped = append(grouped[:j], grouped[j+1:]...)
				merged = true
				break
			}
		}
	}
	sort.SliceStable(grouped, func(i, j int) bool { return grouped[i].ChangedPixels > grouped[j].ChangedPixels })
	return grouped
}

func regionsWithinGap(first, second diff.Bounds, gap int) bool {
	return first.X <= second.X+second.Width+gap && second.X <= first.X+first.Width+gap &&
		first.Y <= second.Y+second.Height+gap && second.Y <= first.Y+first.Height+gap
}

func mergeImageRegions(first, second diff.Region) diff.Region {
	left, top := min(first.Bounds.X, second.Bounds.X), min(first.Bounds.Y, second.Bounds.Y)
	right := max(first.Bounds.X+first.Bounds.Width, second.Bounds.X+second.Bounds.Width)
	bottom := max(first.Bounds.Y+first.Bounds.Height, second.Bounds.Y+second.Bounds.Height)
	return diff.Region{Bounds: diff.Bounds{X: left, Y: top, Width: right - left, Height: bottom - top}, ChangedPixels: first.ChangedPixels + second.ChangedPixels}
}

func filterImageRegions(regions []diff.Region, minimumPixels int) []diff.Region {
	filtered := make([]diff.Region, 0, len(regions))
	for _, region := range regions {
		if region.ChangedPixels >= minimumPixels {
			filtered = append(filtered, region)
		}
	}
	return filtered
}

func parseImageRegion(value string) (*diff.Bounds, error) {
	if value == "" {
		return nil, nil
	}
	parts := strings.Split(value, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("--region must be x,y,width,height")
	}
	values := make([]int, 4)
	for index, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("--region must contain integers: %w", err)
		}
		values[index] = value
	}
	return &diff.Bounds{X: values[0], Y: values[1], Width: values[2], Height: values[3]}, nil
}

func samePath(first, second string) bool {
	firstAbsolute, firstErr := filepath.Abs(first)
	secondAbsolute, secondErr := filepath.Abs(second)
	if firstErr != nil || secondErr != nil {
		return filepath.Clean(first) == filepath.Clean(second)
	}
	return filepath.Clean(firstAbsolute) == filepath.Clean(secondAbsolute)
}

type outputEnvelope struct {
	diff.ImageComparison
	VisualContext *imagecontext.Result     `json:"visualContext,omitempty"`
	Configuration *comparisonConfiguration `json:"configuration,omitempty"`
}

func writeProbeCSV(command *cobra.Command, output probeOutput) error {
	writer := command.OutOrStdout()
	if _, err := fmt.Fprintln(writer, "x,y,ref,act,delta,input_ref,input_act"); err != nil {
		return err
	}
	for _, point := range output.Points {
		inputReference, inputActual := "", ""
		if point.InputPoint != nil {
			inputReference = formatProbePoint(point.InputPoint.Reference)
			inputActual = formatProbePoint(point.InputPoint.Actual)
		}
		if _, err := fmt.Fprintf(writer, "%d,%d,%s,%s,%d,%s,%s\n", point.Point.X, point.Point.Y, point.Reference.Hex, point.Actual.Hex, maxAbsDelta(point.Delta), inputReference, inputActual); err != nil {
			return err
		}
	}
	return nil
}

func writeScanCSV(command *cobra.Command, output scanOutput) error {
	writer := command.OutOrStdout()
	if _, err := fmt.Fprintln(writer, "image,axis,index,start,end,length,hex,input_axis,input_index"); err != nil {
		return err
	}
	if err := writeScanRunsCSV(writer, "ref", output.Axis, output.Index, output.Reference, output.InputLine, true); err != nil {
		return err
	}
	return writeScanRunsCSV(writer, "act", output.Axis, output.Index, output.Actual, output.InputLine, false)
}

func writeScanRunsCSV(writer io.Writer, imageName string, axis string, index int, runs []scanRun, inputLine *scanInputLine, reference bool) error {
	inputAxis, inputIndex := "", ""
	if inputLine != nil {
		line := inputLine.Actual
		if reference {
			line = inputLine.Reference
		}
		inputAxis = line.Axis
		inputIndex = strconv.Itoa(line.Index)
	}
	for _, run := range runs {
		if _, err := fmt.Fprintf(writer, "%s,%s,%d,%d,%d,%d,%s,%s,%s\n", imageName, axis, index, run.Start, run.End, run.Length, run.Hex, inputAxis, inputIndex); err != nil {
			return err
		}
	}
	return nil
}

func formatProbePoint(point probePoint) string {
	return fmt.Sprintf("%d:%d", point.X, point.Y)
}

func maxAbsDelta(delta probeDelta) int {
	return max(max(absInt(delta.R), absInt(delta.G)), max(absInt(delta.B), absInt(delta.A)))
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func writeJSON(command *cobra.Command, value any) error {
	encoder := json.NewEncoder(command.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

// NewCommand creates the standalone image comparison command.
func NewCommand() *cobra.Command {
	command := newCommand(nil)
	command.Use = "pixel-perfect <reference.png> <actual.png>"
	command.Example = `  pixel-perfect reference.png actual.png
  pixel-perfect reference.png actual.png --overlay overlay.png
  pixel-perfect reference.png actual.png --max-changed-ratio 0.01`
	command.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return cli.NewUsageError(fmt.Errorf("%w\n\n%s", err, availableFlagHelp(cmd)))
	})
	cli.MarkUsageErrors(command)
	return command
}

func requireImagePair(action, example string) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if len(args) == 2 {
			return nil
		}
		return fmt.Errorf("%s requires <reference.png> and <actual.png>; received %d argument(s)\n\nExample: %s", action, len(args), example)
	}
}

func availableFlagHelp(command *cobra.Command) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Available flags for %q:\n", command.CommandPath())
	flags := strings.TrimRight(command.LocalFlags().FlagUsages(), "\n")
	if flags == "" {
		builder.WriteString("  (none)")
	} else {
		builder.WriteString(flags)
	}
	fmt.Fprintf(&builder, "\n\nRun `%s --help` for details.", command.CommandPath())
	return builder.String()
}
