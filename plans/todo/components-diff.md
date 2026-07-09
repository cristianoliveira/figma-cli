# components --diff (design system parity)

## Problem

In any org with a shared component library, the question "do we have all Figma components
in code?" comes up constantly:

- **Before a release** — are there new components in the design library that haven't been built?
- **After a redesign** — which existing components changed and need code updates?
- **During onboarding** — what's in the library? What's in code? Where's the gap?

Currently this is answered via:
- A manually-maintained spreadsheet (goes stale within a sprint)
- A design audit meeting (2–4 people × 1 hour, repeat quarterly)
- "I think that component exists… does it?" (the most common answer)

## Success criteria

- List all components from a Figma file or team library
- Compare against a codebase directory (by component name)
- Show: matched (exists in both), missing (in Figma, not in code), extra (in code, not in Figma)
- Optionally detect *drift* — components that exist in both but changed since last sync
- Output is CI-friendly so it can run as a check on every PR to the design file

## API mapping

| Data | API endpoint |
|---|---|
| File components | `/files/{file_key}/components` |
| Team components | `/teams/{id}/components` |
| Component metadata (name, description, key) | Response body |

## CLI shape

```
figma components --team <team-url> [flags]

Flags:
  --codebase    Path to component source directory (e.g., src/components)
  --diff        Compare Figma components against codebase
  --drift       Only show components that differ from last snapshot
  --snapshot    Save current Fugma component list as a snapshot file
  --format      Output format: table, json (default: table)
```

## Output example

```
$ figma components --team wire --codebase src/components --diff

Components: 47 in Figma, 41 in code

Matched (39):
  Button              src/components/Button
  Card                src/components/Card
  Input               src/components/Input
  ...

Missing in code (6):
  AvatarGroup         (not found in src/components)
  Breadcrumb          (not found in src/components)
  DateRangePicker     (not found in src/components)
  ...

Extra in code (2):
  LegacyButton        (not in Figma library — candidate for removal)
  OldModal             (not in Figma library — candidate for removal)

Drifted (3):
  Button              last synced 2026-06-15, Figma updated 2026-07-07
  Table               last synced 2026-05-20, Figma updated 2026-07-01
  Toast               last synced 2026-04-10, Figma updated 2026-07-05
```

## How name matching works

Figma component name: `Button / Primary / Large`
Code component path: `src/components/Button/Button.tsx`

Match rule:
1. Extract the base name: `Button / Primary / Large` → `Button`
2. Convert to fileNameCase: `Button` → `button`
3. Search `src/components/` for `button.*` (case-insensitive)
4. Also try the full path: `button/primary/large` → `src/components/button/primary/large.*`

## Snapshot file format

```json
{
  "syncedAt": "2026-07-07T10:00:00Z",
  "components": {
    "Button": { "key": "abc123", "description": "Primary button component" },
    "Card": { "key": "def456", "description": "Content card" }
  }
}
```

A snapshot lets us answer "what changed?" without fetching the whole file every time.

## Edge cases

- Component names with `/` separators (variants) → flatten for matching
- Components that are part of a component set → show set name as prefix
- Team library vs file library → team flag uses different endpoint
- Multiple codebase directories → accept multiple `--codebase` flags
- Non-React codebases → file matching should be configurable (`--name-pattern`)
