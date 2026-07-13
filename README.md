# Figma CLI

**Turn Figma designs into implementation-ready data—and prove the result matches.**

Figma CLI gives coding agents a direct, scriptable path from a Figma URL to structured design context, generated assets, CSS, tokens, and deterministic visual validation.

The main goal is agent self-validation without relying on vision models. Instead of asking a model to repeatedly inspect full screenshots, agents can query semantic design data, measure raster differences, inspect exact pixels, and narrow failures to actionable regions using deterministic tools.

No clicking through panels. No manually copying node IDs. No guessing from screenshots. No spending vision tokens just to learn that a button is two pixels too wide.

```bash
figma inspect --handoff "https://www.figma.com/design/KEY/App?node-id=42-1"
figma assets --output ./assets "https://www.figma.com/design/KEY/App?node-id=42-1"
figma tokens --format css "https://www.figma.com/design/KEY/App"
pixel-perfect reference.png implementation.png --report visual-diff.html
```

## Why this project exists

A Figma design contains more than pixels: layout constraints, component relationships, text hierarchy, design tokens, assets, comments, and version history. Screenshots hide that information, while the Figma UI makes it difficult to automate.

Figma CLI exposes the design as stable command-line contracts built for humans, scripts, CI pipelines, and AI agents.

It ships two complementary tools:

- **`figma`** explains what the design intends.
- **`pixel-perfect`** measures what the implementation rendered.

Together they close the loop from design exploration to visual verification, giving an agent evidence to inspect, implement, measure, diagnose, and correct its own work.

### Progressive disclosure by design

Agents should request only the context needed for the next decision. The CLI supports that workflow from broad discovery to exact evidence:

1. Start with compact file or frame context using `meta` and `layout`.
2. Narrow by name or type with `find`.
3. Inspect only the relevant node with `inspect`, `texts`, `colors`, or `components`.
4. Generate only required implementation inputs with `css`, `tokens`, `assets`, or `export`.
5. Measure the rendered result with `pixel-perfect`.
6. Drill into mismatch regions with `probe` and `scan` instead of sending entire screenshots back to a vision model.

This progressive disclosure keeps outputs focused and token use proportional to the problem. Vision remains an optional advisory layer, not a prerequisite for validation.

## What you can do

### Understand a design before writing code

```bash
# Get file-level context
figma meta <figma-url>

# Find every matching layer, with node IDs
figma find --name "Button" <figma-url>

# Inspect one node as an implementation handoff
figma inspect --handoff <figma-node-url>

# Read layout, copy, components, and colors
figma layout <figma-node-url>
figma texts <figma-node-url>
figma components <figma-node-url>
figma colors <figma-node-url>
```

Layer names do not need to be unique. Results retain node IDs and hierarchy so consumers can act on the right element instead of guessing.

### Generate implementation inputs

```bash
# Scriptable equivalent of Dev Mode CSS
figma css --recursive <figma-node-url> > design.css

# Produce CSS variables, Tailwind configuration, or JSON tokens
figma tokens --format css <figma-url> > tokens.css
figma tokens --format tailwind <figma-url> > tailwind.tokens.js
figma tokens --format json <figma-url> > tokens.json

# Download assets from an entire node tree
figma assets --output ./public/assets <figma-node-url>

# Export one exact node
figma export --format png --output reference.png <figma-node-url>
```

### Understand design changes

```bash
figma versions <figma-url>
figma diff text --from <version-1> --to <version-2> <figma-url>
figma diff blame --to <version> <figma-node-url>
figma changes --from <version-1> --to <version-2> <figma-url>
```

Use design history as structured evidence: identify changed copy, find when text was introduced, or detect frontend-relevant structural changes before implementation drifts.

### Verify the rendered result

`pixel-perfect` is a standalone PNG comparison CLI. It does not require a Figma token and does not silently resize or align images.

```bash
pixel-perfect reference.png implementation.png \
  --threshold 8 \
  --overlay diff-overlay.png \
  --report visual-diff.html \
  --max-changed-ratio 0.02
```

It reports deterministic evidence including:

- changed-pixel ratio and normalized RMSE
- perceptual, luminance, alpha, and edge differences
- mismatch bounds and connected regions
- dominant color pairs
- likely geometry, fill, or sparse-raster mismatches
- optional masks, directional overlays, and HTML reports
- non-zero exit status when CI thresholds fail

For targeted diagnosis:

```bash
pixel-perfect probe reference.png implementation.png --at 316,300
pixel-perfect scan reference.png implementation.png --y 300
```

## Compact implementation outlines

Recursive inspection can render only properties needed for current task, avoiding ad hoc Python filters and large JSON payloads:

```bash
figma inspect --recursive --depth 3 \
  --format text \
  --fields name,type,relativeBounds,layout.mode,layout.gap,layout.paddingTop,layout.paddingRight,layout.paddingBottom,layout.paddingLeft,fills \
  <figma-node-url>
```

`--fields` accepts comma-separated inspect JSON properties and nested paths. Child hierarchy remains implicit and is represented by indentation. Unknown fields fail explicitly. JSON remains default output for backward compatibility.

## The design-to-code loop

The two CLIs integrate without coupling generic image comparison to Figma:

```bash
# 1. Export the reference and its logical crop metadata
figma export \
  --format png \
  --output reference.png \
  --metadata reference.json \
  <figma-node-url>

# 2. Export design-node coordinates for meaningful mismatch labels
figma inspect \
  --recursive \
  --annotations-output annotations.json \
  <figma-node-url>

# 3. Compare against your implementation screenshot
pixel-perfect reference.png implementation.png \
  --reference-metadata reference.json \
  --annotations annotations.json \
  --overlay diff.png \
  --report visual-diff.html
```

Now a mismatch is not merely “some pixels changed.” It can be localized to a design region, measured objectively, and traced back to semantic Figma context.

## Built for agent self-validation

- Uses semantic API data and deterministic image metrics before optional vision.
- Supports progressive disclosure from file metadata to nodes, regions, rows, and individual pixels.
- Produces compact, machine-readable evidence agents can use in an iterative correction loop.
- Accepts full Figma URLs or bare file keys.
- Infers node scope from `node-id` in URLs.
- Normalizes user-facing and API-facing node ID formats.
- Emits structured, stable JSON for query commands.
- Supports global `--json` envelopes for text and file-producing commands.
- Uses deterministic formatting for generated CSS and tokens.
- Fails explicitly on invalid input and unsupported ambiguity.
- Keeps network access separate from pure extraction and image analysis.

This makes the CLI a reliable capability layer for coding agents: enough tooling to validate their own implementation while controlling context size and avoiding unnecessary vision calls.

## Command overview

### `figma`

| Goal | Commands |
| --- | --- |
| Validate access and navigate workspaces | `me`, `projects`, `files` |
| Explore files and nodes | `meta`, `find`, `inspect`, `layout` |
| Extract implementation context | `texts`, `colors`, `components`, `css`, `tokens` |
| Produce files | `assets`, `export` |
| Review collaboration and history | `comments`, `versions`, `diff`, `changes` |
| Compare responsive frames | `layout compare` |

Run `figma <command> --help` for command-specific examples and flags.

### `pixel-perfect`

| Command | Purpose |
| --- | --- |
| `pixel-perfect <reference> <actual>` | Compare PNGs and produce metrics and artifacts |
| `pixel-perfect probe` | Inspect exact colors at points or sampled lines |
| `pixel-perfect scan` | Inspect compact color runs along a row or column |

See [`cmd/pixel-perfect/README.md`](cmd/pixel-perfect/README.md) for metrics, profiles, masks, crops, visual context, and CI gates.

## Installation

### Go

Requires Go 1.25.5 or newer.

```bash
go install github.com/cristianoliveira/figma-cli/cmd/figma@latest
go install github.com/cristianoliveira/figma-cli/cmd/pixel-perfect@latest
```

### Build from source

```bash
git clone https://github.com/cristianoliveira/figma-cli.git
cd figma-cli
go build -o figma ./cmd/figma
go build -o pixel-perfect ./cmd/pixel-perfect
```

### Nix

```bash
nix develop
```

The development shell provides the expected Go toolchain, `golangci-lint`, and `goimports`.

## Configuration

`figma` requires a [Figma personal access token](https://www.figma.com/developers/api#access-tokens):

```bash
export FIGMA_ACCESS_TOKEN="your-personal-access-token"
figma me
```

You can also place it in a local `.env` file:

```dotenv
FIGMA_ACCESS_TOKEN=your-personal-access-token
```

`pixel-perfect` does not require Figma credentials. Its optional visual-context feature requires a configured OpenRouter or OpenAI provider.

## Architecture

```text
cmd/                         Cobra command contracts and entry points
internal/cli/                Runtime dependency wiring
internal/env/                Environment configuration
internal/output/             Stable stdout and JSON envelopes
internal/figma/              Figma URLs, node IDs, HTTP, and API boundary
internal/figma/api/          Generated OpenAPI models
internal/extract/            Pure document-tree transformations
internal/assets/             Asset discovery and download workflows
internal/comments/           Comment retrieval and node scoping
internal/diff/               Design-history diff workflows
internal/imagediff/          Generic deterministic PNG comparison
internal/pixelperfectcmd/    Standalone image CLI orchestration
internal/imagecontext/       Optional multimodal descriptions
```

The architectural rule is simple: command code stays thin, Figma networking stays at the Figma boundary, extraction remains testable, and generic image comparison never depends on Figma.

## Development

Enter the reproducible toolchain with `direnv allow` or `nix develop`, then run:

```bash
goimports -w <changed-go-files>
golangci-lint run ./...
go test ./...
```

Generated API models in `internal/figma/api/api.gen.go` come from `openapi/` and must be regenerated rather than edited manually.

## Who this is for

- Developers implementing Figma designs
- Coding agents that need structured design context
- Teams automating design handoff
- CI pipelines enforcing visual regression thresholds
- Design-system maintainers extracting tokens and assets
- Reviewers investigating when and how a design changed

If your current workflow is “open Figma, inspect manually, copy values, take screenshots, and eyeball the result,” this project turns that process into a repeatable interface.

## License

[MIT](LICENSE)
