# Figma CLI

Read Figma designs from the terminal. Export layout, text, assets, CSS, and tokens.
Compare a rendered UI with its reference using deterministic image measurements.
The commands serve developers, scripts, and coding agents that need focused design
data and repeatable checks instead of manual inspection.

This repository ships two tools:

- **`figma`** reads design data from the Figma API. It requires a Figma access token.
- **`pixel-perfect`** compares PNG files. It works offline without Figma credentials
  or a vision model.

Use them together for design handoff and visual checks, or use either on its own.
Neither tool edits Figma designs or captures browser screenshots.

## Install

### Build from source

Requires Git and Go 1.25.5 or newer. Run these commands from your terminal:

```bash
git clone https://github.com/cristianoliveira/figma-cli.git
cd figma-cli
go build -o bin/figma ./cmd/figma
go build -o bin/pixel-perfect ./cmd/pixel-perfect
export PATH="$PWD/bin:$PATH"
```

To install into your Go binary directory instead, run this from the repository
root, then add your Go binary directory to `PATH`:

```bash
go install ./cmd/figma ./cmd/pixel-perfect
```

### Nix

From the repository root:

```bash
nix run .#figma -- --help
nix run .#pixel-perfect -- --help
```

Use `nix build .#figma .#pixel-perfect` to build both tools, or `nix develop` to
enter the development shell.

## Configure Figma access

Create a [Figma personal access token](https://www.figma.com/developers/api#access-tokens)
with access to the files you need. Set it in your shell, then check access:

```bash
export FIGMA_ACCESS_TOKEN="replace-with-your-token"
figma me
```

Prefer your shell's secret manager or CI secret store. Do not commit tokens.
The CLI reads environment variables; it does **not** load `.env` files itself.
If you use an environment loader, configure it to export `FIGMA_ACCESS_TOKEN`.

Running `figma` or `pixel-perfect` without arguments shows the executable path
and suggested commands. These views do not make network requests or read images.
`figma me`, unlike the no-argument view, checks access with Figma.

## Quick start

Replace the example URL with a frame URL from your Figma file. Quote URLs so
shell characters such as `&` are passed unchanged.

```bash
FIGMA_URL="https://www.figma.com/design/KEY/App?node-id=42-1"

figma inspect --handoff "$FIGMA_URL"
figma layout "$FIGMA_URL"
figma assets --output ./assets "$FIGMA_URL"
figma export --output reference.png "$FIGMA_URL"
```

Capture `actual.png` from your implementation with your browser tooling. Then run:

```bash
pixel-perfect reference.png actual.png --report visual-diff.html
```

The comparison writes a difference mask and reports metrics. Differences alone
**do not fail the command**. Add a validation limit for CI, such as
`--max-changed-ratio 0` when you require an exact match at the default threshold.
Both prepared images must have equal dimensions; the tool does not resize or
align them automatically.

## Features

### Find and inspect designs

| Command | Use it to |
| --- | --- |
| `me` | Check your identity and Figma access. |
| `projects`, `files` | List a team's projects and a project's files. |
| `meta` | Read file metadata. |
| `frames` | List screen frames in a selected page or section. |
| `find` | Search layers by name, type, or both. |
| `inspect` | Read node bounds, layout, appearance, text, and component properties. |
| `layout` | Read an ordered frame tree with copy and optional spacing measurements. |
| `layout compare` | Compare explicitly selected responsive frames. |
| `texts`, `colors` | Extract ordered copy and a color palette. |
| `components` | List components, sets, instances, and instance usage. |

Start with a small scope, then inspect only the nodes you need:

```bash
figma find --name Button --limit 10 "$FIGMA_URL"
figma inspect --handoff "$FIGMA_URL"
figma layout --depth 2 --measure-spacing "$FIGMA_URL"
figma texts "$FIGMA_URL"
figma components --kind instance --usage "$FIGMA_URL"
```

Names can repeat. Use returned node IDs to select the right element with `--id`
(or its alias `--node`). A URL's `node-id` supplies the scope when no override is set.

For responsive comparison, select at least two frames in the intended order:

```bash
figma layout compare --id 42:1 --id 42:2 "$FIGMA_URL"
```

This compares design frames; it does not infer or test browser breakpoints.

### Generate implementation inputs

| Command | Output |
| --- | --- |
| `css` | CSS rules from layout and appearance, optionally including descendants. |
| `tokens` | CSS variables, Tailwind theme extension, or style-dictionary JSON. |
| `assets` | Batch image, instance, and vector exports with a manifest. |
| `export` | One node as PNG, JPG, SVG, or PDF; optional export metadata. |

```bash
figma css --recursive --output design.css "$FIGMA_URL"
figma tokens --format css --output tokens.css "$FIGMA_URL"
figma tokens --format tailwind --output tailwind.tokens.js "$FIGMA_URL"
figma tokens --format json --output tokens.json "$FIGMA_URL"
figma export --format png --width 800 --output reference.png "$FIGMA_URL"
```

Token source selection defaults to Variables, then Styles, then a document scan.
Scan tokens are named by value. Use `--scan-fallback=false` for named sources only,
or pin `--source variables`, `--source styles`, or `--source scan` in CI. Automatic
selection can fall back after API errors as well as empty results. Team-library
extraction through `tokens --team` is not supported.

Asset downloads fail on partial results unless you set `--allow-partial`. Check the
manifest before using that option. Generated CSS is a starting point, not a complete
responsive implementation.

### Review components, comments, and history

```bash
figma components --diff --codebase ./src/components "$FIGMA_URL"
figma comments --state open "$FIGMA_URL"
figma versions "$FIGMA_URL"
figma changes --from VERSION_A --to VERSION_B "$FIGMA_URL"
figma diff text --from VERSION_A --to VERSION_B "$FIGMA_URL"
figma diff blame --to VERSION_B "$FIGMA_URL"
```

Replace version placeholders with IDs from `versions`.

- Component/codebase comparison matches normalized names from immediate component
  directories and root source files. It does not compare behavior, source semantics,
  or visual appearance.
- Comments support node scope, thread state, author, date range, and ancestor filters.
- `changes` reports frontend-relevant structural differences. `diff text` compares
  copy; `diff blame` searches history for the version that introduced target text.

## Keep output small and predictable

Structured results use **TOON** by default. Pass global `--json` for compatibility
JSON. CSS, token files, recursive text views, and probe/scan CSV keep their own formats.

Most collection commands return at most 100 local results. Read `total` and
`truncated` rather than assuming every result was returned. Use the emitted hint
or `--full` when needed; do not combine `--full` with `--limit`. API pagination and
command-specific depth controls are separate.

`layout` defaults to depth 4. Its `--full` removes the local depth bound and cannot
be combined with explicit `--depth`. For a compact recursive inspection:

```bash
figma inspect --recursive --depth 3 --format text \
  --fields name,type,relativeBounds,layout.mode,layout.gap,fills \
  "$FIGMA_URL"
```

See [command contracts](docs/command-contracts.md) for output fields, bounds, and
compatibility rules. Run `figma COMMAND --help` for local flags and examples.

## Compare a design with a browser capture

Export the same node for the reference, logical crop metadata, and annotations:

```bash
figma export --output reference.png --metadata reference.json "$FIGMA_URL"
figma inspect --recursive --annotations-output annotations.json "$FIGMA_URL"

pixel-perfect reference.png actual.png \
  --reference-metadata reference.json \
  --annotations annotations.json \
  --overlay overlay.png \
  --report visual-diff.html
```

The implementation screenshot must use the same logical bounds and scale.
Annotations label intersecting mismatch regions; they do not change measurements.
An explicit inspect `--depth` also bounds annotation traversal. The stdout result
limit does not limit the annotation file.

For closer diagnosis:

```bash
pixel-perfect probe reference.png actual.png --at 20,20
pixel-perfect scan reference.png actual.png --row 20
```

Probe and scan use CSV by default, with a limit of 25 points or 25 runs per image.
Use `--format json`, `--limit`, or `--full` as needed. If the comparison used crops,
pass the same crop options to probe and scan.

For repeatable captures, fix the browser, OS, fonts, viewport, device scale, and
application state. Wait for fonts and disable animations. Review behavior and
accessibility separately: matching pixels do not prove a working UI.

See the [pixel-perfect guide](cmd/pixel-perfect/README.md) for metrics, crops,
profiles, masks, validation gates, reports, and optional visual descriptions.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Success, including empty queries or comparisons without a failed gate. |
| `1` | Operational failure, failed comparison gate, or a documented quiet no-match. |
| `2` | Invalid arguments or flags. |

Results and structured errors go to stdout. Diagnostics go to stderr. A failed
pixel validation gate still emits the comparison result to stdout and explains
the failure on stderr. Check the exit code before treating output as a success.

## Development

Run development commands from the repository root. `nix develop` supplies Go,
`goimports`, `golangci-lint`, and test tooling.

Start with tests for the package you change:

```bash
go test ./internal/extract -run TestExtractLayout -count=1
go test ./internal/imagediff -run TestCompareImages -count=1
git diff --check
```

Use the repository watcher or CI for full checks. Run binary smoke tests with
`-count=1` because they build subprocesses. Live Figma checks are
[opt-in](docs/live-figma-smoke.md); normal tests do not require credentials.

### Code layout

| Directory | Responsibility |
| --- | --- |
| `cmd/` | Command definitions and executable entry points. |
| `internal/cli/`, `internal/output/` | Dependency wiring, formats, errors, and output contracts. |
| `internal/figma/` | Figma inputs, HTTP, and typed API adapters. |
| `internal/extract/` | Pure document transformations. |
| `internal/assets/`, `internal/comments/`, `internal/diff/` | Export, comment, and history workflows. |
| `internal/components/`, `internal/annotations/` | Name-based parity checks and coordinate annotations. |
| `internal/imagediff/` | Generic PNG measurements, independent of Figma. |
| `internal/pixelperfectcmd/`, `internal/pixelperfectreport/` | Image workflow and HTML reports. |
| `internal/imagecontext/` | Optional provider descriptions, separate from measurements. |

Regenerate models in `internal/figma/api/` from `openapi/`; do not edit generated
code by hand. Keep commands focused on wiring and pure logic in its owning package.

### More documentation

- [Figma agent workflow](skills/figma-cli/SKILL.md)
- [Figma-to-browser workflow](skills/figma-pixel-perfect-loop/SKILL.md)
- [Screenshot implementation workflow](skills/pixel-perfect/SKILL.md)
- [Skill evaluation guide](skills/pixel-perfect/evals/README.md)
- [Open plans](plans/todo/README.md)

## License

[MIT](LICENSE)
