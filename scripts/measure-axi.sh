#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
artifact=".tmp/axi-measure"
work="$root/$artifact"
trap 'rm -rf "$work"' EXIT
rm -rf "$work"
mkdir -p "$work/bin"

go build -o "$work/bin/figma" ./cmd/figma

printf 'scenario\tstdout_bytes\tstderr_bytes\texit\tcommands\trecovery_round_trips\n'
measure() {
  local scenario="$1" commands="$2" recovery="$3"
  shift 3
  local stdout="$work/stdout" stderr="$work/stderr" status
  set +e
  "$@" >"$stdout" 2>"$stderr"
  status=$?
  set -e
  printf '%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$scenario" "$(wc -c <"$stdout" | tr -d ' ')" "$(wc -c <"$stderr" | tr -d ' ')" "$status" "$commands" "$recovery"
}

measure figma-discovery 1 0 env -u FIGMA_ACCESS_TOKEN "$work/bin/figma"
measure figma-invalid-toon 1 1 "$work/bin/figma" inspec
measure figma-invalid-json 1 1 "$work/bin/figma" --json inspec
