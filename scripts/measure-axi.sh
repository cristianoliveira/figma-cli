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
go build -o "$work/bin/pixel-perfect" ./cmd/pixel-perfect

fixture="tests/smoke/fixtures/image-diff/reference.png"
changed="tests/smoke/fixtures/image-diff/two-regions.png"
large="tests/smoke/fixtures/image-diff/real-ui-reference.png"

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
measure pixel-identical-toon 1 0 "$work/bin/pixel-perfect" "$fixture" "$fixture" --output "$artifact/identical-toon.png"
measure pixel-identical-json 1 0 "$work/bin/pixel-perfect" --json "$fixture" "$fixture" --output "$artifact/identical-json.png"
measure pixel-failed-gate-toon 1 0 "$work/bin/pixel-perfect" "$fixture" "$changed" --output "$artifact/gate-toon.png" --max-changed-ratio 0
measure pixel-failed-gate-json 1 0 "$work/bin/pixel-perfect" --json "$fixture" "$changed" --output "$artifact/gate-json.png" --max-changed-ratio 0
measure pixel-probe-truncated-csv 1 0 "$work/bin/pixel-perfect" probe "$large" "$large" --at 100,100 --radius 10
measure pixel-probe-truncated-json 1 0 "$work/bin/pixel-perfect" probe "$large" "$large" --at 100,100 --radius 10 --format json
measure pixel-scan-csv 1 0 "$work/bin/pixel-perfect" scan "$fixture" "$changed" --row 0
measure pixel-scan-json 1 0 "$work/bin/pixel-perfect" scan "$fixture" "$changed" --row 0 --format json
