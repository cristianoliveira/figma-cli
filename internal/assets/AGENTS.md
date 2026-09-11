# Purpose

`internal/assets` owns the asset export application workflow: candidate filtering, deterministic filename/collision policy, manifest construction, and partial-failure policy. It is pure (no HTTP/filesystem) and depends on consumer-owned ports.

# Boundaries

The application owns policy, not transport or persistence. Figma fetching, URL resolution, and HTTP/filesystem delivery live in the [assetsedge](internal/assetsedge/AGENTS.md) edge adapter package, which implements the ports defined here. Commands only validate flags and render the resulting manifest.

# Connections

- [Figma boundary](internal/figma/AGENTS.md): provides export URLs and document/API access.
- [Extraction](internal/extract/AGENTS.md): provides asset candidates from document trees.
- [Asset edge adapters](internal/assetsedge/AGENTS.md): implement `AssetSource`, `ExportURLSource`, and `AssetSink` against Figma/HTTP/filesystem.

# Landmarks

- `internal/assets/ports.go`: `AssetSource`, `ExportURLSource`, `AssetSink` ports.
- `internal/assets/app.go:Application.Run`: orchestrates filtering, naming, collision, and manifest policy.
- `internal/assets/manifest.go`: `AssetExportItem` and `AssetExportManifest`.
- `internal/assets/exportpath.go:DefaultExportOutputPath`: deterministic default path.

# Boundary flows

- Information flow: `internal/extract/assets.go:ExtractAssets` -> `internal/assets/app.go:Application.Run` via the Figma `AssetSource` adapter; value: `AssetExportManifest`.

# Placement

Place asset-specific naming, partial-result, and delivery policy here. Put Figma/HTTP/filesystem mechanics in `internal/assetsedge`, Figma URL construction in the Figma boundary, and generic byte persistence in `internal/artifact`.
