#!/usr/bin/env bash
# spec-guard — puente hook→validador (franja-artefactos D6, RF-131): UN solo validador
# (skills/spec-writer/scripts/validate_spec), dos puntos de enganche:
#   posttooluse → tras Write|Edit sobre spec.md; bloquea SOLO con JSON decision:block
#                 (en PostToolUse el exit-2 NO bloquea — ya pasó, L1.5 de hooks).
#   stop        → gate final de completitud; respeta stop_hook_active (cap 8, jamás bucle).
# Guía sin bloqueo (p6): sin jq o sin spec.md aplicable, el hook se retira en silencio.
set -euo pipefail

modo="${1:?uso: spec-guard.sh <posttooluse|stop>}"
input="$(cat)"

command -v jq >/dev/null 2>&1 || { echo "spec-guard: jq ausente — sin gate (guía sin bloqueo)" >&2; exit 0; }

raiz="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
validador="$raiz/skills/spec-writer/scripts/validate_spec"

bloquear() { # $1 = razón
  jq -n --arg r "$1" '{decision: "block", reason: $r}'
  exit 0
}

case "$modo" in
posttooluse)
  file="$(jq -r '.tool_input.file_path // empty' <<<"$input")"
  [[ "$(basename "$file")" == "spec.md" ]] || exit 0
  if ! salida="$("$validador" "$file" 2>&1)"; then
    bloquear "spec.md no valida contra la plantilla (references/plantilla-spec.md):
$salida"
  fi
  ;;
stop)
  [[ "$(jq -r '.stop_hook_active // false' <<<"$input")" == "true" ]] && exit 0
  cwd="$(jq -r '.cwd // empty' <<<"$input")"
  spec="${cwd:-$PWD}/spec.md"
  [[ -f "$spec" ]] || exit 0 # sin spec.md en el cwd no hay gate que aplicar aquí.
  if ! salida="$("$validador" "$spec" 2>&1)"; then
    bloquear "gate final: spec.md incompleto — corrige antes de terminar:
$salida"
  fi
  ;;
*)
  echo "spec-guard: modo desconocido '$modo'" >&2
  exit 0
  ;;
esac

exit 0
