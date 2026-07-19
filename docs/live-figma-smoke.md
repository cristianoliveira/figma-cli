# Live Figma smoke

Run this read-only check only with a dedicated least-privilege Figma token and a
maintainer-controlled stable fixture:

```bash
FIGMA_ACCESS_TOKEN=... \
FIGMA_SMOKE_FILE_URL='https://www.figma.com/design/...' \
FIGMA_SMOKE_NODE_URL='https://www.figma.com/design/...?...' \
scripts/live-smoke.sh
```

It validates `me`, `meta`, scoped `inspect`, bounded and empty `find`, and
bounded `layout`. When `FIGMA_SMOKE_TEAM_ID` or `FIGMA_SMOKE_PROJECT_ID` is
set, it also validates read-only project and file discovery; otherwise it prints
an explicit skip. It stores command output only in a temporary directory and
never prints the token or headers. It is deliberately excluded from `go test`.

GitHub Actions exposes this only as a manual dispatch in the protected
`figma-live-smoke` environment. Configure its `FIGMA_ACCESS_TOKEN` secret and
fixture URL variables there. Fixture owners rotate the token and update URLs.
For 401 rotate the token; 403 means missing access; 404 means a removed or
inaccessible fixture; 429 should be retried later; a JSON/schema failure is a
CLI contract investigation.
