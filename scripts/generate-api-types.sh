#!/usr/bin/env bash
# Generate Go types from the Figma REST API OpenAPI spec.
#
# Pipeline:
#   1. Down-convert OpenAPI 3.1 → 3.0 (oapi-codegen stable only handles 3.0 fully)
#   2. Fix leftover `type: 'null'` patterns from the 3.1→3.0 conversion
#   3. Generate Go types via oapi-codegen
#
# Prerequisites:
#   - Node.js (for @apiture/openapi-down-convert)
#   - Go (for fix-null-types.go)
#   - oapi-codegen (go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SPEC_DIR="$REPO_ROOT/openapi"
SCRIPTS_DIR="$REPO_ROOT/scripts"
SPEC_ORIG="$SPEC_DIR/openapi.yaml"
SPEC_30="$SPEC_DIR/openapi-3.0.yaml"
SPEC_PATCHED="$SPEC_DIR/openapi-3.0-patched.yaml"
GEN_OUT="$REPO_ROOT/internal/figma/api/api.gen.go"
CONFIG="$SPEC_DIR/oapi-codegen.yaml"

echo "→ Converting OpenAPI 3.1 → 3.0..."
npx --yes @apiture/openapi-down-convert -i "$SPEC_ORIG" -o "$SPEC_30"

echo "→ Patching null types..."
go run "$SCRIPTS_DIR/fix-null-types.go" "$SPEC_30" > "$SPEC_PATCHED"

echo "→ Generating Go types..."
mkdir -p "$(dirname "$GEN_OUT")"
oapi-codegen -config "$CONFIG" "$SPEC_PATCHED"

echo "→ Cleaning up temp files..."
rm "$SPEC_30" "$SPEC_PATCHED"

echo "✓ Generated $GEN_OUT"
