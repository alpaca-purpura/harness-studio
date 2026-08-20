#!/usr/bin/env bash
# Builds the Next.js UI as a static export for embedding in the Go binary.
# API routes can't exist under output:'export' — they're moved aside during
# the build (the Go server implements them).
set -euo pipefail

COCKPIT_DIR="$(cd "$(dirname "$0")/../ui" && pwd)"
GO_DIR="$(cd "$(dirname "$0")" && pwd)"
API_DIR="$COCKPIT_DIR/app/api"
API_STASH="$COCKPIT_DIR/.api-stash"

cleanup() {
  if [ -d "$API_STASH" ]; then
    rm -rf "$API_DIR"
    mv "$API_STASH" "$API_DIR"
  fi
}
trap cleanup EXIT

mv "$API_DIR" "$API_STASH"

cd "$COCKPIT_DIR"
rm -rf .next out
COCKPIT_STATIC=1 pnpm exec next build

rm -rf "$GO_DIR/ui"
cp -r out "$GO_DIR/ui"
echo "✓ static UI exported to $GO_DIR/ui ($(du -sh "$GO_DIR/ui" | cut -f1))"
