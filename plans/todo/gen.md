# gen (targeted code generation)

## Problem

Full "Figma to code" is a trap — generated code is unmaintainable, doesn't follow team
conventions, and creates a write-only artifact. But *targeted* code generation is useful
for specific, well-bounded cases:

- **Simple components** — buttons, inputs, badges, avatars. Static structure, few variants.
- **Design tokens** — already covered by `tokens` command
- **Page skeletons** — the structure of a page (sections, headings, placeholders) without
  the styling details
- **Storybook boilerplate** — the args table and story structure for a component

The goal isn't to replace developers writing code. It's to eliminate the mechanical
transcription of Figma specs into component boilerplate.

## Success criteria

- Generate a React/Vue/Svelte component skeleton from a Figma COMPONENT or COMPONENT_SET
- Output respects a configurable template (so teams can customize)
- Only generate structure — layout hierarchy, component properties as props, text as children
- Never generate pixel-level inline styles (use class names or CSS modules)
- Complement `tokens` and `specs` — the three together produce a working component

## Non-goals

- Full visual output (don't try to match pixel-perfect Figma rendering)
- Complex components (nested tables, charts, canvases)
- State management, event handlers, business logic
- Auto-layout → CSS conversion (that's `layout` command's job)

## API mapping

| Data | API endpoint |
|---|---|
| Component tree | `/files/{file_key}/nodes?ids=...` |
| Component properties | Component node data (variants, boolean, text props) |
| Style references | Resolved via `styles` map |

## CLI shape

```
figma gen <component-name> <file-url> [flags]

Flags:
  --framework     Target framework: react, vue, svelte, html (default: react)
  --typescript    Generate TypeScript (default: true)
  --css           CSS approach: modules, tailwind, styled, css (default: modules)
  --output        Write to file (default: stdout)
  --template      Custom template file path
  --page          Generate page structure (all sections on a page)
```

## Output example — React component

```
$ figma gen "Button / Primary" --framework react --typescript --css modules
```

```tsx
// Button.tsx
import styles from "./Button.module.css";

export interface ButtonProps {
  /** Variant of the button */
  variant: "Primary" | "Secondary" | "Tertiary";
  /** Whether the button is disabled */
  disabled?: boolean;
  /** Button label */
  children: React.ReactNode;
}

export function Button({
  variant = "Primary",
  disabled = false,
  children,
}: ButtonProps) {
  return (
    <button
      className={`${styles.button} ${styles[variant]}`}
      disabled={disabled}
      type="button"
    >
      {children}
    </button>
  );
}
```

```css
/* Button.module.css */
.button {
  /* Extracted from Button specs — see `figma specs Button` */
  /* Colors from `figma tokens` */
  composes: button-base from "../../tokens/button.css";
}

.primary {
  background: var(--color-primary);
}

.secondary {
  background: var(--color-secondary);
}
```

## Generation pipeline

```
figma tokens <file> --format css    → src/tokens/button.css
figma specs "Button" --format ts    → ButtonProps interface
figma gen "Button" --framework react → Button.tsx (uses both above)
```

## Edge cases

- Component with no variants → generate a single component, no variant prop
- Component properties that are INSTANCE_SWAP → generate as `React.ReactNode` or generic slot
- TEXT nodes → generate as children (text content)
- Nested component instances → generate as child components (import they)
- Component with zero text nodes (icon-only button) → still generate valid structure
- Page generation → walk page frames, generate section components, compose in a page layout
