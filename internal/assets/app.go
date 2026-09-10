package assets

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// ErrMissingPort indicates that an asset application call supplied
// candidates (or attempted to) without a URLSource / Sink / Filename
// collaborator. This is an operational configuration error, not an
// item-level export failure.
var ErrMissingPort = errors.New("assets application: URLSource, Sink, and Filename are required when candidates exist")

const (
	assetFormatAuto = "auto"
	assetKindAll    = "all"
	assetKindIcon   = "icon"
	assetKindVector = "vector"
)

// FilenameFunc computes the base name (no extension, no directory) for
// one asset. Collision suffixes and format extensions are added by the
// application.
type FilenameFunc func(extract.Asset) string

// ApplicationRequest captures the user's intent. The application is
// deterministic given the same Source + URLSource + Sink and the same
// request; ordering follows the Source's candidate order.
type ApplicationRequest struct {
	Source    AssetSource
	URLSource ExportURLSource
	Sink      AssetSink

	Kind       string
	Format     string
	NameFilter string
	Filename   FilenameFunc

	OutputDirectory string
}

// Application is the deterministic asset workflow. It owns filtering,
// naming, collision policy, manifest construction, and partial-failure
// policy. It does not import net/http, os, or *figma.Client; transport
// is injected through the three ports above.
type Application struct{}

// NewApplication returns a fresh application instance.
func NewApplication() *Application { return &Application{} }

// Run executes the workflow and returns a complete deterministic
// manifest. A nil Source is treated as "no candidates" (returns an
// empty non-nil manifest); URLSource and Sink errors are recorded as
// item-level failures; later assets always continue.
func (a *Application) Run(ctx context.Context, req ApplicationRequest) (AssetExportManifest, error) {
	if req.Source == nil {
		return emptyManifest(), nil
	}

	candidates, err := req.Source.Candidates(ctx)
	if err != nil {
		// Top-level source failure (e.g. document fetch). The existing
		// contract returned this as an error so the command can fail
		// exit 1; consumers should treat top-level failures as
		// operational errors distinct from item-level export failures.
		return AssetExportManifest{}, err
	}

	kind := req.Kind
	if kind == "" {
		kind = assetKindAll
	}
	format := req.Format
	if format == "" {
		format = assetFormatAuto
	}
	filtered := filterAssets(candidates, kind, format, req.NameFilter)
	manifest := AssetExportManifest{Items: make([]AssetExportItem, 0, len(filtered))}
	if len(filtered) == 0 {
		return manifest, nil
	}
	if req.Sink == nil || req.URLSource == nil {
		// Defensive: missing collaborators would mean an operational
		// configuration error. Surface it so the command fails cleanly.
		return manifest, ErrMissingPort
	}
	filename := req.Filename
	if filename == nil {
		filename = defaultFilename
	}

	usedPaths := make(map[string]int)
	for _, asset := range filtered {
		item := AssetExportItem{
			NodeID: asset.ID,
			Name:   asset.Name,
			Kind:   asset.Kind,
			Format: asset.Format,
		}

		assetURL, urlErr := req.URLSource.ExportURL(ctx, asset.ID, asset.Format)
		if urlErr != nil {
			item.Error = urlErr.Error()
			manifest.Failed++
			manifest.Items = append(manifest.Items, item)
			continue
		}

		basePath := filepath.Join(req.OutputDirectory, filename(asset))
		usedPaths[basePath]++
		path := basePath
		if usedPaths[basePath] > 1 {
			path += "-" + strconv.Itoa(usedPaths[basePath])
		}
		path += "." + asset.Format
		item.Path = path
		if sinkErr := req.Sink.Write(ctx, path, assetURL); sinkErr != nil {
			item.Path = ""
			item.Error = sinkErr.Error()
			manifest.Failed++
		} else {
			manifest.Succeeded++
		}
		manifest.Items = append(manifest.Items, item)
	}
	return manifest, nil
}

// defaultFilename is the fallback base-name function matching the
// previous exporter's behaviour (asset ID, no extension). The command
// always passes a richer function, but the application must not panic
// if the caller omits one.
func defaultFilename(asset extract.Asset) string {
	return asset.ID
}

func emptyManifest() AssetExportManifest {
	return AssetExportManifest{Items: make([]AssetExportItem, 0)}
}

// filterAssets is the deterministic policy that decides which
// candidate nodes survive into the export attempt.
//
// Filtering rules:
//   - kind=icon keeps instance and vector nodes; everything else skips.
//   - kind=all keeps everything; kind=image keeps only image assets;
//     kind=instance keeps only instance assets; kind=vector keeps only
//     vector assets. (Other kinds are rejected by command-layer
//     validation.)
//   - When format is anything other than auto, it overrides the
//     candidate's format so the exporter requests the right asset.
//   - nameFilter is a case-insensitive substring match against the
//     candidate's Name; empty matches every name.
//
// The output preserves the source's candidate order; duplicate node IDs
// are not de-duplicated here — the Source is the source of truth and
// the application must faithfully execute every candidate it
// receives.
func filterAssets(assets []extract.Asset, kind, format, nameFilter string) []extract.Asset {
	filtered := make([]extract.Asset, 0, len(assets))
	for _, asset := range assets {
		if kind == assetKindIcon && asset.Kind != "instance" && asset.Kind != assetKindVector {
			continue
		}
		if kind != assetKindAll && kind != assetKindIcon && asset.Kind != kind {
			continue
		}
		if nameFilter != "" && !strings.Contains(strings.ToLower(asset.Name), strings.ToLower(nameFilter)) {
			continue
		}
		if format != assetFormatAuto {
			asset.Format = format
		}
		filtered = append(filtered, asset)
	}
	return filtered
}
