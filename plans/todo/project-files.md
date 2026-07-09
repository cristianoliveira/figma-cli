# projects & files (discovery primitives)

> **Status (09-07-26):** Implemented locally; pending live URL validation and
> commit. `figma me` (the first discovery slice) is implemented in `410040c`.
> `figma projects` and `figma files` now exist as thin, single-endpoint
> **primitives** — composable via `--json` — not a dashboard. See
> "Reconciliation with team.md".

## Problem

Today every command requires a Figma file key or full URL up front. There is no
way to **discover** what exists:

- *"Which project holds the Checkout design?"* — you must already know the file URL.
- *"What files are in this project?"* — no command answers it; you go to the web app.
- *"Is my token even valid before I run a 20-command script?"* — `me` now answers this.

The navigation hierarchy is simple — **team → projects → files** — and each hop
is one REST call. The CLI should expose each hop as its own noun command so a
user (or a shell script) can walk the tree.

## Reconciliation with `plans/todo/team.md`

`team.md` proposes a single monolithic `figma team <id>` command with
`--list / --stats / --activity / --components / --find` flags — an org dashboard.

**Decision: build the flat nouns first, defer the dashboard.**
- `projects` and `files` are one endpoint each → trivial, testable, shippable now.
- The `team` dashboard is a **composition** over `projects` + `files` + stats +
  activity. With `--json`, it is scriptable *today* (`projects --json` → loop
  `files --json`). Building the monolith first inverts the dependency and bakes
  in assumptions before the primitives exist.
- Recommendation: implement `projects`/`files` here, then revisit `team.md` as a
  thin formatting layer that calls them. Do **not** build `team`'s stats/activity
  until the primitives prove out.

This mirrors the house rule: *prefer noun commands with flags over positional
filter subcommands*, and the earlier "scripting + `--json`" insight.

## Success criteria

- `figma projects [team-url|team-id]` lists a team's projects, defaulting to `FIGMA_TEAM_ID`
- `figma files <project-url|project-id>` lists a project's files
- Both accept a bare ID **or** a Figma URL (consistent with `colors`, `versions`)
- Both emit curated camelCase JSON and honor `--json`
- Both flow through `RunE` (no `cli.Die` — it was removed)
- Idempotent: same input → same output (CI-safe)

## API mapping (verified against `openapi/openapi.yaml` + `api.gen.go`)

| Command | Endpoint | Params | Response type | Pagination |
|---|---|---|---|---|
| `projects` | `GET /v1/teams/{team_id}/projects` | `team_id` (path) | `GetTeamProjectsResponse{Name, []Project{Id,Name}}` | **none** |
| `files` | `GET /v1/projects/{project_id}/files` | `project_id` (path), `branch_data` (query, bool, default false) | `GetProjectFilesResponse{Name, Files[]{Key,Name,LastModified,ThumbnailUrl*}}` | **none** |
| *(later)* `project` | `GET /v1/projects/{project_id}/meta` | `project_id` | `GetProjectMetaResponse{CreatedAt,FileCount,Id,Name,...}` | none |

**Correction to earlier assumption:** I previously flagged pagination as "the
real work." The spec shows **no cursor/page_size on either endpoint** — both
return the full list. YAGNI: skip `--cursor`/`--limit` unless real-world use
proves otherwise (the `team.md` "50+ projects" edge case is speculative).

**Hard UX constraint (from the spec):** *"it is not currently possible to
programmatically obtain the team id of a user just from a token."* So users
**must** supply the team URL/ID — `me` will not return it. The command help and
errors must say this plainly rather than fail silently.

## CLI shape

```
figma projects [team-url|team-id] # argument or FIGMA_TEAM_ID
figma files    <project-url|project-id> [--branches]
```

- `--branches` (files only) → sets `branch_data=true`, surfaces branch metadata.
  Default off, to keep output stable.
- No positional filter; no `--cursor` (see pagination correction above).
- Both inherit `--json` from root.

### Resolving IDs

```
# team from a team URL
figma projects https://www.figma.com/files/team/123456789/Wire
# team from a bare ID
figma projects 123456789

# project from a project URL
figma files https://www.figma.com/files/project/987654321/Design-System
# project from a bare ID
figma files 987654321
```

## The real boundary work: URL parsing

`figma.ParseInput` today only understands **file** keys/URLs. Both new commands
need team/project URL parsing — a new boundary helper, parallel to `ParseInput`.

**Known unknown (verify before implementing):** the exact Figma URL shapes for
team and project pages. Likely forms (need confirmation against a live URL):
- team:    `figma.com/files/team/<TEAM_ID>/<slug>`  or  `figma.com/team/<TEAM_ID>/...`
- project: `figma.com/files/project/<PROJECT_ID>/<slug>` or `figma.com/project/<PROJECT_ID>/...`

Task 0 remains a release check: validate one real team URL and one real project
URL. The parser currently pins both documented candidate forms above and fails
loudly for other hosts/path kinds. Bare IDs (all-digits) are accepted
unconditionally. Ambiguous IDs → prefer
treating as the type the command expects (`projects` arg = team id, `files`
arg = project id), so no cross-command confusion.

## Output shapes (curated, camelCase — matches `me`)

```jsonc
// figma projects <team>
{
  "team": "Wire",
  "projects": [ { "id": "123", "name": "Design System" } ]
}

// figma files <project>
{
  "project": "Design System",
  "files": [
    { "key": "grnV...", "name": "Primitives", "lastModified": "2026-07-07T...", "thumbnailUrl": "https://..." }
  ]
}
```

Curate at the `cmd` layer into small structs (`projectsOutput`, `filesOutput`),
mapping `last_modified`→`lastModified`, `thumbnail_url`→`thumbnailUrl`. Keeps
the output contract decoupled from `api.gen.go` renames (same rationale as
`meOutput`).

## Implementation / file layout (mirror the `me` slice)

```
internal/figma/urls.go        BuildTeamProjectsURL(teamID) / BuildProjectFilesURL(projectID, branchData)
internal/figma/urls_test.go   string-equality tests (mirror TestBuildVersionsURL)
internal/figma/projects.go    FetchTeamProjects(client, apiURL) (api.GetTeamProjectsResponse, error)
internal/figma/files.go       FetchProjectFiles(client, apiURL)   (api.GetProjectFilesResponse, error)
                              ^ apiURL passed in → httptest-injectable (FetchExportURL seam)
internal/figma/input.go       ParseTeamInput / ParseProjectInput (or extend ParseInput) — URL parsing
internal/figma/input_test.go  URL + bare-ID cases
cmd/projects.go               thin RunE → LoadClient + Fetch + curate + NewPrinter(cmd).JSON
cmd/files.go                  thin RunE (+ --branches flag)
```

The fetch functions take the **full apiURL** as a parameter (not calling
`Build*URL` internally) so tests point them at an httptest server — this is the
established seam (`FetchExportURL`, `FetchMe`). Commands compose
`Build*URL` → `Fetch*`.

## TDD order

1. `BuildTeamProjectsURL` / `BuildProjectFilesURL` — string tests (red → impl)
2. `FetchTeamProjects` / `FetchProjectFiles` — httptest + testify (happy + 403)
3. `ParseTeamInput` / `ParseProjectInput` — URL + bare-ID + invalid cases
4. `cmd/projects.go`, `cmd/files.go` — thin, follow `inspect.go` RunE shape
5. `go fmt`, `golangci-lint`, `go test -race ./...`, build, `--help` smoke

## Edge cases

- **Team ID unobtainable from token** → help text + error must state this; do not
  imply `me` can provide it.
- **Bare ID vs URL ambiguity** → resolve by command context (`projects` arg is a
  team id, `files` arg is a project id).
- **Empty team / empty project** → return `{team/project, []}` not an error.
- **`branch_data`** → opt-in flag; default output must not change shape.
- **403 (token lacks `projects:read`)** → flows through `RunE` → `error: ...`, exit 1.
- **URL format drift** → pin parser to observed Figma URLs (Task 0); fail loudly
  on unrecognised shapes rather than guessing.

## Ordering / next steps

1. ✅ `figma me` — done (`410040c`)
2. ✅ `figma projects` — implemented locally
3. ✅ `figma files` — implemented locally
4. ♻️ Revisit `team.md` — build the dashboard as a composition layer over
   `projects`/`files`, not a standalone monolith.
5. Later discovery buckets (separate plans): dev-resources, library-usage
   analytics, component-sets.
