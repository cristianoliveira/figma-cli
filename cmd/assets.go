package cmd

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

const (
	assetFormatAuto   = "auto"
	assetKindAll      = "all"
	assetKindIcon     = "icon"
	assetKindImage    = "image"
	assetKindInstance = "instance"
	assetKindVector   = "vector"
)

var assetFilenameCharacters = regexp.MustCompile(`[^a-z0-9]+`)

var assetsCmd = newAssetsCommand(cli.LoadClient, nil)

func newAssetsCommand(loadClient func() (*figma.Client, error), downloadClient *http.Client) *cobra.Command {
	command := &cobra.Command{
		Use:   "assets [figma-url-or-file-id]",
		Short: "Download image, instance, and vector assets from a Figma node tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return err
			}
			outputDirectory, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")
			kind, _ := cmd.Flags().GetString("kind")
			nameFilter, _ := cmd.Flags().GetString("name")
			filenameMode, _ := cmd.Flags().GetString("filename")
			trimNamePrefix, _ := cmd.Flags().GetString("trim-name-prefix")
			allowPartial, _ := cmd.Flags().GetBool("allow-partial")
			if format != assetFormatAuto {
				if err := figma.ValidateExportFormat(format); err != nil {
					return err
				}
			}
			if kind != assetKindAll && kind != assetKindIcon && kind != assetKindImage && kind != assetKindInstance && kind != assetKindVector {
				return fmt.Errorf("invalid kind %q: expected all, icon, image, instance, or vector", kind)
			}
			if filenameMode != "name" && filenameMode != "name-id" {
				return fmt.Errorf("invalid filename mode %q: expected name or name-id", filenameMode)
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			nodeIDs, err := figma.ResolveRequiredNodeIDs(input, nodeID, "assets")
			if err != nil {
				return err
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			filename := assetFilename
			if filenameMode == "name" {
				filename = func(asset extract.Asset) string { return assetNameFilename(asset, trimNamePrefix) }
			}
			manifest, err := assets.ExportAssets(assets.AssetExportRequest{
				Client:          client,
				FileID:          input.FileID,
				NodeIDs:         nodeIDs,
				OutputDirectory: outputDirectory,
				Kind:            kind,
				Format:          format,
				NameFilter:      nameFilter,
				DownloadClient:  downloadClient,
				Filename:        filename,
			})
			if err != nil {
				return err
			}

			asJSON, _ := cmd.Flags().GetBool("json")
			if asJSON {
				if err := cli.NewPrinter(cmd).JSON(manifest); err != nil {
					return err
				}
			} else {
				for _, item := range manifest.Items {
					if item.Error != "" {
						cmd.PrintErrf("warning: node %s: %s\n", item.NodeID, item.Error)
						continue
					}
					cmd.Println(item.Path)
				}
				cmd.PrintErrf("exported %d asset(s), %d failed\n", manifest.Succeeded, manifest.Failed)
			}
			return assetExportResult(manifest, allowPartial)
		},
	}
	addNodeIDFlag(command, "node ID to inspect; defaults to URL node-id")
	command.Flags().StringP("output", "o", "assets", "output directory")
	command.Flags().String("format", assetFormatAuto, "export format: auto, png, jpg, svg, or pdf")
	command.Flags().String("kind", assetKindAll, "asset kind: all, icon, image, instance, or vector")
	command.Flags().String("name", "", "filter layers by case-insensitive name substring")
	command.Flags().String("filename", "name-id", "filename mode: name or name-id")
	command.Flags().String("trim-name-prefix", "", "prefix to remove in name filename mode")
	command.Flags().Bool("allow-partial", false, "exit successfully when only some assets export")
	return command
}

func assetExportResult(manifest assets.AssetExportManifest, allowPartial bool) error {
	if manifest.Failed == 0 || allowPartial {
		return nil
	}
	return &cli.ExitCodeError{Code: 1}
}

func assetFilename(asset extract.Asset) string {
	id := strings.NewReplacer(":", "-", ";", "-").Replace(asset.ID)
	return normalizedAssetName(asset.Name) + "_" + id
}

func assetNameFilename(asset extract.Asset, trimPrefix string) string {
	name := asset.Name
	if len(name) >= len(trimPrefix) && strings.EqualFold(name[:len(trimPrefix)], trimPrefix) {
		name = name[len(trimPrefix):]
	}
	return normalizedAssetName(name)
}

func normalizedAssetName(name string) string {
	normalized := strings.Trim(assetFilenameCharacters.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if normalized == "" {
		return "asset"
	}
	return normalized
}

func init() {
	rootCmd.AddCommand(assetsCmd)
}
