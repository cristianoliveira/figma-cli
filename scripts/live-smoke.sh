#!/usr/bin/env bash
# Runs read-only Figma smoke checks. It never prints FIGMA_ACCESS_TOKEN.
set -euo pipefail

for variable in FIGMA_ACCESS_TOKEN FIGMA_SMOKE_FILE_URL FIGMA_SMOKE_NODE_URL; do
  if [[ -z "${!variable:-}" ]]; then
    printf 'live smoke configuration error: set %s before running\n' "$variable" >&2
    exit 2
  fi
done

figma_bin="${FIGMA_BIN:-figma}"
if ! command -v "$figma_bin" >/dev/null 2>&1; then
  printf 'live smoke configuration error: figma binary not found; set FIGMA_BIN or add figma to PATH\n' >&2
  exit 2
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

run_json() {
  local name="$1"
  shift
  local stdout="$workdir/$name.stdout"
  local stderr="$workdir/$name.stderr"
  if ! "$figma_bin" "$@" >"$stdout" 2>"$stderr"; then
    printf 'live smoke failed: %s; check token permissions, fixture access, and Figma availability\n' "$name" >&2
    exit 1
  fi
  if ! jq -e . "$stdout" >/dev/null; then
    printf 'live smoke failed: %s returned invalid JSON; inspect the command contract\n' "$name" >&2
    exit 1
  fi
  printf 'ok: %s\n' "$name"
}

run_json me me
jq -e '(.id // .result.id // .handle // .result.handle) | strings | length > 0' "$workdir/me.stdout" >/dev/null || {
  printf 'live smoke failed: me response has no identity\n' >&2; exit 1;
}

run_json meta meta "$FIGMA_SMOKE_FILE_URL"
run_json inspect inspect "$FIGMA_SMOKE_NODE_URL"
jq -e '.scope.fileKey and .result.id and .result.name and .result.type' "$workdir/inspect.stdout" >/dev/null || {
  printf 'live smoke failed: inspect response misses required scope or node fields\n' >&2; exit 1;
}

find_name="${FIGMA_SMOKE_FIND_NAME:-Button}"
run_json find find --name "$find_name" --limit 1 "$FIGMA_SMOKE_FILE_URL"
jq -e '.query.name and (.total | numbers) and (.results | arrays)' "$workdir/find.stdout" >/dev/null || {
  printf 'live smoke failed: find response misses query, total, or results\n' >&2; exit 1;
}

missing_name="__figma_cli_live_smoke_missing__"
run_json find-empty find --name "$missing_name" "$FIGMA_SMOKE_FILE_URL"
jq -e --arg name "$missing_name" '.query.name == $name and .total == 0 and .results == []' "$workdir/find-empty.stdout" >/dev/null || {
  printf 'live smoke failed: empty find response contract mismatch\n' >&2; exit 1;
}

run_json layout layout --depth 1 "$FIGMA_SMOKE_NODE_URL"
jq -e '.query.maxDepth == 1 and (.traversal.returnedNodes | numbers) and (.traversal.totalNodes | numbers) and (.traversal.truncated | booleans) and (.result.id | strings)' "$workdir/layout.stdout" >/dev/null || {
  printf 'live smoke failed: layout traversal contract mismatch\n' >&2; exit 1;
}

printf 'live smoke passed\n'
