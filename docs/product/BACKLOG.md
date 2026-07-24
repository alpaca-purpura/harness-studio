# BACKLOG — ArnesIA (lo-que-viene)

> vig: activo · revisar: 2026-08-01
> Hoja atómica de LO-QUE-VIENE (4º… 1er eje futuro del árbol). Solo items ABIERTOS. Item
> cerrado se BORRA (su cierre vive en `ledger/`). Formato: `[origen] item · tag`. Tags: `gate`
> (firma humana) · `deuda` · `bloqueo` (depende de otra cosa).
> SSoT del futuro: reemplaza los `Siguiente`/`Deuda` que vivían enterrados en LEDGER.

## Prioridad actual — Mejorar un arnés conversando (spike de spec, 2026-07-22)

- [ ] **Atar chat-cc-funcional + kit ② inyectado + reindex-en-vivo + Mapa + historial + rotación de
  contexto en UNA experiencia continua** — el operador señaló que hoy son piezas separadas: el chat ya
  edita el árbol real del arnés, pero el Mapa sirve una foto cacheada, el historial se borra al cerrar
  una sesión, y no hay rotación de contexto para conversaciones largas. Decidido esta sesión: alcance =
  chat libre + doctrina de fondo · Mapa refresca tras cada turno (medido `loader.LoadArnes` ~0.4ms, sin
  costo). **Faltan 3 forks por firmar** antes de `spec.md` (cada uno con recomendación en
  `spike-spec.md`): **A** instalación editable contradice la ley anti-drift (HS-25 sin guard, recom.
  bloquear) · **B** `Close()` borra el historial en vez de archivarlo (recom. archivar + listado por
  arnés) · **C** rotación de contexto por umbral de `ctxPct` hacia un proceso `claude` fresco, invisible
  para el usuario (mecanismo recomendado, 4 sub-preguntas abiertas).
  → [`stories/2026-07-22-mejorar-arnes-conversando/INDEX.md`](stories/2026-07-22-mejorar-arnes-conversando/INDEX.md) · `gate`
- Nota: NO es lo mismo que el ítem "Fase 1 forja-ciclo-vivo" de más abajo (ese es un flujo GUIADO
  `init`/`doctor`/`loop-forward`, pausado, fuera de alcance de este spike) ni el re-key
  `(home,id,scope)` de la deuda viva (plomería de identidad del Portafolio, sin relación directa).

## Rebrand PRENTER — gate FIRMADO (HS-25), deuda residual de diseño

- [x] ✅ **Chrome de tokens (color/tipografía/radios/sombras) reemplazado por el sistema PRENTER**
  (dark-first, negro+teal `#00b7aa`) — construido + gate humano 🧑‍⚖️ FIRMADO 2026-07-22 (HS-25). `color.kind`/
  `health`/`heat` (Mapa) INTACTOS. → [`stories/2026-07-15-rebrand-prenter-design-system/paridad.md`](stories/2026-07-15-rebrand-prenter-design-system/paridad.md)
- [ ] **Re-derivar `mockups/arnesia-shell-A-galaxia.html` (🔒 firmado) a los tokens PRENTER** + re-estampar
  su fila en `mockups/INDEX.md` — es el acto de cierre real que quedó pendiente de la firma de arriba
  (trabajo de diseño, no de paperwork); decidir alcance de botones/CTA por-módulo que siguen en
  `font.sans` sin QA visual (D8) · `deuda`

## Idea en research — Instalador público + licencias org (2026-07-15)

- [ ] **Instalador `curl\|sh` público + activación por key de organización** (org→N-sub-keys) —
  `state: idea`, research de mercado hecho y verificado en vivo (Keygen.sh CE recomendado).
  Alcance acotado con el operador: interno, NO venta externa de ArnesIA (no toca `vision.md`).
  Bloquea `refining`: confirmar en código real que `Groups` de Keygen vive en CE (no EE) +
  decidir módulo destino en el seam (`self-update` extendido vs. `distribucion` nuevo).
  → [`stories/2026-07-15-instalador-publico-licencias-org/INDEX.md`](stories/2026-07-15-instalador-publico-licencias-org/INDEX.md) · `deuda`

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

## Outcome — Programa «Portafolio · ciclo de vida del arnés» (2026-07-13)

> Del SPIKE [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md)
> (S-D5/S-D6, forks resueltos). Portafolio = **front-door del ciclo de vida** (crear·mapear·observar·mejorar,
> VISION). Reencuadrado spike→programa (F5). Agrupa por `empresa`; **`empresa` ⊥ `marketplace`** (ambos YA en
> `arnes.l0.json` + `domain.Arnes`). Procedencia = **lock `.devstudio/arneses.yaml`** (HS-12, detector 3°).
> Auth GitHub = **`gh` si existe / PAT fallback** (user-owned, auth-terms firmado — app = conductor). Update =
> **híbrido** (notify propio semver-vs-tag + re-materializa plugin-native CC). Cierre del spike = firmar el modelo de entidades.
> **Unidad = arnés por identidad `(home, id)`** (S-D10/F-A; home = marketplace autor-declarado): 1 = **canónico** (única
> copia editable) + N **instalaciones** (espejos read-only). Arnés↔empresa↔registry-de-adquisición = **N:M:M** (facetas).
> **Ley anti-drift (descriptiva de ArnesIA):** ArnesIA nunca introduce una copia editable divergente nueva; observa +
> re-materializa (reparar) + backportea; el marketplace `home` = SSoT (pull-antes-de-push). **`deriva`** (ex-«drift») =
> hash vs `home/plugins/<id>/<versión>/`. **Recut post-revisión adversaria (S-D9/S-D10):** el FE no es construible sin cimientos.

- [x] ✅ **0. Cimientos del Portafolio** (dominio/backend, SIN FE) — **CONSTRUIDO Y FIRMADO 🧑‍⚖️** (Sonnet 5,
  T1-T9 · gate firmado 2026-07-13, HS-23): store de portafolio **separado** de `arneses.json` (degrada honesto) ·
  **walker** `.claude/plugins/<id>/` + **fallback `plugin.json`** · campo `version` · `empresas[]` (`registries[]`
  vive en `EntradaPortafolio`, S0-D3) · identidad `(home,id)` + canonicalización · cadena de origen **collect-all**
  (lock≙cc-plugins>git-plugin>manifiesto) · **`deriva`** = hash vs `home/plugins/<id>/<v>/` (o `deriva-no-evaluable`) ·
  detector lock · HTTP+CLI · capabilities+boundary as-code. 27 tests Go nuevos + E2E vivo contra la máquina real
  (cazó y corrigió 2 bugs reales, S0-D14/D15).
  → [`stories/2026-07-13-portafolio-slice0-cimientos/paridad.md`](stories/2026-07-13-portafolio-slice0-cimientos/paridad.md)
  - [x] ✅ **eslabón CC investigado** (Fable 5, 2026-07-13, S0-D1) + **corregido con evidencia real** (Sonnet 5,
    S0-D14): `~/.claude/plugins/installed_plugins.json` (v2) anida las entradas bajo una clave `plugins` —
    `{"version":2,"plugins":{"id@marketplace":[...]}}`, no en el nivel raíz como se había asumido en S0-D1.
  - **Deuda** (registrada en `paridad.md`, no bloquea la firma): ~~re-key del índice in-memory a
    `(home,id,scope)` calificado~~ **CERRADA 2026-07-23** — `IndexPort.Upsert` gana la clave explícita,
    dos arneses del mismo id pelado ya no se pisan (ver `docs/product/capabilities/portafolio/
    observar-en-mapa.yaml`). Sigue abierta: política definitiva de conservación de entradas corruptas
    del store más allá de un ciclo save (hoy: sobrevive si es JSON sintácticamente válido, se pierde
    si el archivo entero está roto — ver `paridad.md` desviación #4).
- [x] ✅ **1. FE Portafolio** — construido (Sonnet 5, T1-T8, 2026-07-14) + gate humano 🧑‍⚖️ FIRMADO
  2026-07-22 (HS-25). 3 superficies (Lista/lente empresa · Wizard-Proyecto/carpeta-local · Drawer READ) +
  Observar (Abrir en Mapa, cierra GAP-1) + Desvincular con confirmación; 34 stories `play()` +
  `selectors.test.ts` · 4 capabilities `fe-portafolio/*` `vivo` (R4) · E2E vivo real.
  → [`stories/2026-07-13-portafolio-slice1-fe/paridad.md`](stories/2026-07-13-portafolio-slice1-fe/paridad.md)
- [ ] **2. Agregar de marketplace → clonar + mejorar** — git url → **validar `marketplace.json`** → listar → elegir → checkout
  `<checkouts>/<home-slug>/<id>/` (`gh`/PAT) → chat/mejorar · `gate` · `bloqueo`(1). Paquete abierto
  (sin mockup aún, 1 nota PENDIENTE-RESOLVER sobre reconciliación proyecto↔marketplace-de-origen) →
  [`stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md`](stories/2026-07-23-portafolio-agregar-marketplace/INDEX.md)
- [ ] **3. Publicar** — pull/rebase → conformance-verde → bump semver → changelog obligatorio → push → **tag-tras-push** · `gate` · `bloqueo`(2)
- [ ] **4. Update-check + notify** — versión vs último tag de `home` + changelog + **alias de rename** (no cegar el aviso) · `gate` · `bloqueo`(3)
- [ ] **5. Reparar + Backport** — reparar overwrite SOLO dir privado (superficies compartidas = merge) + bloqueo si lock DevStudio +
  `conformance --arnes`; backport instalación→canónico (alineación de versión) · `gate` · `bloqueo`(1)
- Asociación de negocio (`reporta_a`, arneses que se nutren mutuamente) → materializa en **Galaxia** (Mapa), fuera de alcance inmediato · `deuda`

## Deuda viva (registrada, no bloquea la línea principal)

- [a11y] **`.text-warn` no cumple contraste mínimo** (axe `color-contrast`, ratio 3.76 vs 4.5
  requerido — `#c96a2e` sobre `#ffffff`): destapado 2026-07-23 corriendo
  `new-session-picker.stories.tsx` (4 stories fallan en la a11y gate cuando
  `ConversacionesDelArnes` renderiza su error de fetch, `.text-warn`). Preexistente, sin
  relación con el re-key — fix de token de color, no tocado en este cierre · `deuda`
- [HS-09/11/telemetria-de-nacimiento] **telemetría embebida vía OTel nativo ⇒ desbloquea la capa
  Tokens del Mapa** — **arquitectura RESUELTA 2026-07-24** (ya NO es "bloqueo" de diseño; era 2
  ítems del BACKLOG que resultaron ser el mismo trabajo). Verificado con doc oficial de Claude
  Code: `CLAUDE_CODE_ENABLE_TELEMETRY` + OTel nativo YA emite `claude_code.token.usage`/
  `claude_code.cost.usage` con atribución `skill.name`/`tool_name` limpia — sin hook custom, sin
  Langfuse, sin parsear JSONL (coherente con `conductor-no-parsea-jsonl.md`). Diseño: receptor
  OTLP embebido loopback-only en el daemon + scaffold inyecta env vars (nunca infraestructura
  externa — orden del operador: el instalador jamás depende de Langfuse/Docker). Detalle en
  [`docs/architecture/boundaries/telemetria-de-nacimiento.md`](../architecture/boundaries/telemetria-de-nacimiento.md)
  v2.0 → [`stories/2026-07-24-telemetria-embebida-otel/INDEX.md`](stories/2026-07-24-telemetria-embebida-otel/INDEX.md).
  Falta: mockup de la capa Tokens (ni el mockup "destino" la dibuja) → spec → build → PARIDAD ·
  `deuda` (ya no `bloqueo` — el diseño está resuelto, falta construir)
- [HS-09/11] **capas Desempeño/Proceso del Mapa** — estas SÍ siguen genuinamente bloqueadas (no
  recortables como Tokens): Desempeño necesita latencia/reintentos que solo OTel `hook_*`/
  `api_retry` ve + diseño de qué es "desempeño" a nivel Mapa; Proceso necesita mapear eventos a
  fases del arnés (diseño no trivial). Mismo receptor OTLP embebido del ítem de arriba las
  alimentaría, pero necesitan su propio diseño de datos primero · `bloqueo`
- [HS-11] **FE del botón «Correr» de una caja** — el backend YA es async con gate post-run
  (`POST .../boxes/{id}/run` → 202+run_id · `GET .../runs/{runId}` · construido 2026-07-23);
  `inspector.tsx` («Corridas donde actuó») solo tiene prosa «al implementar…», sin botón, sin
  fetch, sin spinner. **Paquete completo abierto y PAUSADO 2026-07-23** (mockup publicado +
  decisiones D1-D4 firmes: mecanismo SSE `event: run` no polling, error 409 inline sin toast) —
  el operador pidió la explicación funcional del botón antes de firmar el mockup, se le dio
  (grounded en vision.md A1-A3 + `run_service.go` T3 real), y eligió pausar sin rechazar ni
  pedir cambios → [`stories/2026-07-23-boton-correr-caja/INDEX.md`](stories/2026-07-23-boton-correr-caja/INDEX.md) · `deuda`
- [doctrina-una-fuente-dos-targets] `kit/doctrine.md` es prosa mantenida a mano, sin drift-check
  contra `docs/architecture/knowledge/`/METODOLOGIA — construir el freshness-check (candidato:
  verificar que las secciones citadas por número siguen existiendo en esa forma) · `deuda`
- [chat] fase presentación: assistant-ui + CodeMirror merge + widgets ricos (decisión #5) · `deuda`
- [HS-09] 212 checks `deferred` → correr en CI (hoy solo la ruta `--arnes`) · `deuda`
- [HS-16] loader reconocedores `deferred`: subagent · plugin-nodo-raíz · edges-de-librería (necesitan diseño) · `deuda`
- [HS-16 Grupo A] 6 checks composición `deferred` (rediseño de motor; bloqueado por SQLite fase5 / OTel / modo-por-fase) · `bloqueo`
- [HS-16 Grupo C] `gate-honesto`: necesita diseño previo · `deuda`

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
