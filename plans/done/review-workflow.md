# Design review workflow

> **Status (10-07-26): Implemented.** Comments emit ordered threads with state, author, date, node path, and direct URL. Node-scoped ancestor lookup uses scoped file requests and works on large files. Text diff and blame include readable parent paths; blame accepts explicit `--id` or URL scope.

## Problem

`comments`, text diffs, and blame expose useful data but not enough context for frontend review. Comments are flat rather than threaded; copy changes lack frame paths; blame only accepts URL node scope. Developers still return to Figma to understand where and why change happened.

## Solution

Shape collaboration commands around review tasks: grouped threads, explicit state, author/time ordering, node paths, and consistent node scope.

## How

- Group comments by thread/root and include replies, resolved state, author, timestamps, and node context.
- Add filters for open/resolved, author, node, and date range.
- Add parent/frame path to text diff and blame output.
- Add optional `--id` to blame while retaining URL inference.
- Follow `versions` precedent: readable human summary plus stable JSON.
- Add fixtures for replies, deleted nodes, resolved threads, and moved text nodes.

## Success criteria

- Developer can review active feedback without opening every Figma thread.
- Copy changes identify where node lives.
- Collaboration commands follow shared scope/output contracts.
