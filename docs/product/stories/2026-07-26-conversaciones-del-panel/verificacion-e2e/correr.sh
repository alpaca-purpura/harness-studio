#!/usr/bin/env bash
# correr.sh — el circuito E2E completo contra el BINARIO INSTALADO, reproducible.
#
#   bash verificacion-e2e/correr.sh [guion…]      # sin argumentos: los 6
#
# LO QUE NUNCA HACE, y es el punto:
#   · no toca `~/.arnesia` — el daemon corre con HOME apuntando a un sandbox
#   · no toca los proyectos del operador — el arnés de prueba vive en el sandbox
#   · no corre `make bump-*` (`make installer` depende de `bump-patch`; para el Modo A
#     alcanza `make dev-sync`, que compila la MISMA SPA y el MISMO daemon sin publicar)
#
# ⚠ `make dev-sync` REEMPLAZA ~/.local/bin/arnesia y le mata el daemon al operador.
#   Respaldalo antes:  cp -p ~/.local/bin/arnesia ~/.local/bin/arnesia.respaldo-$(date +%F-%H%M%S)
set -euo pipefail

RAIZ="$(cd "$(dirname "$0")/../../../../.." && pwd)"
AQUI="$(cd "$(dirname "$0")" && pwd)"
SB="${E2E_SANDBOX:-/tmp/arnesia-e2e-$(date +%s)}"
BIN="${E2E_BIN:-$HOME/.local/bin/arnesia}"
PIDF="$SB/daemon.pid"
GUIONES=("${@:-}")
[[ -z "${GUIONES[0]:-}" ]] && GUIONES=(1 2 3 4 5 6)

echo "== sandbox: $SB"
mkdir -p "$SB/.arnesia" "$SB/.claude/projects" "$SB/.config" "$SB/arneses-e2e/vitalia"
# semilla: COPIA (jamás mv) del registro real, para ejercitar la migración de verdad.
cp "$HOME/.arnesia/sessions.json" "$SB/.arnesia/sessions.json"
cp "$HOME/.arnesia/sessions.json" "$SB/REFERENCIA-sessions-v1.json"
# `sesiones-cerradas.json` NO se siembra: CV-D6 lo elimina y resucitarlo probaría lo contrario.
printf '# arnés de prueba\n' > "$SB/arneses-e2e/vitalia/README.md"
cat > "$SB/.arnesia/arneses.json" <<EOF
[{"arnes":"vitalia","path":"$SB/arneses-e2e/vitalia"},
 {"arnes":"sin-home~vitalia~vitalia","path":"$SB/arneses-e2e/vitalia"},
 {"arnes":"arnesia","path":"$SB/arneses-e2e/vitalia"}]
EOF
: > "$SB/.arnesia/mock-ctl"

arranca() { # arranca [flags…]
  env -i HOME="$SB" PATH="$PATH" USER="${USER:-x}" TERM=dumb \
    "$BIN" serve --addr 127.0.0.1:4200 --log - --claude "$AQUI/mock-claude-conv.sh" "$@" \
    > "$SB/daemon.log" 2>&1 &
  echo $! > "$PIDF"
  for _ in $(seq 1 80); do curl -sf http://127.0.0.1:4200/api/version >/dev/null 2>&1 && return 0; sleep 0.3; done
  echo "el daemon no levantó:"; tail -20 "$SB/daemon.log"; return 1
}
# Sin `pkill`: el patrón matchea la propia línea de comando de este script y se auto-mata.
para() { [[ -f "$PIDF" ]] || return 0; kill "$(cat "$PIDF")" 2>/dev/null || true; sleep 1; rm -f "$PIDF"; }
trap 'para' EXIT

reset() { para; rm -f "$SB/.arnesia/sesiones.json" "$SB/.arnesia"/sessions.json.v1-*.bak "$SB/.arnesia/mock-seq"
          cp "$SB/REFERENCIA-sessions-v1.json" "$SB/.arnesia/sessions.json"; : > "$SB/.arnesia/mock-ctl"
          rm -rf "$SB/.arnesia/sessions"; mkdir -p "$SB/arneses-e2e/vitalia"; }

# ── ASERCIÓN DE IDENTIDAD: se mide el binario que se acaba de construir, o nada ──
reset; arranca
SELLO="$(curl -s http://127.0.0.1:4200/api/version | sed -n 's/.*"version":"\([^"]*\)".*/\1/p')"
echo "== sello de build del binario probado: $SELLO"
[[ -n "$SELLO" ]] || { echo "sin sello: se aborta antes de medir otro binario" >&2; exit 1; }

for g in "${GUIONES[@]}"; do
  echo; echo "════ E2E-$g ════"
  case "$g" in
    1|2) reset; arranca ;;
    3)   reset; arranca --rotacion-umbral 5 ;;
    4)   reset; arranca ;;
    5)   reset; arranca; sleep 1; para
         jq --arg p "$SB/arneses-e2e/vitalia" '.sesiones |= map(.cwd = $p)' "$SB/.arnesia/sesiones.json" > "$SB/.arnesia/t" && mv "$SB/.arnesia/t" "$SB/.arnesia/sesiones.json"
         arranca ;;
    6)   reset; arranca; sleep 1; para
         jq '.sesiones[0].conversaciones |= map(if .activa then .claude_session_id = "cc-gc-eliminado-9999" else . end)' "$SB/.arnesia/sesiones.json" > "$SB/.arnesia/t" && mv "$SB/.arnesia/t" "$SB/.arnesia/sesiones.json"
         echo 'MOCK_RESUME_MUERE=cc-gc-eliminado-9999' > "$SB/.arnesia/mock-ctl"; arranca ;;
  esac
  MOCK_CTL="$SB/.arnesia/mock-ctl" E2E_ARNES_DIR="$SB/arneses-e2e/vitalia" \
    node "$AQUI"/e2e-"$g"-*.mjs || echo "── E2E-$g cerró con fallas (ver arriba)"
done

para
echo; echo "== md5 del registro del operador (tiene que ser el mismo de antes de correr):"
md5sum "$HOME/.arnesia/sessions.json"
echo "== sandbox conservado en $SB (borralo cuando termines de mirarlo)"
