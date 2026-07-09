# lint (design quality rules)

## Problem

Design files accumulate debt the same way code does. But unlike code, there are no linters.
Common design debt:

- **Naming chaos** — frames named "Frame 123", "Rectangle Copy 15", "Untitled"
- **Detached instances** — a component was pasted and detached, breaking the design system link
- **Hidden layers** — a layer is invisible but still in the tree (dead weight)
- **No auto-layout** — frames positioned manually that should use auto-layout
- **Mixed units** — some text in px, some in rem, no consistency
- **Missing variants** — a component has only one state (Default) when it should have
  hover, active, disabled, focused
- **Unlinked styles** — text/colors that don't use named styles (hard-coded values)

Nobody catches these until:
- Someone tries to reuse a "Frame 123" and has no idea what it is
- A component update doesn't cascade because the instance was detached
- Design-to-code tooling breaks because of unstructured layout

## Success criteria

- A CLI command that scans a Figma file and reports rule violations
- Each rule is toggleable (on/off via flag or config file)
- Output is CI-friendly so it can run in a pre-commit hook or GitHub check
- Rules are extensible — users can add custom rules

## API mapping

| Data | API endpoint |
|---|---|
| Full document tree | `/files/{file_key}` |
| Component metadata | `/files/{file_key}/components` |

## CLI shape

```
figma lint <file-url> [flags]

Flags:
  --rules       Comma-separated rules to run (default: all)
  --page        Lint only a specific page
  --format      Output format: table, json, compact (default: table)
  --config      Path to rules config file (default: .figmalint.yml)
  --fix         Auto-fix where possible (rename, delete)
  --severity    Minimum severity: error, warning, info (default: warning)
```

## Rules

| Rule | Severity | Description |
|---|---|---|
| `unnamed-layers` | error | Flags layers with auto-generated names (e.g., "Rectangle 3") |
| `detached-instances` | error | INSTANCE nodes whose `componentId` is empty |
| `hidden-layers` | warning | Layers with `visible: false` |
| `no-autolayout` | info | FRAME nodes without `layoutMode` set |
| `hardcoded-text-styles` | warning | TEXT nodes not using a named style |
| `hardcoded-colors` | warning | Fill values not referencing a named color style |
| `single-variant-component` | info | COMPONENT_SETS with only 1 variant |
| `layer-name-convention` | error | Names not matching a regex pattern |
| `empty-frames` | warning | FRAME nodes with no children |
| `locked-layers` | info | Nodes with `locked: true` |

## Output example

```
$ figma lint --page "Customization page"

Customization page — 4 errors, 6 warnings, 3 info

errors:
  unnamed-layers: Rectangle 3, Rectangle 3 Copy 15 (2)
  layer-name-convention: "get-more-storage" should be "Button / Get more storage" (kebab-case)
  detached-instances: "Old Button" (node 20089:685899) (1)

warnings:
  hardcoded-text-styles: "Drive" (20089:685898) uses Inter/14 without named style
  hardcoded-colors: "Storage bar" uses #0667C8 without named style
  hidden-layers: "tooltip", "debug-overlay" (2)

info:
  single-variant-component: "Button / Text Button" has 1 variant, consider adding more
  locked-layers: "background-grid" (1)
```

## Config file

```yaml
# .figmalint.yml
rules:
  unnamed-layers: error
  layer-name-convention:
    severity: error
    pattern: "^[A-Z][a-z]+( / [A-Z][a-z]+)*$"  # PascalCase with / separators
  detached-instances: error
  hidden-layers: warning
  no-autolayout: off  # disabled
  hardcoded-text-styles: warning
  hardcoded-colors: warning
  single-variant-component: info
```

## Edge cases

- Pages with hundreds of layers → paginate output, show top N violations
- Non-component INSTANCE nodes (just regular instances) → not a violation, they should have
  a `componentId`
- Figma sections → treat like frames for linting purposes
- Library components → some rules don't apply (e.g., `no-autolayout` may be intentional)
- Rule that needs multiple nodes (e.g., "this fills doesn't match the style") → mark as "info"
  since it might be intentional
