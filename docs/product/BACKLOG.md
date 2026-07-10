# BACKLOG — ArnesIA (lo-que-viene)

> vig: activo · revisar: 2026-08-01
> Hoja atómica de LO-QUE-VIENE (4º… 1er eje futuro del árbol). Solo items ABIERTOS. Item
> cerrado se BORRA (su cierre vive en `ledger/`). Formato: `[origen] item · tag`. Tags: `gate`
> (firma humana) · `deuda` · `bloqueo` (depende de otra cosa).
> SSoT del futuro: reemplaza los `Siguiente`/`Deuda` que vivían enterrados en LEDGER.

## Gates humanos pendientes (código listo, falta firma 🧑‍⚖️ PARIDAD)

- **Ninguno.** Los 4 (chat-cc-funcional · franja-artefactos · boton-actualizar · inspector-drawer)
  quedaron FIRMADOS 2026-07-09 (HS-20) — sus 7+7+6+5 desviaciones aceptadas; cierre en `ledger/HS-20.md`.

## Deuda viva (registrada, no bloquea la línea principal)

- [HS-09/11] telemetría JSONL → indexer real ⇒ desbloquea capas Tokens/Desempeño/Proceso del Mapa · `bloqueo`
- [HS-11/chat] spike `control_response` vs claude real (confirmar en papel/e2e) · `deuda`
- [HS-11] run async del `/boxes/{id}/run` + gate post-run · `deuda`
- [HS-11] 3 boundaries de research → materializar en `arch/` · `deuda`
- [chat] fase presentación: assistant-ui + CodeMirror merge + widgets ricos (decisión #5) · `deuda`
- [HS-09] 212 checks `deferred` → correr en CI (hoy solo la ruta `--arnes`) · `deuda`
- [HS-12] loader detector 3°: leer lock `.devstudio/arneses.yaml` y resolver multi-arnés · `deuda`
- [HS-16] loader reconocedores `deferred`: subagent · plugin-nodo-raíz · edges-de-librería (necesitan diseño) · `deuda`
- [HS-16 Grupo A] 6 checks composición `deferred` (rediseño de motor; bloqueado por SQLite fase5 / OTel / modo-por-fase) · `bloqueo`
- [HS-16 Grupo C] `gate-honesto`: necesita diseño previo · `deuda`
- [HS-14] deep-link `arnesia://` en callback single-instance (ojo bug tauri#12726) · `deuda`

## Capabilities / doctrina (reorg 2026-07-09, FIRMADO)

- [x] ✅ validador R1/R2 construido + boundary `codigo-traza-a-capability` **enforced** (arch-test real, 0 colgantes / 0 huérfanos / cobertura 100%) + job lefthook `capabilities` — HECHO 2026-07-09
- [x] ✅ **R4 aterrizado (HS-20):** `cap-estado-consistente` + `cap-puntero-estable` = arch-test determinista (`TestCapabilityStatusConsistent`/`TestCapabilityPointersStable`), `--todo` pass 40→42 — HECHO 2026-07-09
- [x] ✅ **`scripts/estado.sh` genera TODO el bloque de cifras** (ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities) — mata la deuda «tecleadas»; cazó la línea stale «24 vivo/46 sin-check/6 STUB» → real `38 vivo · 40 vivo·nc · 1 parcial · 3 stub` — HECHO 2026-07-09 (HS-20)
- [x] ✅ **dead-code decidido (HS-20): NADA es dead real** — `Banda`/`Canal`/`Procedencia` vivos (loader/index) · `Origen` vivo en schema+FE (el tipo Go es espejo del dominio) · `UnidadDeTrabajo` = vocabulario del Spine (motor pendiente) · `toggleTheme` único «sin caller» → todos **seed-futuro**, cero borrado (no se destruye vocabulario de schema firmado)
- [capabilities] **derivación LIVE del estado** (`vivo ⟺ check verde` corriendo cada test, no solo consistencia) → cablear a CI · `deuda`
- [capabilities] cablear `scripts/estado.sh` a un hook/CI para que las cifras se regeneren solas (hoy es manual) · `deuda`
- [capabilities] validar los ~40 `vivo·nc` (sin-check): construir el test que falta por-cap; parte de FE sin tests (solo stories) — es un paquete propio · `deuda`
- [capabilities] `toggleTheme` seed-futuro: cablear el toggle a la vista Ajustes (RF-100) o cortar · `deuda`
- [capabilities] índice SQLite real (CAP-21) + watcher fsnotify (CAP-23) + seed→JSONL corpus (CAP-22) — sale la fase-5 del stack · `bloqueo`

## Homologación de metodología — continuaciones (HS-19 cerrada, estos son los siguientes)

- [homologacion] **upstream del método al kit** `harness@prenter-marketplace` (nueva versión): `/pm` +
  `cap_doctor.py` + scaffolder + capability schema + soporte plugin-mode del seam (desviación #1) · `deuda`
- [homologacion] **replicar `docs/` + seam + `/pm` a cockpit y dev-studio** (el árbol ya probado E2E en harness-studio) · `deuda`
- [homologacion] **forjar los 4 arneses secundarios** `/po` · `/architect` · `/dev-team` · `/auditor` (hoy `/pm` los referencia por auto-chain pero no existen — desviación #4) · `deuda`

## Fuera de alcance ahora (anotado para no perderlo)

- [reorg-docs] `docs/product/ux.md` (15k) y `docs/process/metodologia.md` (8.5k) tienen el mismo mal
  (backlog+historia mezclados) — atacar en paquete aparte, no mezclar ejes · `deuda`
- [reorg-docs] 30 links rotos en snapshots históricos (`stories/`·`research/`·`ledger/HS-NN.md`) — se
  dejaron a propósito en el cierre HS-19 (registros congelados, PARIDAD desviación #7); re-apuntarlos
  (si se decide) es trabajo de este paquete, no de la homologación · `deuda`
