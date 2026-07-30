# Decisiones — ciclo conversacional, cables del Dock

> **FIRMADAS 🧑‍⚖️ 2026-07-30 (Chris)** — vía aprobación del plan de sesión del MVP.

- **C-D1 — el único cable real de la tarjeta de mejora es `openChat()`.** Corrección
  al diagnóstico previo: descartar YA está cableado end-to-end
  (`workspace-stage.tsx:469-475` ↔ `router.go:52-55` ↔ `client.ts`) y `onProponer` ya
  puebla el buzón del composer. Falta: abrir el Dock (`openChat()` — 0 callers) +
  fijar el alcance a la caja del punto. El contexto viaja por el chip de alcance
  (`sendTurn` ya antepone `[alcance: …]`), NO ensuciando el composer. **BR-M12
  intacto:** prellenar + focus, JAMÁS auto-enviar; la tarjeta sigue sin escribir
  archivos.
- **C-D2 — «Editar conversando» se habilita por prop, solo cuando el arnés visto es
  el de la sesión** (CH-D6): `inspector.tsx` gana `onEditarConversando?`; con prop ⇒
  botón vivo (setScope + openChat); sin prop ⇒ disabled con title HONESTO nuevo
  («disponible cuando el arnés visto es el de la sesión» — el actual «Se cablea en
  Fase 3/4» dejó de ser verdad). «Editar fuente» (CodeMirror) queda Fase 2, sin tocar.
- **C-D3 — stretch watcher-instalaciones SOLO si sobra:** `watch.New(reg, extra)` con
  dedup por prefijo, reorden de wiring en `main.go` (watcher después del Portafolio),
  `ReevaluarDeriva` existente, SSE `event: portafolio` NUEVO (no se contamina
  `event: map`, cuyo contrato es `harness_id` del índice). Hereda RF-210 documentado.
  Recorte interno: FE → SSE → ticket entero.
