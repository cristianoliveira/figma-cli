# Agent Instructions for `cmd/pixel-perfect/`

## Purpose

This directory is the composition root for the standalone `pixel-perfect` PNG comparison CLI. It does not require Figma credentials.

## Rules

- Keep `main.go` limited to process wiring, command execution, stderr reporting, and exit status.
- Put Cobra flags and command orchestration in `internal/pixelperfectcmd`.
- Put deterministic image decoding, metrics, regions, masks, overlays, probes, and scans in `internal/imagediff`.
- Put HTML report rendering in `internal/pixelperfectreport`.
- Keep visual context advisory: it must not alter deterministic metrics, classifications, or validation gates.
- Never silently resize, align, or mutate comparison inputs. Reject incompatible dimensions and report suggested offsets only as evidence.
- Preserve stable JSON/CSV output because agents and CI consume it.

## Testing

- Test behavior in the owning `internal/*` package rather than through `main.go`.
- Cover successful comparisons and invalid input or validation-gate failures.
- Use generated or checked-in deterministic PNG fixtures; do not require network access unless testing an injected visual-context adapter.
- Run `go test ./internal/imagediff ./internal/pixelperfectcmd ./internal/pixelperfectreport`, then `go test ./...`.
