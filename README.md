# Figma CLI

**Give coding agents design data and measurable feedback to implement Figma UIs.**

Screenshots show appearance, but hide layout rules, component relationships, and
design tokens. Figma CLI reads those facts from Figma and measures differences
between a reference image and your rendered UI. A vision model is not required.

The workflow has three steps:

1. **[Understand the design](#understand-the-design):** read structure, copy, styles,
   and history instead of guessing from screenshots.
2. **[Generate implementation inputs](#generate-implementation-inputs):** export
   assets, CSS, tokens, and reference images instead of copying values by hand.
3. **[Check the result](#check-the-result):** measure image differences, locate
   mismatches, and enforce explicit limits instead of relying only on visual judgment.

Two independent tools support this workflow:

| Tool | Responsibility | Requirements |
| --- | --- | --- |
| `figma` | Read design data and export files. | Network access and a Figma access token. |
| `pixel-perfect` | Compare PNGs and produce metrics and reports. | Local PNG files; no Figma token. |

Use either tool on its own or combine them. Neither edits Figma designs, writes
your application, or captures browser screenshots.

New here? [Install the tools](#install) and [configure access](#configure-figma-access)
before running the examples below.

## Understand the design

**Start with the smallest scope that answers your question.** Find the relevant
frame or node, then request its implementation details.

Replace this URL with your Figma frame URL. Later examples reuse `FIGMA_URL`.
Quote URLs so shell characters such as `&` are passed unchanged.

```bash
FIGMA_URL="https://www.figma.com/design/KEY/App?node-id=42-1"

figma find --name Button --limit 10 "$FIGMA_URL"
figma inspect --handoff "$FIGMA_URL"
figma layout --depth 2 --measure-spacing "$FIGMA_URL"
```

| Question | Commands |
| --- | --- |
| Which projects and files can I access? | `me`, `projects`, `files` |
| What is in this file or page? | `meta`, `frames`, `find` |
| How is this node built? | `inspect`, `layout` |
| What copy, colors, and components does it use? | `texts`, `colors`, `components` |
| How do selected responsive frames differ? | `layout compare` |
| What feedback and changes should I review? | `comments`, `versions`, `changes`, `diff text`, `diff blame` |

Names can repeat. Use returned node IDs with `--id` (or `--node`) to select the
right element. Otherwise, a URL's `node-id` supplies the scope. `frames` discovers
screen frames within a selected page or section.

### Compare structure and review changes

```bash
figma layout compare --id 42:1 --id 42:2 "$FIGMA_URL"
figma components --kind instance --usage "$FIGMA_URL"
figma components --diff --codebase ./src/components "$FIGMA_URL"
figma comments --state open "$FIGMA_URL"
figma versions "$FIGMA_URL"
figma changes --from VERSION_A --to VERSION_B "$FIGMA_URL"
figma diff text --from VERSION_A --to VERSION_B "$FIGMA_URL"
figma diff blame --to VERSION_B "$FIGMA_URL"
```

Replace example frame and version IDs with IDs from your file.

- Responsive comparison uses explicitly selected frames in the order supplied.
  It does not infer or test browser breakpoints.
- Component usage groups instances by component ID. Codebase comparison matches
  normalized names from immediate directories and root source files; it does not
  compare implementation behavior or visual appearance.
- Comments support node scope, thread state, author, date range, and ancestors.
- `changes` reports frontend-relevant structural differences. `diff text` compares
  copy; `diff blame` searches for the version that introduced target text.

## Generate implementation inputs

**Export the data and files your implementation needs.** Keep application design
and behavior in your codebase; generated CSS is a starting point, not a complete UI.

```bash
figma css --recursive --output design.css "$FIGMA_URL"
figma tokens --format css --output tokens.css "$FIGMA_URL"
figma assets --output ./assets "$FIGMA_URL"
figma export --output reference.png --metadata reference.json "$FIGMA_URL"
```

| Command | Available output |
| --- | --- |
| `css` | CSS rules from layout and appearance, optionally including descendants. |
| `tokens` | CSS variables, Tailwind theme extension, or style-dictionary JSON. |
| `assets` | Batch image, instance, and vector exports with a manifest. |
| `export` | A node as PNG, JPG, SVG, or PDF; raster scale or target width; optional metadata. |

### Choose explicit sources for repeatable output

Tokens default to Variables, then Styles, then a document scan. Scan tokens are
named by value. Automatic selection can fall back after API errors as well as
empty results.

For CI, pin `--source variables`, `--source styles`, or `--source scan`. Use
`--scan-fallback=false` for named sources only. `tokens --team` is not supported.

Asset downloads fail on partial results unless you set `--allow-partial`. Inspect
the manifest before accepting a partial export.

## Check the result

**Measure differences, then investigate the smallest relevant region.** Capture
`actual.png` with your browser tooling and compare it with the exported reference:

```bash
pixel-perfect reference.png actual.png \
  --reference-metadata reference.json \
  --overlay overlay.png \
  --report visual-diff.html
```

This writes a difference mask, a directional overlay, and a self-contained report.
Structured output includes changed-pixel ratios, raw and perceptual error metrics,
mismatch bounds, and local color evidence.

Both prepared images must have equal dimensions and matching logical bounds and
scale. The tool does not resize or align them automatically. Export metadata can
apply an explicit logical crop to the reference.

### Locate a mismatch

Export annotations from the same Figma node to label intersecting image regions:

```bash
figma inspect --recursive --annotations-output annotations.json "$FIGMA_URL"

pixel-perfect reference.png actual.png \
  --reference-metadata reference.json --annotations annotations.json
pixel-perfect probe reference.png actual.png \
  --reference-metadata reference.json --at 20,20
pixel-perfect scan reference.png actual.png \
  --reference-metadata reference.json --row 20
```

Choose probe and scan coordinates inside the prepared image. Keep crop options
consistent across comparisons, probes, and scans. Annotations add context without
changing measurements. Explicit inspect depth also bounds annotation traversal;
the stdout result limit does not limit the annotation file.

The [pixel-perfect guide](cmd/pixel-perfect/README.md) covers crops, masks, ignored
regions, profiles, movement suggestions, and optional provider descriptions.

### Set acceptance limits deliberately

Differences alone **do not fail the comparison**. For exact raw-pixel matching at
the default threshold, add a zero changed-ratio limit:

```bash
pixel-perfect reference.png actual.png \
  --reference-metadata reference.json --max-changed-ratio 0
```

For tolerances, calibrate `--max-rmse`, `--max-changed-ratio`, and
`--max-perceptual-changed-ratio` against accepted captures and known regressions.
Do not copy a threshold from another UI or raise it just to make a check pass.

Fix browser, OS, fonts, viewport, device scale, and application state between
captures. Wait for fonts and disable animations. Keep behavior, accessibility,
and human visual review separate: lower image error does not prove a working or
acceptable UI. Optional model descriptions are advisory, not validation evidence.

## Install

### Build from source

Requires Git and Go 1.25.5 or newer:

```bash
git clone https://github.com/cristianoliveira/figma-cli.git
cd figma-cli
go build -o bin/figma ./cmd/figma
go build -o bin/pixel-perfect ./cmd/pixel-perfect
export PATH="$PWD/bin:$PATH"
```

To install into your Go binary directory instead, run this from the repository
root, then add that directory to `PATH`:

```bash
go install ./cmd/figma ./cmd/pixel-perfect
```

### Nix

From the repository root:

```bash
nix run .#figma -- --help
nix run .#pixel-perfect -- --help
```

Use `nix build .#figma .#pixel-perfect` to build both tools. Use `nix develop` for
the development shell.

## Configure Figma access

Create a [Figma personal access token](https://www.figma.com/developers/api#access-tokens)
with access to the files you need. Export it in your shell and verify access:

```bash
export FIGMA_ACCESS_TOKEN="replace-with-your-token"
figma me
```

Prefer a shell secret manager or CI secret store. Never commit tokens. The CLI
reads environment variables; it does **not** load `.env` files itself.

Running either tool without arguments shows its executable path and suggested
commands without network requests or image reads. Use `figma me` to check
credentials with Figma. Local image comparison needs no credentials; optional
visual descriptions require provider configuration and send images to that provider.

## Use predictable command contracts

**Request only what you need, and check the exit code before consuming results.**

| Contract | Behavior |
| --- | --- |
| Structured output | TOON by default; global `--json` selects compatibility JSON. |
| Text and file output | CSS, tokens, and recursive text views retain their own formats. |
| Collection bounds | Most collection commands return at most 100 local results. Read `total` and `truncated`. |
| Layout depth | Defaults to 4; `--full` removes the local depth bound. |
| Pixel inspection | Probe/scan default to CSV and 25 points or 25 runs per image; use `--format json` for JSON. |
| Streams | Results and structured errors use stdout; diagnostics use stderr. |
| Exit codes | `0` success; `1` operational failure, failed gate, or documented quiet no-match; `2` usage error. |

Use truncation hints or `--full` to retrieve more output. Do not combine `--full`
with `--limit`, or layout `--full` with explicit `--depth`. Local limits, traversal
depth, and API pagination are separate controls.

For a compact implementation outline, select fields instead of filtering a large
payload yourself:

```bash
figma inspect --recursive --depth 3 --format text \
  --fields name,type,relativeBounds,layout.mode,layout.gap,fills \
  "$FIGMA_URL"
```

A failed pixel gate still writes the comparison result to stdout and a diagnosis
to stderr. See [command contracts](docs/command-contracts.md) for exact semantics,
and use each command's `--help` for supported options.

## Develop and extend

Keep commands focused on wiring. Keep Figma transport, document extraction, image
measurement, and output rendering separate so each can be tested independently.

| Directory | Responsibility |
| --- | --- |
| `cmd/` | Command definitions and executable entry points. |
| `internal/cli/`, `internal/output/` | Dependency wiring, formats, and error contracts. |
| `internal/figma/` | Figma input parsing, HTTP, and typed API adapters. |
| `internal/extract/` | Pure document transformations. |
| `internal/assets/`, `internal/comments/`, `internal/diff/` | Export, comment, and history workflows. |
| `internal/components/`, `internal/annotations/` | Name-based parity and coordinate annotations. |
| `internal/imagediff/` | Generic PNG measurements, independent of Figma. |
| `internal/pixelperfectcmd/`, `internal/pixelperfectreport/` | Image workflow and HTML reports. |
| `internal/imagecontext/` | Optional provider descriptions. |

Run focused tests from the repository root:

```bash
go test ./internal/extract -run TestExtractLayout -count=1
go test ./internal/imagediff -run TestCompareImages -count=1
git diff --check
```

Use the repository watcher or CI for full checks. Run subprocess-based smoke tests
with `-count=1`. Live Figma checks are [opt-in](docs/live-figma-smoke.md); normal
tests need no credentials. Regenerate API models from `openapi/` instead of editing
`internal/figma/api/` generated code by hand.

### Further reading

- [Figma agent workflow](skills/figma-cli/SKILL.md)
- [Figma-to-browser workflow](skills/figma-pixel-perfect-loop/SKILL.md)
- [Screenshot implementation workflow](skills/pixel-perfect/SKILL.md)
- [Skill evaluation guide](skills/pixel-perfect/evals/README.md)
- [Open plans](plans/todo/README.md)

## License

[MIT](LICENSE)
