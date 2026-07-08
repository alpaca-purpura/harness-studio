#!/usr/bin/env bash
# Bundle end-to-end de ArnesIA (HS-11, «instalable en cualquier computadora»):
#   SPA (vite) → daemon Go (con SPA + doctrina + kit EMBEBIDOS, → bin/arnesia) →
#   sidecar Tauri → instaladores de escritorio (.deb / .AppImage / lo que el target soporte).
#
# Modo `--daemon-only` (paquete boton-actualizar, RF-104 ②): SOLO pasos 1–2 —
# SPA + go build → bin/arnesia. Es lo que corre el self-update del daemon; rustc/Tauri
# NO son prereq en ese camino.
#
# Prereqs de máquina de build (NO del usuario final): go · pnpm · [solo bundle completo:
# rust/cargo + deps webkit de Tauri 2 (Linux: libwebkit2gtk-4.1-dev …, ver ci.yml job rust)].
# El usuario final solo necesita Claude Code instalado y logueado — todo lo demás
# viaja dentro del paquete y se provisiona solo (~/.arnesia, cuerpos ①+②).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DAEMON_ONLY=0
[[ "${1:-}" == "--daemon-only" ]] && DAEMON_ONLY=1
STEPS=$((3 - DAEMON_ONLY))

echo "── 1/$STEPS SPA (vite build → web/dist, la embebe el daemon)"
cd "$ROOT/web"
pnpm install --frozen-lockfile
pnpm run build

echo "── 2/$STEPS daemon (go build con SPA+doctrina+kit dentro) → bin/arnesia"
cd "$ROOT"
mkdir -p "$ROOT/bin"
go build -trimpath -ldflags '-s -w' -o "$ROOT/bin/arnesia" ./cmd/arnesia

if [[ "$DAEMON_ONLY" == 1 ]]; then
  echo "OK — daemon en bin/arnesia (--daemon-only: sin sidecar ni instaladores)"
  exit 0
fi

# El sidecar Tauri es el MISMO binario; el triple se calcula solo aquí (exige rustc).
TRIPLE="$(rustc -vV | sed -n 's/^host: //p')"
SIDECAR="$ROOT/web/src-tauri/binaries/arnesia-daemon-$TRIPLE"
mkdir -p "$(dirname "$SIDECAR")"
cp "$ROOT/bin/arnesia" "$SIDECAR"

echo "── 3/$STEPS Tauri bundle (instaladores en web/src-tauri/target/release/bundle/)"
cd "$ROOT/web"
pnpm exec tauri build

echo "OK — instaladores en web/src-tauri/target/release/bundle/"
