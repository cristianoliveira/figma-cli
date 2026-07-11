package pixelperfectcmd

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	diff "github.com/cristianoliveira/figma-cli/internal/imagediff"
	"github.com/spf13/cobra"
)

type imageComparer func(referencePath, actualPath, maskPath string, threshold uint8, perceptualThreshold float64, region *diff.Bounds, ignored []diff.Bounds) (diff.ImageComparison, error)

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
			comparisonMask, _ := cmd.Flags().GetString("mask")
			if comparisonMask != "" {
				maskedRegions, maskErr := diff.IgnoredRegionsFromMask(comparisonMask, args[0])
				if maskErr != nil {
					return maskErr
				}
				ignored = append(ignored, maskedRegions...)
			}
			result, err := compare(args[0], args[1], output, threshold, perceptualThreshold, region, ignored)
			if err != nil {
				return err
			}
			if overlay != "" {
				if err := diff.WriteImageOverlay(args[0], args[1], overlay, region, ignored); err != nil {
					return err
				}
				result.Overlay = overlay
			}
			if offsetRadius > 0 {
				suggestedOffset, offsetErr := diff.SuggestImageOffset(args[0], args[1], offsetRadius, region, ignored)
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
				metrics, metricsErr := diff.MeasureImageRegion(args[0], args[1], result.Regions[index].Bounds, threshold, ignored)
				if metricsErr != nil {
					return metricsErr
				}
				result.Regions[index].ChangedPixels = metrics.ChangedPixels
				result.Regions[index].ChangedRatio = metrics.ChangedRatio
				result.Regions[index].RMSE = metrics.RMSE
				result.Regions[index].EdgeRMSE = metrics.EdgeRMSE
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
			return writeJSON(cmd, result)
		},
	}
	command.Flags().StringP("output", "o", "", "path for transparent PNG difference mask")
	command.Flags().Uint8("threshold", 0, "ignore per-channel differences at or below this value (0-255)")
	command.Flags().Float64("perceptual-threshold", diff.DefaultPerceptualThreshold, "OKLab HyAB distance above which a pixel is perceptually changed (non-negative)")
	command.Flags().String("region", "", "compare only x,y,width,height")
	command.Flags().StringArray("ignore-region", nil, "exclude x,y,width,height; repeat for multiple areas")
	command.Flags().String("mask", "", "full-size PNG selecting compared pixels (visible non-black includes)")
	command.Flags().String("overlay", "", "path for directional overlay (reference red, actual green)")
	command.Flags().Int("suggest-offset", 0, "report best translation within this pixel radius without applying it")
	command.Flags().Int("region-gap", 0, "group mismatch regions separated by at most this many pixels")
	command.Flags().Int("min-region-pixels", 1, "omit disconnected regions smaller than this many changed pixels")
	command.Flags().Float64("max-rmse", -1, "fail when normalized RMSE exceeds this value")
	command.Flags().Float64("max-changed-ratio", -1, "fail when changed-pixel ratio exceeds this value")
	command.Flags().Float64("max-perceptual-changed-ratio", -1, "fail when perceptual changed-pixel ratio exceeds this value")
	return command
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
