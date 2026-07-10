#!/bin/bash
# mock-claude — binario `claude` falso que habla stream-json REAL por stdout/stdin
# (patrón CUI tests/__mocks__/claude, investigación §Frente ①). Guioniza la casuística
# del spec: allow-sesión+grant · deny · allow-una-vez+re-pregunta · interrupt.
# Todo lo recibido/emitido queda en $MOCK_LOG para asserts.
LOG="${MOCK_LOG:-/tmp/mock-claude.log}"
echo "ARGS: $*" >>"$LOG"

emit() {
  printf '%s\n' "$1"
  echo "OUT: $1" >>"$LOG"
}

emit '{"type":"system","subtype":"init","session_id":"cc-mock-0001","model":"claude-mock"}'

turn=0
while IFS= read -r line; do
  echo "IN: $line" >>"$LOG"
  case "$line" in
  *'"subtype":"interrupt"'*)
    emit '{"type":"result","subtype":"success","result":"interrumpido — no completé el turno","usage":{"input_tokens":900}}'
    ;;
  *'"type":"user"'*)
    turn=$((turn + 1))
    case $turn in
    1) # allow-sesión: pide Edit; tras el allow pide OTRO Edit (grant debe auto-permitir).
      emit '{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"Voy a endurecer el gate de builder. "}}}'
      emit '{"type":"control_request","request_id":"cr-a1","request":{"subtype":"can_use_tool","tool_name":"Edit","input":{"file_path":"skills/builder/SKILL.md","old_string":"    tipo: manual","new_string":"    tipo: auto\n    eval: pnpm test --run"},"tool_use_id":"toolu_a1"}}'
      ;;
    2) # deny del operador.
      emit '{"type":"control_request","request_id":"cr-b1","request":{"subtype":"can_use_tool","tool_name":"Write","input":{"file_path":"skills/builder/SKILL.md","content":"contenido nuevo completo"},"tool_use_id":"toolu_b1"}}'
      ;;
    3) # allow-una-vez: el grant de 1s expira antes del segundo ask ⇒ re-pregunta.
      emit '{"type":"control_request","request_id":"cr-c1","request":{"subtype":"can_use_tool","tool_name":"MultiEdit","input":{"file_path":"skills/builder/SKILL.md","edits":[{"old_string":"p","new_string":"q"}]},"tool_use_id":"toolu_c1"}}'
      ;;
    4) # interrupt: ask de Bash que nadie contestará — el Stop lo deniega e interrumpe.
      emit '{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"Necesito correr los tests. "}}}'
      emit '{"type":"control_request","request_id":"cr-d1","request":{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"pnpm test --run"},"tool_use_id":"toolu_d1"}}'
      ;;
    *) # cualquier turno extra: eco simple.
      emit '{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"ok"}}}'
      emit '{"type":"result","subtype":"success","result":"ok","usage":{"input_tokens":100}}'
      ;;
    esac
    ;;
  *'"type":"control_response"'*)
    case "$line" in
    *cr-a1*)
      if [[ "$line" == *'"behavior":"allow"'* ]]; then
        emit '{"type":"control_request","request_id":"cr-a2","request":{"subtype":"can_use_tool","tool_name":"Edit","input":{"file_path":"skills/builder/SKILL.md","old_string":"success: viejo","new_string":"success: suite verde"},"tool_use_id":"toolu_a2"}}'
      else
        emit '{"type":"result","subtype":"success","result":"entendido, no toco nada","usage":{"input_tokens":500}}'
      fi
      ;;
    *cr-a2*)
      emit '{"type":"stream_event","event":{"type":"content_block_delta","delta":{"type":"text_delta","text":"Gate endurecido y capability alineada."}}}'
      emit '{"type":"result","subtype":"success","result":"Gate endurecido y capability alineada.","usage":{"input_tokens":1200}}'
      ;;
    *cr-b1*)
      emit '{"type":"result","subtype":"success","result":"ok, no toqué el archivo","usage":{"input_tokens":600}}'
      ;;
    *cr-c1*)
      if [[ "$line" == *'"behavior":"allow"'* ]]; then
        sleep 1.6
        emit '{"type":"control_request","request_id":"cr-c2","request":{"subtype":"can_use_tool","tool_name":"MultiEdit","input":{"file_path":"skills/builder/SKILL.md","edits":[{"old_string":"r","new_string":"s"}]},"tool_use_id":"toolu_c2"}}'
      else
        emit '{"type":"result","subtype":"success","result":"denegado, paro","usage":{"input_tokens":300}}'
      fi
      ;;
    *cr-c2*)
      emit '{"type":"result","subtype":"success","result":"segunda edición aplicada","usage":{"input_tokens":700}}'
      ;;
    esac
    ;;
  esac
done
