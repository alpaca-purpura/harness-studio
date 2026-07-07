#!/usr/bin/env bash
# Bundle end-to-end de ArnesIA (HS-11, «instalable en cualquier computadora»):
#   SPA (vite) → daemon Go (con SPA + doctrina + kit EMBEBIDOS) → sidecar Tauri →
#   instaladores de escritorio (.deb / .AppImage / lo que el target soporte).
#
# Prereqs de máquina de build (NO del usuario final): go · pnpm · rust/cargo ·
# deps webkit de Tauri 2 (Linux: libwebkit2gtk-4.1-dev …, ver ci.yml job rust).
# El usuario final solo necesita Claude Code instalado y logueado — todo lo demás
# viaja dentro del paquete y se provisiona solo (~/.arnesia, cuerpos ①+②).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TRIPLE="$(rustc -vV | sed -n 's/^host: //p')"
SIDECAR="$ROOT/web/src-tauri/binaries/arnesia-daemon-$TRIPLE"

echo "── 1/3 SPA (vite build → web/dist, la embebe el daemon)"
cd "$ROOT/web"
pnpm install --frozen-lockfile
pnpm run build

echo "── 2/3 daemon (go build con SPA+doctrina+kit dentro) → sidecar $TRIPLE"
cd "$ROOT"
mkdir -p "$(dirname "$SIDECAR")"
go build -trimpath -ldflags '-s -w' -o "$SIDECAR" ./cmd/arnesia

echo "── 3/3 Tauri bundle (instaladores en web/src-tauri/target/release/bundle/)"
cd "$ROOT/web"
pnpm exec tauri build

echo "OK — instaladores en web/src-tauri/target/release/bundle/"
