# Opt-in live Figma smoke tests

## Problem

Local tests use injected clients and test servers, which correctly keep CI deterministic but cannot detect live Figma permission, endpoint, or response-shape changes. Running ad hoc commands with a personal token is not repeatable and risks leaking credentials in logs.

## Proposed outcome

Provide an explicit, read-only live smoke command and optional manual/scheduled CI workflow. Normal pull-request tests remain credential-free and deterministic.

## Safety contract

- never run on untrusted fork pull requests;
- never print token or request headers;
- use read-only endpoints and commands;
- make no Figma mutations;
- use dedicated least-privilege token and stable smoke file;
- skip with explicit reason when optional scope fixtures are absent;
- operational failures exit 1 with one recovery hint;
- live smoke is never hidden inside `go test ./...`.

## Required configuration

```text
FIGMA_ACCESS_TOKEN       required secret
FIGMA_SMOKE_FILE_URL     required stable read-only file
FIGMA_SMOKE_NODE_URL     required stable selected frame/node
FIGMA_SMOKE_TEAM_ID      optional workspace discovery scope
FIGMA_SMOKE_PROJECT_ID   optional project file-list scope
```

Fixture ownership, expected permissions, and rotation process must be documented. Use a file controlled by project maintainers, not a personal mutable design.

## Proposed smoke sequence

1. `figma me`
   - valid JSON;
   - non-empty ID/handle;
   - no token/header leakage.
2. `figma meta "$FIGMA_SMOKE_FILE_URL"`
   - valid JSON and expected file identity field.
3. `figma inspect --id/URL-node "$FIGMA_SMOKE_NODE_URL"`
   - scope contains configured file/node;
   - result has ID, name, type.
4. `figma find --name <known-term> --limit 1 "$FIGMA_SMOKE_FILE_URL"`
   - query metadata preserved;
   - total/results/truncation types valid.
5. `figma find --name <guaranteed-missing-sentinel> "$FIGMA_SMOKE_FILE_URL"`
   - explicit empty array with exact query context.
6. `figma layout --depth 1 "$FIGMA_SMOKE_NODE_URL"` after Task 02
   - traversal metadata valid and bounded.
7. Optional team/project commands only when IDs and token scopes are configured.

Do not assert volatile timestamps, thumbnail URLs, complete node payloads, or exact version counts.

## Delivery slices

### Slice 1 — local runner

Create one script/command that:

- validates required environment before network;
- creates private temporary directory;
- captures stdout and stderr separately;
- validates JSON with `jq` or typed helper;
- prints concise per-step status;
- cleans temporary artifacts;
- never enables shell tracing around secrets.

### Slice 2 — test runner itself without network

Test environment validation, redaction, JSON assertions, skips, and failure summaries using fake binaries or injected command executor. Same inputs must produce same output.

### Slice 3 — manual CI workflow

Add `workflow_dispatch` with protected secrets/environment. Consider scheduled run only after manual stability is proven. Upload sanitized result summary, never raw headers/config.

### Slice 4 — failure ownership

Document:

- token owner/rotation;
- smoke file owner;
- permission scopes;
- expected response to 401, 403, 404, 429, and schema assertion failures;
- how to reproduce locally.

## Test matrix

| Failure | Expected diagnosis |
|---|---|
| token missing | configuration error before network |
| 401 | token invalid/expired; rotate secret |
| 403 | exact missing Figma scope/permission |
| 404 | fixture file/node removed or inaccessible |
| 429 | rate limited; retry later, no tight loop |
| invalid JSON | endpoint/upstream response regression |
| schema mismatch | command mapping contract regression |
| optional team ID absent | explicit skip, overall core smoke may pass |

## Acceptance criteria

- One documented local command runs core live smoke.
- Manual CI workflow uses protected secrets and cannot run with fork secrets.
- Every operation is read-only.
- Token never appears in stdout, stderr, artifacts, or failure messages.
- Core smoke validates success, empty result, bounded result, and auth paths.
- Ordinary `go test -count=1 ./...` remains offline and deterministic.

## Risks

- Mutable Figma fixtures create false alarms.
- Scheduled workflow can consume API quota.
- Personal tokens create ownership and rotation problems.
- Raw response artifacts may contain user email or design data; sanitize or avoid upload.

## Implementation freedom

Shell, Go, or a small dedicated command is acceptable. Prefer the smallest design that gives redaction, typed assertions, and actionable failures.
