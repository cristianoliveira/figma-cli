package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/spf13/cobra"
)

type imageComparer func(referencePath, actualPath, maskPath string, threshold uint8) (diff.ImageComparison, error)

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
			result, err := compare(args[0], args[1], output, threshold)
			if err != nil {
				return err
			}
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
	command.Flags().Float64("max-rmse", -1, "fail when normalized RMSE exceeds this value")
	command.Flags().Float64("max-changed-ratio", -1, "fail when changed-pixel ratio exceeds this value")
	return command
}

func init() {
	diffCmd.AddCommand(newDiffImageCommand(diff.CompareImages))
}
