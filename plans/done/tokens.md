# tokens

> **Status (09-07-26):** Implemented for the file path — three token sources
> (Variables, Styles, and document Scan), all three formats (css/tailwind/json),
> `--source`, `--mode`, `--prefix`, `--output`, `--scan-fallback`. Auto tries
> Variables → Styles (named tokens); the raw-fill document scan is opt-in via
> `--scan-fallback` or `--source scan` so output stays deterministic per source.
> Verified end-to-end against a real raw-fill file (0 styles, 403 variables).
> Deferred: `--team` (returns a clear not-supported error) and `--watch`.
> See `cmd/tokens.go`, `internal/extract/tokens.go`, `internal/figma/tokens.go`.

> **Follow-up (09-07-26):** `figma css` shipped — generates layout+fills+type
> CSS per node from the public file API (no Enterprise needed). Mirrors Dev
> Mode's CSS panel but scriptable/batchable. See `cmd/css.go`,
> `internal/extract/css.go`. Next candidate: `figma scaffold --react`.
>
> **Open finding:** Variables REST API is Enterprise-gated (403 here), so
> semantic token names (`--Base-Primary`) are unreachable without Enterprise;
> Dev Mode shows them via its private client. `tokens --source scan` falls
> back to value-named tokens. Dev Mode's CSS for some nodes differs from raw
> autolayout values (padding 43.716 vs 40) due to instance scale transforms;
> we emit the literal stored values.

## Problem

Frontend developers manually copy hex codes, font sizes, spacing values, shadows, and
border radii from Figma into CSS/Tailwind/design-token files. This is:

- **Error-prone** — a typo in a hex code (`#0667C8` vs `#0667cc`) is invisible until QA
- **Slow** — a medium design system has 80–200 tokens, manually transcribed per handoff
- **Drifts** — the Figma file updates, the codebase doesn't. Nobody notices for 2 sprints.
- **Undiscoverable** — tokens live in Figma styles, which devs rarely see. The result is
  hard-coded magic values spread across 50 files instead of centralized `--color-primary`.

## Success criteria

- A single command extracts all color/text/spacing/shadow tokens from a Figma file or team
  library and outputs them as ready-to-use CSS custom properties, Tailwind config,
  or JSON (for style-dictionary/Token Studio)
- Output is idempotent — running it twice produces the same file (CI-safe)
- The command surfaces *where* a token is used, not just its name (naming aliases)

## API mapping

| Figma concept | API endpoint | Our mapping |
|---|---|---|
| Color styles | `/files/{key}/styles` + `/styles/{key}` | `--format css:--color-` |
| Text styles | same | `--format css:--font-` |
| Effect styles (shadows, blurs) | same | `--format css:--shadow-` |
| Variables (Figma's token system) | `/files/{key}/variables/local` + `/files/{key}/variables/published` | `--format json` (native variables) |
| Team library tokens | `/teams/{id}/styles` | `--team` flag |

## CLI shape

```
figma tokens <file-url> [flags]

Flags:
  --format     Output format: css, tailwind, json (default: css)
  --team       Pull from team library instead of a specific file
  --prefix     CSS custom property prefix (default: "")
  --output     Write to file instead of stdout
  --watch      (future) Poll for changes, regenerate on update
```

## Output examples

### CSS
```css
:root {
  --color-primary: #0667C8;
  --color-background: #FFFFFF;
  --color-text-primary: #18181A;
  --color-text-muted: #696C6E;
  --font-body-family: "Inter";
  --font-body-size: 14px;
  --font-body-weight: 400;
  --shadow-card: 0px 2px 8px rgba(0, 0, 0, 0.08);
}
```

### Tailwind
```js
module.exports = {
  theme: {
    extend: {
      colors: {
        primary: "#0667C8",
        background: "#FFFFFF",
        "text-primary": "#18181A",
        "text-muted": "#696C6E",
      },
      fontFamily: { body: ["Inter"] },
      fontSize: { body: "14px" },
    },
  },
};
```

### JSON
```json
{
  "color": {
    "primary": { "value": "#0667C8", "type": "color" },
    "background": { "value": "#FFFFFF", "type": "color" }
  }
}
```

## Edge cases to handle

- Styles with no name (auto-generated IDs) → skip or warn
- Nested style groups in Figma → flatten with `/` separator (e.g., `color/primary/500`)
- Duplicate style names → append a discriminator or warn
- Text styles with multiple fonts → use first, note others in a comment
- Gradient fills → output as `linear-gradient(...)` string (not usable as a color variable)
- Missing API access (team library needs `file_content:read` scope) → clear error message
