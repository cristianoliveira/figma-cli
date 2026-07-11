package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/spf13/cobra"
)

type imageComparer func(referencePath, actualPath, maskPath string, threshold uint8, region *diff.Bounds) (diff.ImageComparison, error)

func newDiffImageCommand(compare imageComparer) *cobra.Command {
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
			region, err := parseImageRegion(cmd.Flags().Lookup("region").Value.String())
			if err != nil {
				return err
			}
			result, err := compare(args[0], args[1], output, threshold, region)
			if err != nil {
				return err
			}
			overlay, _ := cmd.Flags().GetString("overlay")
			if overlay != "" {
				if err := diff.WriteImageOverlay(args[0], args[1], overlay, region); err != nil {
					return err
				}
				result.Overlay = overlay
			}
			minRegionPixels, _ := cmd.Flags().GetInt("min-region-pixels")
			result.Regions = filterImageRegions(result.Regions, minRegionPixels)
			maxRMSE, _ := cmd.Flags().GetFloat64("max-rmse")
			if maxRMSE >= 0 && result.RMSE > maxRMSE {
				return fmt.Errorf("image diff validation failed: RMSE %.6f exceeds maximum %.6f", result.RMSE, maxRMSE)
			}
			maxChangedRatio, _ := cmd.Flags().GetFloat64("max-changed-ratio")
			if maxChangedRatio >= 0 && result.ChangedRatio > maxChangedRatio {
				return fmt.Errorf("image diff validation failed: changed ratio %.6f exceeds maximum %.6f", result.ChangedRatio, maxChangedRatio)
			}
			return cli.NewPrinter(cmd).JSON(result)
		},
	}
	command.Flags().StringP("output", "o", "", "path for transparent PNG difference mask")
	command.Flags().Uint8("threshold", 0, "ignore per-channel differences at or below this value (0-255)")
	command.Flags().String("region", "", "compare only x,y,width,height")
	command.Flags().String("overlay", "", "path for directional overlay (reference red, actual green)")
	command.Flags().Int("min-region-pixels", 1, "omit disconnected regions smaller than this many changed pixels")
	command.Flags().Float64("max-rmse", -1, "fail when normalized RMSE exceeds this value")
	command.Flags().Float64("max-changed-ratio", -1, "fail when changed-pixel ratio exceeds this value")
	return command
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

func init() {
	diffCmd.AddCommand(newDiffImageCommand(diff.CompareImagesInRegion))
}
