# Figma CLI

**Give coding agents design data to implement Figma UIs.**

Screenshots show appearance, but hide layout rules, component relationships, and
design tokens. Figma CLI reads those facts from Figma and exports focused,
implementation-ready data and assets.

The workflow has two steps:

1. **[Understand the design](#understand-the-design):** read structure, copy, styles,
   and history instead of guessing from screenshots.
2. **[Generate implementation inputs](#generate-implementation-inputs):** export
   assets, CSS, tokens, and reference images instead of copying values by hand.

Figma CLI requires network access and a Figma access token. It does not edit Figma
designs, write your application, capture browser screenshots, or compare images.

New here? [Install the CLI](#install) and [configure access](#configure-figma-access)
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
figma export --output reference.png "$FIGMA_URL"
```

| Command | Available output |
| --- | --- |
| `css` | CSS rules from layout and appearance, optionally including descendants. |
| `tokens` | CSS variables, Tailwind theme extension, or style-dictionary JSON. |
| `assets` | Batch image, instance, and vector exports with a manifest. |
| `export` | A node as PNG, JPG, SVG, or PDF; raster scale or target width. |

### Choose explicit sources for repeatable output

Tokens default to Variables, then Styles, then a document scan. Scan tokens are
named by value. Automatic selection can fall back after API errors as well as
empty results.

For CI, pin `--source variables`, `--source styles`, or `--source scan`. Use
`--scan-fallback=false` for named sources only. `tokens --team` is not supported.

Asset downloads fail on partial results unless you set `--allow-partial`. Inspect
the manifest before accepting a partial export.

## Install

### Build from source

Requires Git and Go 1.25.5 or newer:

```bash
git clone https://github.com/cristianoliveira/figma-cli.git
cd figma-cli
go build -o bin/figma ./cmd/figma
export PATH="$PWD/bin:$PATH"
```

To install into your Go binary directory instead, run this from the repository
root, then add that directory to `PATH`:

```bash
go install ./cmd/figma
```

### Nix

From the repository root:

```bash
nix run .#figma -- --help
```

Use `nix build .#figma` to build the CLI. Use `nix develop` for the development
shell.

## Configure Figma access

Create a [Figma personal access token](https://www.figma.com/developers/api#access-tokens)
with access to the files you need. Export it in your shell and verify access:

```bash
export FIGMA_ACCESS_TOKEN="replace-with-your-token"
figma me
```

Prefer a shell secret manager or CI secret store. Never commit tokens. The CLI
reads environment variables; it does **not** load `.env` files itself.

Running `figma` without arguments shows its executable path and suggested commands
without a network request. Use `figma me` to check credentials with Figma.

## Use predictable command contracts

**Request only what you need, and check the exit code before consuming results.**

| Contract | Behavior |
| --- | --- |
| Structured output | TOON by default; global `--json` selects compatibility JSON. |
| Text and file output | CSS, tokens, and recursive text views retain their own formats. |
| Collection bounds | Most collection commands return at most 100 local results. Read `total` and `truncated`. |
| Layout depth | Defaults to 4; `--full` removes the local depth bound. |
| Streams | Results and structured errors use stdout; diagnostics use stderr. |
| Exit codes | `0` success; `1` operational failure or documented quiet no-match; `2` usage error. |

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

See [command contracts](docs/command-contracts.md) for exact semantics,
and use each command's `--help` for supported options.

## Develop and extend

Keep commands focused on wiring. Keep Figma transport, document extraction, and
output rendering separate so each can be tested independently.

| Directory | Responsibility |
| --- | --- |
| `cmd/` | Command definitions and executable entry points. |
| `internal/cli/`, `internal/output/` | Dependency wiring, formats, and error contracts. |
| `internal/figma/` | Figma input parsing, HTTP, and typed API adapters. |
| `internal/extract/` | Pure document transformations. |
| `internal/assets/`, `internal/comments/`, `internal/diff/` | Export, comment, and history workflows. |
| `internal/components/` | Name-based component parity. |

Run focused tests from the repository root:

```bash
go test ./internal/extract -run TestExtractLayout -count=1
git diff --check
```

Use the repository watcher or CI for full checks. Run subprocess-based smoke tests
with `-count=1`. Live Figma checks are [opt-in](docs/live-figma-smoke.md); normal
tests need no credentials. Regenerate API models from `openapi/` instead of editing
`internal/figma/api/` generated code by hand.

### Further reading

- [Figma agent workflow](skills/figma-cli/SKILL.md)
- [Open plans](plans/todo/README.md)

## License

[MIT](LICENSE)
