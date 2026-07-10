# team (organization dashboard)

> **Status (09-07-26):** Deferred / re-scoped. The discovery primitives it
> depended on now exist: `figma me`, `figma projects`, `figma files` (see
> `plans/done/project-files.md`). **Decision: do not build this as the
> standalone monolith described below.** Build `team` as a thin composition
> layer over `projects` + `files` (one REST call each, already shipped), or
> skip a dedicated command and compose in shell with `--json`. The
> `--list/--stats/--activity/--components/--find` flags below are a wishlist for
> whenever someone does build the dashboard; they are not the next step.

## Problem

In a medium-to-large org, Figma files multiply fast. The questions that arise:

- **"Where is the design for X?"** — which project/file contains the Checkout flow?
- **"Is this file active?"** — last modified 3 months ago; is it abandoned or just stable?
- **"Who owns this?"** — no clear owner, nobody updates it
- **"How big is our design library?"** — how many components, styles, variables across teams?
- **"What files reference this component?"** — if I change this button component, what
  impacts?

There's no dashboard for this. The Figma web app shows a file browser but no org-level view,
no activity feed, and no cross-file dependency graph. Teams manage this through conventions
or (more commonly) Slack DMs asking "do you know where the file for X is?"

## Success criteria

- List all projects and files in a team
- Show file metadata: last modified, owner, thumbnail URL, version count
- Cross-reference components: which files use a specific component?
- Activity feed: recently modified files, recently commented files
- Library health: component count, style count, variable count
- Output is both human-readable (table) and machine-readable (JSON)

## API mapping

| Data | API endpoint |
|---|---|
| Team projects | `/teams/{team_id}/projects` |
| Project files | `/projects/{project_id}/files` |
| Team components | `/teams/{team_id}/components` |
| Team styles | `/teams/{team_id}/styles` |
| Activity logs | `/activity_logs` |
| File metadata | `/files/{file_key}` (metadata only) |

## CLI shape

```
figma team <team-id-or-url> [flags]

Flags:
  --list         List all projects and files (default action)
  --stats        Show aggregate statistics
  --activity     Show recent activity (last N days)
  --components   Show component library overview
  --find         Find files by name pattern within the team
  --format       Output format: table, json (default: table)
  --since        Filter activity/modifications since date
```

### Resolving team ID

```
# From a project URL
figma team https://www.figma.com/files/team/123456789/Wire

# From an explicit ID
figma team 123456789

# From a file URL (look up the team)
figma team --file https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-
```

## Output examples

### Team overview
```
$ figma team wire --list

Wire (team 123456789)
  Design System (project)
    ├─ Primitives · last modified 2026-07-07 by Olga Skoczylas
    ├─ Components · last modified 2026-07-06 by Olaf Sulich
    └─ Icons · last modified 2026-06-30 by François Benaiteau

  Product (project)
    ├─ Drive (Cells) · last modified 2026-07-07 by Olga Skoczylas
    ├─ Conversations · last modified 2026-07-01 by Mathias Nibouliès
    └─ Settings · last modified 2026-06-15 by Sarah

  Archive (project)
    └─ Legacy Design System · last modified 2025-11-20 by Olga Skoczylas
```

### Team stats
```
$ figma team wire --stats

Wire · 3 projects · 7 files

Components: 47 · Styles: 28 · Variables: 12

Activity (last 30 days):
  7 files modified
  142 comments
  3 new versions
```

### Find files
```
$ figma team wire --find "checkout"

2 matches for "checkout":
  Product / Checkout Flow · last modified 2026-06-20
  Archive / Old Checkout · last modified 2025-08-15
```

## Edge cases

- Large teams with 50+ projects → paginate, accept `--filter` to narrow
- Team ID vs project ID ambiguity → detect from URL structure
- Personal files (not in a team) → `--team` flag won't work; use `--user` instead
- Rate limiting on activity logs → cache results, respect `Retry-After` headers
- Deleted/archived files → mark as [archived] in output
