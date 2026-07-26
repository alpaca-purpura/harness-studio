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

# Identidad del build (RF-231): semver + sello de compilación, inyectados por -ldflags.
#
# Va ACÁ y no en el Makefile a propósito: el self-update corre este mismo script
# (`--daemon-only`), así que un binario producido por el botón Actualizar sella igual que uno
# hecho a mano. Si esto viviera en el Makefile, actualizarse desde la app devolvería un binario
# que se reporta como «dev» — la clase de mentira que este RF vino a sacar.
#
# La fuente de verdad del semver es Cargo.toml, la misma que ya usa el Makefile para bumpear.
# Hora LOCAL: el número se compara contra «cuándo compilé», y esa referencia es el reloj que
# el operador tiene delante.
VERSION="$(sed -n 's/^version = "\(.*\)"/\1/p' "$ROOT/web/src-tauri/Cargo.toml" | head -1)"
BUILD="$(date '+%y%m%d%H%M')"
COMPILADO="$(date '+%Y-%m-%d %H:%M')"
IDENT="github.com/alpacapurpura/arnesia/internal/adapters/selfupdate"
echo "   identidad: ${VERSION:-?}.$BUILD ($COMPILADO)"

go build -trimpath \
  -ldflags "-s -w -X '$IDENT.Version=$VERSION' -X '$IDENT.Build=$BUILD' -X '$IDENT.Compilado=$COMPILADO'" \
  -o "$ROOT/bin/arnesia" ./cmd/arnesia

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
