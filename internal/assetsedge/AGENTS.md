# Purpose

`internal/assetsedge` implements the asset application ports (`AssetSource`, `ExportURLSource`, `AssetSink`) against Figma, HTTP, and the filesystem. It owns asset-specific URL download and streaming file delivery.

# Boundaries

This package is infrastructure at the edge. The application (`internal/assets`) owns policy; this package owns transport and persistence mechanics. It may import Figma, HTTP, and filesystem APIs. Generic byte persistence for non-asset artifacts lives in `internal/artifact`.

# Connections

- [Asset application](internal/assets/AGENTS.md): owns the ports this package implements.
- [Figma boundary](internal/figma/AGENTS.md): provides the Figma client/transport.
- [Artifact persistence](internal/artifact/AGENTS.md): provides `ClassifyFileError` for filesystem failures.

# Landmarks

- `internal/assetsedge/figma_source.go:FigmaAssetSource`: fetches asset candidates.
- `internal/assetsedge/url_source.go:FigmaExportURLSource`: resolves export URLs.
- `internal/assetsedge/http_sink.go:HTTPAssetSink`: downloads a URL and streams to a file.
- `internal/assetsedge/download.go:DownloadFile`: standalone single-asset download.
- `internal/assetsedge/classify.go:classifyDownloadStatus`: safe non-200 status classification.

# Placement

Put asset-specific URL download/streaming here. Do not force URL download concerns into the generic byte-artifact store (`internal/artifact`).
