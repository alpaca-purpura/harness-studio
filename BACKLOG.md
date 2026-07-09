# BACKLOG — ArnesIA (lo-que-viene)

> vig: activo · revisar: 2026-08-01
> Hoja atómica de LO-QUE-VIENE (4º… 1er eje futuro del árbol). Solo items ABIERTOS. Item
> cerrado se BORRA (su cierre vive en `ledger/`). Formato: `[origen] item · tag`. Tags: `gate`
> (firma humana) · `deuda` · `bloqueo` (depende de otra cosa).
> SSoT del futuro: reemplaza los `Siguiente`/`Deuda` que vivían enterrados en LEDGER.

## Gates humanos pendientes (código listo, falta firma 🧑‍⚖️ PARIDAD)

- [chat-cc-funcional] firmar 7 desviaciones `PARIDAD.md` + mockup v1 · `gate`
- [franja-artefactos] gate final lado a lado, 7 desviaciones (HS-13 F5) · `gate`
- [boton-actualizar] gate final, 6 desviaciones `PARIDAD.md` · `gate`
- [inspector-drawer] gate final, 5 desviaciones registradas · `gate`

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
- [x] ✅ `scripts/estado.sh` genera la línea de ruleset a `ESTADO.md` desde conformance (RF-178) — HECHO
- [capabilities] R4 `cap-estado-generado` + `cap-puntero-estable`: siguen `pendiente` (defer honesto) — falta enforcer determinista · `deuda`
- [capabilities] cablear `scripts/estado.sh` a un hook/CI para las cifras de arch/knowledge (hoy tecleadas) · `deuda`
- [capabilities] validar los ~46 `sin-check`: partir de FE sin tests (solo stories) · `deuda`
- [capabilities] resolver dead-code candidatos: `domain/graph.go#UnidadDeTrabajo` · enums Banda/Canal/Procedencia/Origen · `app-store#toggleTheme` — decidir seed-futuro vs borrar · `deuda`
- [capabilities] índice SQLite real (CAP-21) + watcher fsnotify (CAP-23) + seed→JSONL corpus (CAP-22) — sale la fase-5 del stack · `bloqueo`

## Fuera de alcance ahora (anotado para no perderlo)

- [reorg-docs] `UX.md` (15k) y `METODOLOGIA.md` (8.5k) tienen el mismo mal (backlog+historia
  mezclados) — atacar en paquete aparte, no mezclar ejes · `deuda`
