package cmd

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

const assetFormatAuto = "auto"

var assetFilenameCharacters = regexp.MustCompile(`[^a-z0-9]+`)

var assetsCmd = newAssetsCommand(cli.LoadClient, http.DefaultClient)

func newAssetsCommand(loadClient func() (*figma.Client, error), downloadClient *http.Client) *cobra.Command {
	command := &cobra.Command{
		Use:   "assets [figma-url-or-file-id]",
		Short: "Download image, instance, and vector assets from a Figma node tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, _ := cmd.Flags().GetString("id")
			outputDirectory, _ := cmd.Flags().GetString("output")
			format, _ := cmd.Flags().GetString("format")
			kind, _ := cmd.Flags().GetString("kind")
			allowPartial, _ := cmd.Flags().GetBool("allow-partial")
			if format != assetFormatAuto {
				if err := figma.ValidateExportFormat(format); err != nil {
					return err
				}
			}
			if kind != "all" && kind != "image" && kind != "instance" && kind != "vector" {
				return fmt.Errorf("invalid kind %q: expected all, image, instance, or vector", kind)
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
			documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
			if err != nil {
				return err
			}
			assets := filterAssets(extract.ExtractAssets(documents), kind, format)
			if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
				return err
			}

			exporter := cli.AssetExporter{
				HTTPClient: downloadClient,
				Filename:   assetFilename,
				FetchURL: func(nodeID, assetFormat string) (string, error) {
					apiURL, err := figma.BuildExportURL(input.FileID, []string{nodeID}, assetFormat)
					if err != nil {
						return "", err
					}
					return figma.FetchExportURL(client, apiURL, nodeID)
				},
			}
			manifest := exporter.Export(outputDirectory, assets)

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
	command.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	command.Flags().StringP("output", "o", "assets", "output directory")
	command.Flags().String("format", assetFormatAuto, "export format: auto, png, jpg, svg, or pdf")
	command.Flags().String("kind", "all", "asset kind: all, image, instance, or vector")
	command.Flags().Bool("allow-partial", false, "exit successfully when only some assets export")
	return command
}

func filterAssets(assets []extract.Asset, kind, format string) []extract.Asset {
	filtered := make([]extract.Asset, 0, len(assets))
	for _, asset := range assets {
		if kind != "all" && asset.Kind != kind {
			continue
		}
		if format != assetFormatAuto {
			asset.Format = format
		}
		filtered = append(filtered, asset)
	}
	return filtered
}

func assetExportResult(manifest cli.AssetExportManifest, allowPartial bool) error {
	if manifest.Failed == 0 || allowPartial {
		return nil
	}
	return &cli.ExitCodeError{Code: 1}
}

func assetFilename(asset extract.Asset) string {
	name := strings.Trim(assetFilenameCharacters.ReplaceAllString(strings.ToLower(asset.Name), "-"), "-")
	if name == "" {
		name = "asset"
	}
	id := strings.NewReplacer(":", "-", ";", "-").Replace(asset.ID)
	return name + "_" + id
}

func init() {
	rootCmd.AddCommand(assetsCmd)
}
