package pixelperfectcmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/imagecontext"
	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
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
	Version      int                `json:"version"`
	NodeBounds   exportMetadataSize `json:"nodeBounds"`
	ExportBounds exportMetadataSize `json:"exportBounds"`
	LogicalCrop  *diff.Bounds       `json:"logicalCrop"`
}

type exportMetadataSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type probeOutput struct {
	Point     probePoint `json:"point"`
	Reference probeColor `json:"reference"`
	Actual    probeColor `json:"actual"`
	Delta     probeDelta `json:"delta"`
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
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			result, err := compare(inputs.referencePath, inputs.actualPath, output, threshold, perceptualThreshold, region, ignored)
			if err != nil {
				return err
			}
			result.Inputs = inputs.metadata
			if overlay != "" {
				if err := diff.WriteImageOverlay(inputs.referencePath, inputs.actualPath, overlay, region, ignored); err != nil {
					return err
				}
				result.Overlay = overlay
			}
			if offsetRadius > 0 {
				suggestedOffset, offsetErr := diff.SuggestImageOffset(inputs.referencePath, inputs.actualPath, offsetRadius, region, ignored)
				if offsetErr != nil {
					return offsetErr
				}
				if !math.IsInf(suggestedOffset.RMSE, 0) && !math.IsNaN(suggestedOffset.RMSE) {
					result.SuggestedOffset = &suggestedOffset
				}
			}
			result.Regions = groupImageRegions(result.Regions, regionGap)
			result.Regions = filterImageRegions(result.Regions, minRegionPixels)
			if len(result.Regions) > 20 {
				result.Regions = result.Regions[:20]
			}
			for index := range result.Regions {
				result.Regions[index].InputBounds = inputBounds(result.Regions[index].Bounds, inputs.metadata)
				metrics, metricsErr := diff.MeasureImageRegionWithThresholds(inputs.referencePath, inputs.actualPath, result.Regions[index].Bounds, threshold, perceptualThreshold, ignored)
				if metricsErr != nil {
					return metricsErr
				}
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
			outputResult := outputEnvelope{ImageComparison: result}
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
	command.Flags().Int("suggest-offset", 0, "report best translation within this pixel radius without applying it")
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
	return command
}

func newProbeCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "probe <reference.png> <actual.png>",
		Short: "Inspect colors at one pixel in two equal-sized PNGs",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			point, err := parseProbePoint(cmd.Flags().Lookup("at").Value.String())
			if err != nil {
				return err
			}
			output, err := probeImages(args[0], args[1], point)
			if err != nil {
				return err
			}
			return writeJSON(cmd, output)
		},
	}
	command.Flags().String("at", "", "pixel coordinate to inspect: x,y")
	_ = command.MarkFlagRequired("at")
	return command
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

func probeImages(referencePath, actualPath string, point probePoint) (probeOutput, error) {
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
	return probeOutput{
		Point:     point,
		Reference: referenceColor,
		Actual:    actualColor,
		Delta: probeDelta{
			R: int(referenceColor.RGBA[0]) - int(actualColor.RGBA[0]),
			G: int(referenceColor.RGBA[1]) - int(actualColor.RGBA[1]),
			B: int(referenceColor.RGBA[2]) - int(actualColor.RGBA[2]),
			A: int(referenceColor.RGBA[3]) - int(actualColor.RGBA[3]),
		},
	}, nil
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
	r, g, b, a := image.At(point.X, point.Y).RGBA()
	color := probeColor{RGBA: [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}}
	color.Hex = fmt.Sprintf("#%02X%02X%02X", color.RGBA[0], color.RGBA[1], color.RGBA[2])
	return color, nil
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
		return metadata.LogicalCrop
	}
	return &diff.Bounds{
		X:      int((metadata.ExportBounds.Width - metadata.NodeBounds.Width) / 2),
		Y:      int((metadata.ExportBounds.Height - metadata.NodeBounds.Height) / 2),
		Width:  int(metadata.NodeBounds.Width),
		Height: int(metadata.NodeBounds.Height),
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
	VisualContext *imagecontext.Result `json:"visualContext,omitempty"`
}

func writeJSON(command *cobra.Command, value any) error {
	encoder := json.NewEncoder(command.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

// NewCommand creates the standalone image comparison command.
func NewCommand() *cobra.Command {
	command := newCommand(diff.CompareImagesWithThresholds)
	command.Use = "pixel-perfect <reference.png> <actual.png>"
	return command
}
