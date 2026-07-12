package pixelperfectcmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/imagecontext"
	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
	"github.com/spf13/cobra"
)

type imageComparer func(referencePath, actualPath, maskPath string, threshold uint8, perceptualThreshold float64, region *diff.Bounds, ignored []diff.Bounds) (diff.ImageComparison, error)

type preparedImageInputs struct {
	referencePath string
	actualPath    string
	metadata      *diff.ImageInputs
	cleanup       func()
}

func newCommand(compare imageComparer) *cobra.Command {
	command := &cobra.Command{
		Use:   "image <reference.png> <actual.png>",
		Short: "Compare equal-sized PNGs and write a changed-pixel mask",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			threshold, _ := cmd.Flags().GetUint8("threshold")
			if output == "" {
				return fmt.Errorf("--output is required")
			}
			overlay, _ := cmd.Flags().GetString("overlay")
			if samePath(output, args[0]) || samePath(output, args[1]) {
				return fmt.Errorf("--output must not overwrite an input image")
			}
			if overlay != "" && (samePath(overlay, args[0]) || samePath(overlay, args[1])) {
				return fmt.Errorf("--overlay must not overwrite an input image")
			}
			if overlay != "" && samePath(overlay, output) {
				return fmt.Errorf("--overlay must differ from --output")
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
			actualCrop, err := parseOptionalCrop(cmd, "actual-crop")
			if err != nil {
				return err
			}
			inputs, err := prepareImageInputs(args[0], args[1], referenceCrop, actualCrop)
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
			outputResult := outputEnvelope{ImageComparison: result}
			visualContextEnabled, _ := cmd.Flags().GetBool("visual-context")
			if visualContextEnabled {
				provider, _ := cmd.Flags().GetString("visual-context-provider")
				model, _ := cmd.Flags().GetString("visual-context-model")
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
				input := imagecontext.Input{ReferencePath: inputs.referencePath, ActualPath: inputs.actualPath, Regions: regions}
				visualContext, explainErr := client.Describe(context.Background(), input)
				if explainErr != nil {
					return explainErr
				}
				outputResult.VisualContext = &visualContext
			}
			return writeJSON(cmd, outputResult)
		},
	}
	command.Flags().StringP("output", "o", "", "path for transparent PNG difference mask")
	command.Flags().Uint8("threshold", 0, "ignore per-channel differences at or below this value (0-255)")
	command.Flags().Float64("perceptual-threshold", diff.DefaultPerceptualThreshold, "OKLab HyAB distance above which a pixel is perceptually changed (non-negative)")
	command.Flags().String("region", "", "compare only x,y,width,height")
	command.Flags().String("reference-crop", "", "crop reference before comparing: x,y,width,height")
	command.Flags().String("actual-crop", "", "crop actual before comparing: x,y,width,height")
	command.Flags().StringArray("ignore-region", nil, "exclude x,y,width,height; repeat for multiple areas")
	command.Flags().String("mask", "", "full-size PNG selecting compared pixels (visible non-black includes)")
	command.Flags().String("overlay", "", "path for directional overlay (reference red, actual green)")
	command.Flags().Int("suggest-offset", 0, "report best translation within this pixel radius without applying it")
	command.Flags().Int("region-gap", 0, "group mismatch regions separated by at most this many pixels")
	command.Flags().Int("min-region-pixels", 1, "omit disconnected regions smaller than this many changed pixels")
	command.Flags().Float64("max-rmse", -1, "fail when normalized RMSE exceeds this value")
	command.Flags().Float64("max-changed-ratio", -1, "fail when changed-pixel ratio exceeds this value")
	command.Flags().Float64("max-perceptual-changed-ratio", -1, "fail when perceptual changed-pixel ratio exceeds this value")
	command.Flags().Bool("visual-context", false, "add advisory visual descriptions using the configured multimodal model")
	command.Flags().String("visual-context-provider", "openrouter", "visual context provider: openrouter or openai")
	command.Flags().String("visual-context-model", "", "override the visual context model")
	return command
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

func prepareImageInputs(referencePath, actualPath string, referenceCrop, actualCrop *diff.Bounds) (preparedImageInputs, error) {
	referenceWidth, referenceHeight, err := diff.PNGDimensions(referencePath)
	if err != nil {
		return preparedImageInputs{}, fmt.Errorf("decode reference: %w", err)
	}
	actualWidth, actualHeight, err := diff.PNGDimensions(actualPath)
	if err != nil {
		return preparedImageInputs{}, fmt.Errorf("decode actual: %w", err)
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
