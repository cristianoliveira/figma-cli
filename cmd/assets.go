package cmd

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/assets"
	"github.com/cristianoliveira/figma-cli/internal/assetsedge"
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

// AssetApplication is the consumer-owned port the `figma assets`
// command drives. Production composition wires it through
// assets.NewApplication in cmd/deps.go; tests inject fakes. The
// command builds the edge adapters (Figma source, Figma URL source,
// HTTP sink) and hands them, plus the request, to the application.
type AssetApplication interface {
	Run(ctx context.Context, req assets.ApplicationRequest) (assets.AssetExportManifest, error)
}

// newAssetsCommand constructs `figma assets`. The command is now
// responsible only for flag/argument parsing, building the application
// request, and rendering the manifest. Filtering, naming, collision
// policy, and orchestration live in the asset application service.
func newAssetsCommand(deps Deps) *cobra.Command {
	return newAssetsCommandWithClient(deps, nil)
}

// newAssetsCommandWithClient exposes the optional downloadClient injection
// for tests; production callers should use newAssetsCommand.
func newAssetsCommandWithClient(deps Deps, downloadClient *http.Client) *cobra.Command {
	command := &cobra.Command{
		Use:   "assets [figma-url-or-file-id]",
		Short: "Download image, instance, and vector assets from a Figma node tree",
		Example: `  figma assets "<url>?node-id=42-1"
  figma assets --id 42:1 --kind icon --format svg <file-key>
  figma assets --id 42:1 --output public/assets --json <file-key>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
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
					return cli.NewUsageError(err)
				}
			}
			if kind != assetKindAll && kind != assetKindIcon && kind != assetKindImage && kind != assetKindInstance && kind != assetKindVector {
				return cli.NewUsageError(fmt.Errorf("invalid kind %q: expected all, icon, image, instance, or vector", kind))
			}
			if filenameMode != "name" && filenameMode != "name-id" {
				return cli.NewUsageError(fmt.Errorf("invalid filename mode %q: expected name or name-id", filenameMode))
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			nodeIDs, err := figma.ResolveRequiredNodeIDs(input, nodeID, "assets")
			if err != nil {
				return cli.NewUsageError(err)
			}
			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			filename := assetFilename
			if filenameMode == "name" {
				filename = func(asset extract.Asset) string { return assetNameFilename(asset, trimNamePrefix) }
			}

			// Application service and edge adapters are wired from the
			// Figma client. The application itself owns filtering,
			// collision policy, and manifest construction; the
			// adapters own transport + filesystem.
			app := deps.AssetApplication()
			manifest, err := app.Run(cmd.Context(), assets.ApplicationRequest{
				Source:          assetsedge.NewFigmaAssetSource(client, input.FileID, nodeIDs),
				URLSource:       assetsedge.NewFigmaExportURLSource(client, input.FileID),
				Sink:            newAssetsSink(client, downloadClient),
				Kind:            kind,
				Format:          format,
				NameFilter:      nameFilter,
				Filename:        filename,
				OutputDirectory: outputDirectory,
			})
			if err != nil {
				return err
			}

			asJSON, _ := cmd.Flags().GetBool("json")
			if asJSON {
				if err := cli.NewPrinter(cmd).Structured(manifest); err != nil {
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

// newAssetsSink builds the production HTTP/filesystem sink, honouring
// the optional downloadClient dependency the command carries for tests.
func newAssetsSink(client *figma.Client, downloadClient *http.Client) assets.AssetSink {
	httpClient := downloadClient
	if httpClient == nil {
		httpClient = client.HTTP
	}
	return assetsedge.NewHTTPAssetSink(httpClient)
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
