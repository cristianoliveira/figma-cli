# Purpose

`internal/artifact` owns durable artifact persistence: writing bytes to a destination path. It exposes a narrow consumer-owned `Writer` port and a filesystem adapter.

# Boundaries

The `Writer` port has one responsibility (`Write(ctx, path, data) error`): no reads, downloads, rendering, or arbitrary I/O. The `FileWriter` adapter owns parent-directory creation (0755), fixed file permissions (0644), open/write/close, overwrite semantics, context cancellation, and OS error translation via `internal/operr`.

# Connections

- [Output rendering](internal/output/AGENTS.md): renders result envelopes; does not persist.
- `internal/operr`: provides the neutral artifact-access classification.
- [Asset edge adapters](internal/assetsedge/AGENTS.md): reuse `ClassifyFileError` for streaming downloads.

# Landmarks

- `internal/artifact/artifact.go:Writer`: the consumer-owned persistence port.
- `internal/artifact/fs.go:FileWriter.Write`: the filesystem adapter.

# Placement

Put generic byte persistence here. Keep URL-aware asset download/streaming in `internal/assetsedge`, and rendering in `internal/output`. Do not add a catch-all repository or I/O interface.
