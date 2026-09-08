# Purpose

`internal/assets` owns Figma asset discovery, export URL resolution, deterministic filenames, downloads, and export manifests.

# Boundaries

It combines Figma asset requests with filesystem/HTTP delivery because it owns the asset workflow. Commands only validate flags and render the resulting manifest. Document traversal and generic output contracts remain elsewhere.

# Connections

- [Figma boundary](internal/figma/AGENTS.md): provides export URLs and document/API access.
- [Extraction](internal/extract/AGENTS.md): provides asset candidates from document trees.
- [Output](internal/output/AGENTS.md): provides filesystem creation for downloaded artifacts.

# Landmarks

- `internal/assets/runner.go:ExportAssets`: coordinates discovery, filtering, and download reporting.
- `internal/assets/asset_exporter.go:AssetExporter.Export`: exports requested items in caller order.
- `internal/assets/export.go:DownloadFile`: writes one downloaded asset to its deterministic path.

# Boundary flows

- Information flow: `internal/extract/assets.go:ExtractAssets` -> `internal/assets/asset_exporter.go:AssetExporter.Export` via `internal/assets/runner.go:ExportAssets`; value: `AssetExportItem`.

# Placement

Place asset-specific naming, partial-result, and delivery policy here. Put generic file creation in output and Figma URL construction in the Figma boundary.
