# inspect is missing computed gaps — making pixel-perfect workflows needlessly iterative

**Date:** 2026-07-12
**Context:** Building a Drive card pixel-perfect against a Figma frame. Used `figma inspect --recursive`, `figma css --recursive`, `figma layout`, `figma export`, `pixel-perfect`, and Playwright.

---

## The experience

When you're doing pixel-perfect, `figma inspect --recursive` is your canonical source. It's the dump you compare DOM bounding boxes against. Every child node has bounds, layout mode, typography, padding. You trust it.

But there's one number that's missing: **the measured gap between adjacent siblings**.

I had a dropdown list (13576:15266, 511x165, vertical auto-layout, `primaryAxisAlignItems: CENTER`). Two children: a sales text (413x48) and a "Contact sales" link (107x24). The inspect tells me all their bounds. But it doesn't tell me that the gap between them is zero.

To figure it out I have to manually subtract:
```
301.718 - (253.718 + 48) = 0
```
That's fine once. Across 15+ nodes in a frame, when you're iterating on CSS and recapturing screenshots, you skip it. You assume the flexbox `gap` on the parent covers it. But the parent gap and the sibling-to-sibling gap aren't always the same thing—especially when containers nest.

## What happened

I spent 5+ image comparison iterations chasing a vertical shift of the blue callout box. `pixel-perfect --suggest-offset` kept saying `y: -8`. I was adjusting card padding, header margins, description height. The actual root cause was simpler: the sales text rendered at 1 line in the browser (24px) instead of Figma's 2 lines (48px), which collapsed the vertical space, which shifted the contact link up by 24px, which shifted the whole dropdown group up, which threw off the chevron position.

If `inspect --recursive` had told me `computedGap: 0` between the sales text and the contact link, I'd have noticed the height mismatch immediately: "This gap is zero, so for the contact link to be at y=301.718, the text above it must be exactly 48px tall. Let me check the browser height."

## The gap

`figma layout --measure-spacing` exists. It produces exactly this data. But it has two problems in practice:

1. **It's a separate command.** When I'm deep in a pixel-perfect loop with `inspect --recursive` as my reference, I don't think to run a second structure command. The mental model is "inspect = source of truth, CSS = implementation output, pixel-perfect = validation."

2. **The output is a tree, not per-node metadata.** The gaps are embedded in a visual tree format. I can't pipe it into `jq` and extract a list of all zero-gap sibling pairs in one line. If `inspect --recursive` included a `gapToNextSibling` (or `gapToPreviousSibling`) field on every non-last child, it would be one `jq '.. | select(.gapToNextSibling == 0)'` away.

## What would have fixed this

Either:

**A.** Add computed gaps to `inspect --recursive` output. Every auto-layout child gets a `gapBelow` (or `gapToNext`) field with the measured pixel gap to the next sibling. Last child gets `null`. This makes the inspect dump truly self-contained.

**B.** Or add a `--json` flag to `figma layout --measure-spacing` so the measured gaps are machine-readable.

Option A is the real fix. `inspect --recursive` should be the one command you need to compare against a DOM render. Right now it's 95% there—it just drops the gap numbers that hold the layout together.
