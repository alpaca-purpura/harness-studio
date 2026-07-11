# BACKLOG — ArnesIA (lo-que-viene)

> vig: activo · revisar: 2026-08-01
> Hoja atómica de LO-QUE-VIENE (4º… 1er eje futuro del árbol). Solo items ABIERTOS. Item
> cerrado se BORRA (su cierre vive en `ledger/`). Formato: `[origen] item · tag`. Tags: `gate`
> (firma humana) · `deuda` · `bloqueo` (depende de otra cosa).
> SSoT del futuro: reemplaza los `Siguiente`/`Deuda` que vivían enterrados en LEDGER.

## Gates humanos pendientes (código listo, falta firma 🧑‍⚖️ PARIDAD)

- **Ninguno.** Los 4 (chat-cc-funcional · franja-artefactos · boton-actualizar · inspector-drawer)
  quedaron FIRMADOS 2026-07-09 (HS-20) — sus 7+7+6+5 desviaciones aceptadas; cierre en `ledger/HS-20.md`.

## Outcome ACTIVO — Fase 1 · Ciclo de forja de arneses vivo (2026-07-10)

> Modelo de terreno FIRMADO 🧑‍⚖️ (`stories/2026-07-10-terreno-conocimiento/`, D0-D20 + `arnes.yaml`).
> Ejecución en `stories/2026-07-10-forja-ciclo-vivo/`. Loop meta: `chat → arnes.yaml → gate → scaffold → Mapa`.
> Agnóstico al rubro (dev = un ejemplo). Slice fino primero (Shape-Up).

- [ ] **1. Dogfood scaffold** — `docs/terreno/{proposito,producto,organizacion}/` + `docs/wip/` derivados del
  `arnes.yaml` (INDEX/dim + hojas atómicas D9 + `knowledge/`), migrando `docs/architecture/`→`terreno/producto/`;
  reconciliar 11 dims ↔ 82 caps sin romper R1-R4 (P6) · `gate`
  - [ ] **1a** dimensión piloto `forma-trabajo` end-to-end (golden/fixture) + raíz terreno + esqueleto wip
  - [ ] **1b** migración `docs/architecture/`→`terreno/producto/` + reconciliación P6 (82 caps)
- [ ] **2. Forjador mínimo + gate de completitud** — motor que LEE `arnes.yaml`, corre el gate (D19, extiende
  `arnesia conformance`) y reproduce el golden determinísticamente (rellena PLANTILLAS). Dogfood: forjar arnés-dev · `gate`
- [ ] **3. Chat forja/edita** (corazón) — chat CC in-app: crear arnés por conversación → yaml→gate→scaffold; editar
  vía init/doctor/loop-forward (D8 sello/deriva/cosecha-back) · `gate`
- [ ] **4. Mapa DESTINO** — superset ESTRICTO del baseline renderizando terreno real (salud · WIP estados→done ·
  overlays calidad+economía · sello/deriva · receta en inspector) → portar a Storybook · `gate`
- [ ] **5. Después** — forjar 2° orden (`/po /architect /dev-team /auditor`) + upstream schema `arnes.yaml`+terreno al kit · `deuda`

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
- [x] ✅ **estado.sh → CI drift-gate + auto-cura (HS-21, eje N2·Paquete A — FIRMADO 2026-07-10):**
  `estado.sh --check` en el job `go` rompe el merge si las cifras del checkpoint quedan stale
  (neutraliza la fecha; `ci-estado-drift`), Y el hook `pre-commit estado-cifras` las **regenera solas**
  en el commit (`--check`→regen→`git add`; `lefthook-precommit-estado`). Cifras stale ya NO se cuelan.
  Cierre en `ledger/HS-21.md`; PARIDAD 10 filas firmada
- [capabilities] validar los ~40 `vivo·nc` (sin-check): construir el test que falta por-cap; parte de FE sin tests (solo stories) — es un paquete propio · `deuda`
- [capabilities] `toggleTheme` seed-futuro: cablear el toggle a la vista Ajustes (RF-100) o cortar · `deuda`
- [capabilities] índice SQLite real (CAP-21) + watcher fsnotify (CAP-23) + seed→JSONL corpus (CAP-22) — sale la fase-5 del stack · `bloqueo`

## Homologación de metodología — continuaciones (HS-19 cerrada, estos son los siguientes)

- [homologacion] **upstream del método al kit** `harness@prenter-marketplace` (nueva versión): `/pm` +
  `cap_doctor.py` + scaffolder + capability schema + soporte plugin-mode del seam (desviación #1) · `deuda`
- [homologacion] **forjar los 4 arneses secundarios** `/po` · `/architect` · `/dev-team` · `/auditor` (hoy `/pm` los referencia por auto-chain pero no existen — desviación #4) · `deuda`

## Backlog de UX (extraído de `ux.md` al sanear el eje historia/vigente, 2026-07-10)

- [ux] Vista A/B real (hoy «Evaluar A/B» solo salta al tren) · `deuda`
- [ux] ¿Flujo canónico/ideal por skill como concepto aparte del replay real? · `deuda`
- [ux] Detalle de evals del gate («Ver evals») + telemetría post-deploy por proyecto · `deuda`
- [ux] Historia: mapa por versión (exige snapshot del índice — decidir en fase 3) · `deuda`
- [ux] Onboarding/captura de base a fondo (hoy solo dock guionado) · `deuda`
- [ux] Multi-proyecto: ¿vista por proyecto instalado? · `deuda`
- [ux] Búsqueda global (componentes, corridas, hallazgos) · `deuda`
- [ux] Accesibilidad teclado completa (hoy parcial) · estados vacíos restantes · `deuda`
- [ux] Taxonomía: ¿subtipos de regla / clase L0 visible en el nodo? · `deuda`
- [ux] Leyenda del mapa: falta filtro para el tipo `command` (se renderiza, no se puede filtrar) · `deuda`
- [ux] Organigrama: posición ¿100% libre vs auto-layout+ajuste fino? ¿persistir posiciones como
  metadato? · marketplace por-arnés ¿override o hereda de la empresa? · reporta-a ¿cross-empresa o
  solo intra? · «＋ crear arnés para un puesto» desde el organigrama · `deuda`
- [ux] Badge de conformidad por-nodo sobre el mapa: conectar los checks del árbol de conocimiento a
  Diagnóstico real por arnés (hoy «ver arneses afectados» es demo) · `deuda`
- [ux] ⌘K quick-switch de sesión (palette P4) como añadido al rail, no reemplazo · `deuda`
- [ux] Persistencia de la lista de sesiones: índice SQLite desechable vs sidecar propio (JSONL
  sigue siendo fuente de verdad) · `deuda`
- [ux] Unificar al portar: shell-A-galaxia (it.13) + shell-A-sessions (it.14) + detalle v3 en un
  solo shell · `deuda`
- [ux] Empleados-IA del producto (Valeria·Lisa…) vs roster de dev — ¿dos vistas separadas? · `deuda`
