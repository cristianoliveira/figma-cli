package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

type downloadedAsset struct {
	extract.Asset
	Path  string `json:"path,omitempty"`
	Error string `json:"error,omitempty"`
}

const assetFormatAuto = "auto"

var assetFilenameCharacters = regexp.MustCompile(`[^a-z0-9]+`)

var assetsCmd = &cobra.Command{
	Use:   "assets [figma-url-or-file-id]",
	Short: "Download image, instance, and vector assets from a Figma node tree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeID, _ := cmd.Flags().GetString("id")
		outputDirectory, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")
		kind, _ := cmd.Flags().GetString("kind")
		if format != assetFormatAuto && format != "png" && format != "svg" {
			return fmt.Errorf("invalid format %q: expected auto, png, or svg", format)
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
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
		if err != nil {
			return err
		}
		assets := extract.ExtractAssets(documents)
		if err := os.MkdirAll(outputDirectory, 0o755); err != nil {
			return err
		}

		downloaded := make([]downloadedAsset, 0, len(assets))
		for _, asset := range assets {
			if kind != "all" && asset.Kind != kind {
				continue
			}
			assetFormat := asset.Format
			if format != assetFormatAuto {
				assetFormat = format
			}
			apiURL, err := figma.BuildExportURL(input.FileID, []string{asset.ID}, assetFormat)
			if err != nil {
				return err
			}
			asset.Format = assetFormat
			assetURL, err := figma.FetchExportURL(client, apiURL, asset.ID)
			if err != nil {
				downloaded = append(downloaded, downloadedAsset{Asset: asset, Error: err.Error()})
				continue
			}
			path := filepath.Join(outputDirectory, assetFilename(asset)+"."+assetFormat)
			if err := cli.DownloadFile(http.DefaultClient, path, assetURL); err != nil {
				downloaded = append(downloaded, downloadedAsset{Asset: asset, Error: err.Error()})
				continue
			}
			downloaded = append(downloaded, downloadedAsset{Asset: asset, Path: path})
		}

		successCount := 0
		failedCount := 0
		for _, asset := range downloaded {
			if asset.Error == "" {
				successCount++
			} else {
				failedCount++
			}
		}
		asJSON, _ := cmd.Flags().GetBool("json")
		if asJSON {
			return cli.NewPrinter(cmd).JSON(map[string]any{
				"assets": downloaded,
				"count":  successCount,
				"failed": failedCount,
			})
		}
		paths := make([]string, 0, successCount)
		for _, asset := range downloaded {
			if asset.Error != "" {
				cmd.PrintErrf("warning: node %s: %s\n", asset.ID, asset.Error)
				continue
			}
			paths = append(paths, asset.Path)
		}
		output := strings.Join(paths, "\n")
		if output != "" {
			output += "\n"
		}
		return cli.NewPrinter(cmd).Text("assets", output)
	},
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
	assetsCmd.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	assetsCmd.Flags().StringP("output", "o", "assets", "output directory")
	assetsCmd.Flags().String("format", assetFormatAuto, "export format: auto, png, or svg")
	assetsCmd.Flags().String("kind", "all", "asset kind: all, image, instance, or vector")
	rootCmd.AddCommand(assetsCmd)
}
