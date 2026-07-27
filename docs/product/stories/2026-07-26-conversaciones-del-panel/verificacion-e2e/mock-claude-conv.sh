#!/usr/bin/env bash
# mock-claude-conv — `claude` falso que habla stream-json REAL, hecho para los guiones de
# ESTE paquete (el de `2026-07-08-chat-cc-funcional/e2e/mock-claude.sh` guioniza la
# casuística de PERMISOS; acá hace falta un turno liso y un turno LENTO).
#
# Modos, por variable de entorno:
#   MOCK_DELAY=<seg>   demora antes de responder  → deja la conversación en `streaming`
#   MOCK_ASK=1         pide un permiso y NO cierra → deja la sesión en `await`
#   MOCK_CC_ID=<id>    el session_id que reporta el init (para verificar la cadena_cc)
LOG="${MOCK_LOG:-/tmp/mock-claude-conv.log}"
echo "ARGS: $*" >>"$LOG"

# MOCK_RESUME_MUERE: si el spawn trae `--resume <id>` con el id marcado como muerto, el
# proceso CAE ANTES del init — que es literalmente lo que hace el CLI real cuando el corpus
# de esa sesión fue GC'd. Es el disparador de `tryHealResume`.
if [[ -f "$HOME/.arnesia/mock-ctl" ]]; then source "$HOME/.arnesia/mock-ctl"; fi
if [[ -n "${MOCK_RESUME_MUERE:-}" && "$*" == *"--resume $MOCK_RESUME_MUERE"* ]]; then
  echo "MUERO: resume no reconocido ($MOCK_RESUME_MUERE)" >>"$LOG"
  echo "No conversation found with session ID: $MOCK_RESUME_MUERE" >&2
  exit 1
fi

emit() { printf '%s\n' "$1"; echo "OUT: $1" >>"$LOG"; }

# Cada SPAWN reporta un session_id distinto. Es lo que hace observable la rotación: el
# `cc-id` nuevo del detalle y el crecimiento de `cadena_cc` sólo significan algo si el
# proceso fresco se identifica distinto del que se cerró.
SEQ_F="$HOME/.arnesia/mock-seq"
SEQ=$(( $(cat "$SEQ_F" 2>/dev/null || echo 0) + 1 ))
echo "$SEQ" > "$SEQ_F"
CC_ID="${MOCK_CC_ID:-$(printf 'cc-mock-%04d' "$SEQ")}"
emit "{\"type\":\"system\",\"subtype\":\"init\",\"session_id\":\"$CC_ID\",\"model\":\"claude-mock\"}"

while IFS= read -r line; do
  echo "IN: $line" >>"$LOG"
  case "$line" in
  *'"subtype":"interrupt"'*)
    emit '{"type":"result","subtype":"success","result":"interrumpido","usage":{"input_tokens":100}}'
    ;;
  *'"type":"user"'*)
    # El control vive en un ARCHIVO y no sólo en el entorno: el daemon hereda su env al
    # spawnear, así que una variable exigiría reiniciar el daemon para cambiar de modo —
    # y el guion necesita alternar «turno liso» y «turno lento» sobre el MISMO proceso.
    CTL="$HOME/.arnesia/mock-ctl"
    [[ -f "$CTL" ]] && source "$CTL"
    [[ -n "${MOCK_DELAY:-}" ]] && sleep "$MOCK_DELAY"
    if [[ "${MOCK_ASK:-}" == "1" ]]; then
      emit '{"type":"control_request","request_id":"cr-z1","request":{"subtype":"can_use_tool","tool_name":"Edit","input":{"file_path":"README.md","old_string":"a","new_string":"b"},"tool_use_id":"toolu_z1"}}'
      continue
    fi
    # MOCK_TOKENS mueve el ctx_pct: el daemon lo calcula como uso/ventana (conductor.go:773),
    # así que reportar tokens es la palanca para llegar al umbral de rotación sin un modelo real.
    emit '{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"recibido. "}}}'
    emit "{\"type\":\"result\",\"subtype\":\"success\",\"result\":\"recibido.\",\"usage\":{\"input_tokens\":${MOCK_TOKENS:-120}}}"
    ;;
  esac
done
