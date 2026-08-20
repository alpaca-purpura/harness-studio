#!/usr/bin/env bash
# Build completo: UI estática + binario.
#
# Vendored ArnesIA: upstream leía la versión de `core-harness/VERSION`, un archivo del
# repo de origen que acá no existe — el grep fallaba en silencio y el binario quedaba
# sellado `dev`. La versión sale del mismo lugar que la del resto del repo host
# (web/src-tauri/Cargo.toml, la fuente que usan bundle.py e installer.ps1).
set -euo pipefail
cd "$(dirname "$0")"
./build-ui.sh
VERSION="$(sed -n 's/^version = "\(.*\)"/\1/p' ../../../web/src-tauri/Cargo.toml 2>/dev/null | head -1 || true)"
go build -ldflags="-s -w -X main.version=${VERSION:-dev}" -o cockpit .
echo "✓ $(ls -lh cockpit | awk '{print $5}') → ./cockpit (v${VERSION:-dev})"
