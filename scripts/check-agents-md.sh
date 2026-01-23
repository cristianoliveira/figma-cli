#!/usr/bin/env bash
set -euo pipefail

# CI check for AGENTS.md line count
# Threshold: 160 lines maximum (allowing slight flexibility over 150)

AGENTS_MD="AGENTS.md"
THRESHOLD=160

echo "Checking $AGENTS_MD line count..."

# Count lines
LINE_COUNT=$(wc -l < "$AGENTS_MD")

echo "  Current line count: $LINE_COUNT"
echo "  Maximum allowed: $THRESHOLD"

if [[ $LINE_COUNT -gt $THRESHOLD ]]; then
    echo "❌ ERROR: $AGENTS_MD exceeds maximum line count ($LINE_COUNT > $THRESHOLD)"
    echo "   Please reduce the file size to stay under $THRESHOLD lines."
    echo "   Consider moving detailed instructions to separate documentation."
    exit 1
else
    echo "✅ SUCCESS: $AGENTS_MD line count is within limit."
fi