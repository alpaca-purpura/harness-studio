#!/usr/bin/env bash
# bump.sh — sube la versión del bundle de escritorio a <X.Y.Z>, con el changelog como condición.
#
# Paquete 2026-07-26-versionado-y-changelog-metodologicos (VC-D2). Un solo punto de bump, igual
# que `bundle.sh` es el único punto de sello (RF-231/B-D2): los tres targets del Makefile
# (`bump-patch`/`bump-minor`/`bump-major`) y por herencia `make installer` pasan por acá, así que
# no existe un camino que suba la versión sin registrar qué cambió.
#
# ORDEN DELIBERADO — el changelog se valida ANTES de tocar un solo manifiesto:
#   1. `changelog.py check --exige-entradas`  → si [Sin publicar] está vacía, aborta acá
#   2. sed sobre los 3 manifiestos (Cargo.toml = fuente de verdad; los otros dos, sincronizados)
#   3. `changelog.py release X.Y.Z`           → promueve [Sin publicar] → ## [X.Y.Z] — fecha
# Si el paso 1 abortara DESPUÉS del sed, quedaría un working tree con la versión subida y sin
# changelog: exactamente el estado a medias que este script existe para que no pase.
#
# NO toca git (ni commit ni tag): el resultado queda en el working tree para revisarlo.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NUEVA="${1:-}"

if [[ ! "$NUEVA" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "uso: bump.sh <X.Y.Z>   (semver plano, sin prefijo v — lo exige Keygen.Release.version)" >&2
  exit 1
fi

CARGO="$ROOT/web/src-tauri/Cargo.toml"
TAURI="$ROOT/web/src-tauri/tauri.conf.json"
PKG="$ROOT/web/package.json"
ACTUAL="$(sed -n 's/^version = "\(.*\)"/\1/p' "$CARGO" | head -1)"

if [[ -z "$ACTUAL" ]]; then
  echo "no pude leer [package].version de $CARGO" >&2
  exit 1
fi
if [[ "$ACTUAL" == "$NUEVA" ]]; then
  echo "la versión ya es $NUEVA — nada que bumpear" >&2
  exit 1
fi

# 1) El gate. Falla acá = no se tocó nada.
python3 "$ROOT/scripts/changelog.py" check --exige-entradas

# 2) Los 3 manifiestos, en lockstep (TestVersionManifestsInSync los vigila).
echo "version: $ACTUAL -> $NUEVA"
sed -i "s/^version = \"$ACTUAL\"/version = \"$NUEVA\"/" "$CARGO"
sed -i "0,/\"version\":/s/\"version\": \"[^\"]*\"/\"version\": \"$NUEVA\"/" "$TAURI"
sed -i "0,/\"version\":/s/\"version\": \"[^\"]*\"/\"version\": \"$NUEVA\"/" "$PKG"

# 3) El changelog pasa a ser el de esta versión.
python3 "$ROOT/scripts/changelog.py" release "$NUEVA"

echo "OK — $NUEVA en los 3 manifiestos + CHANGELOG.md. Falta commitear (el bump no toca git)."
