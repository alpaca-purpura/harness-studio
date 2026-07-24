# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-26** (chat dock legible y ergonómico — 6 quejas + cariño markdown,
CAP-99/CAP-100, 3 firmas en el día, 2026-07-22). Índice de historia → `LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

- **Chat dock · legibilidad y ergonomía (2026-07-22) — CERRADO Y FIRMADO 🧑‍⚖️ (HS-26).**
  Las 6 quejas + cariño markdown, verificadas E2E vivo (claude real contra vitalia); CAP-99 +
  CAP-100. Paquete [`stories/2026-07-22-chat-dock-ux/`](stories/2026-07-22-chat-dock-ux/INDEX.md).
  Deuda viva (no reabre): `make installer` para llevarlo al escritorio → `BACKLOG.md`; errores
  posteriores = bugfix en paquete nuevo.
- **Programa «Portafolio · ciclo de vida del arnés» — modelo FIRMADO 🧑‍⚖️ (HS-22); Slice 0
  FIRMADO 🧑‍⚖️ (HS-23); Slice 1 «FE» FIRMADO 🧑‍⚖️ (HS-25); Slice 2 «sello/Identificar» FIRMADO 🧑‍⚖️ (HS-25).**
  El spike de CARGA se reencuadró en el front-door del ciclo de vida (agregar de marketplace/proyecto · observar · mejorar ·
  publicar · actualizar · reparar). Modelo firmado: identidad **`(home,id)`** · N:M:M · **canónico** (editable) + N **instalaciones**
  (read-only) · ley anti-drift descriptiva · **`deriva`**. Revisión adversaria (4 subagentes) probó que el FE no es construible sin
  cimientos → recut **Slice 0 «Cimientos» (dominio) → Slice 1 «FE»** (épica en `BACKLOG.md`). Paquete
  [`stories/2026-07-10-spike-carga-arneses/`](stories/2026-07-10-spike-carga-arneses/INDEX.md) (specs + casuística + revisión + mockup v2).
  **Hand-off:** Fable 5 (arquitecto) planificó → Sonnet 5 (constructor) ejecutó Slice 0 T1-T9 y Slice 1
  T1-T8 — los 8 tickets YA en `main` (T8 = `b35ff6c`, commiteado por el operador). **Auditoría post-build
  de la carga (Fable 5, 2026-07-14) EJECUTADA:** build sano; CI-rojo por lint preexistente REPARADO,
  motor conformance extendido a tests colocados (pass 42→46), docs stale sincronizadas →
  [`stories/2026-07-14-auditoria-portafolio-carga/`](stories/2026-07-14-auditoria-portafolio-carga/informe.md).
  Ver «Retomar aquí».
- **Fase 1 «forja-ciclo-vivo» — PAUSADA** ([`stories/2026-07-10-forja-ciclo-vivo/`](stories/2026-07-10-forja-ciclo-vivo/INDEX.md)):
  Slice 1a (golden scaffold `docs/terreno/` dim piloto `forma-trabajo` + `docs/wip/`) **construido, pendiente firma
  🧑‍⚖️**; verificado verde (F-D4). **Aclaración F-D5:** la expertise de forja se encarna en el **kit ② embebido**
  (cuerpos ①②, HS-10) — se inyecta al chat in-app; NO es `docs/terreno/` top-level (eso es un ③ dogfood mal-ubicado,
  migrará, no revertir). Firmado «docs/terreno absorbe architecture» → bajo revisión (P9 en terreno-conocimiento).
- **Modelo de TERRENO FIRMADO 🧑‍⚖️ 2026-07-10** ([`stories/2026-07-10-terreno-conocimiento/`](stories/2026-07-10-terreno-conocimiento/INDEX.md):
  D0-D20 + `arnes.yaml` + `estructura-terreno.html` v4). Sigue válido como **schema que el forjador aplica**.
  Slice 0/1/2 los 3 FIRMADOS (`ledger/HS-23.md`, `ledger/HS-25.md`). Slice 1 3 superficies (Lista/
  lente-empresa · Wizard-Proyecto/carpeta-local · Drawer READ) + Observar/Desvincular, E2E vivo (GAP-1
  cerrado); Slice 2 huella de path + modo degradado del Mapa + acción «Identificar» in-situ, E2E vivo
  (circuito crudo→degradado→identificar→sellado). Próximo paso del programa (item 2, «Agregar de
  marketplace → clonar + mejorar») desbloqueado en `BACKLOG.md`.
  Pausadas: **Fase 1 forja** (Slice 1a pendiente firma) · **terreno** (schema firmado). Contexto: memorias
  `hs-forja-fase1-carga-spike` · `hs-terreno-modelo-firmado` · `hs-linea-base-ui-storybook-ssot` · `hs-chat-cc-funcional`.
- **Rebrand PRENTER (2026-07-15) — FIRMADO 🧑‍⚖️ (HS-25).** Tokens dark-first en `web/`, paleta
  funcional del Mapa intacta. Deuda residual (no bloquea): re-derivar `arnesia-shell-A-galaxia.html` +
  re-estampar `mockups/INDEX.md` → `BACKLOG.md`.
- **Shell — Topbar sin empresa + selector de arnés (2026-07-20) — FIRMADO 🧑‍⚖️ (HS-25).**
  `topbar.tsx`/`session-rail.tsx`/`new-session-picker.tsx`/`portafolio-picker-store.ts`, 168/168 verde.
- **Retomar aquí: paquete `mejorar-arnes-conversando` — IMPLEMENTADO COMPLETO 2026-07-22
  (ejecución autónoma /goal), pendiente SOLO gate humano 🧑‍⚖️ de PARIDAD.**
  [`stories/2026-07-22-mejorar-arnes-conversando/INDEX.md`](stories/2026-07-22-mejorar-arnes-conversando/INDEX.md) —
  spike cerrado (Forks A4/B2/C + grounding firmados, MC-D1..D8) → `spec.md` RF-183..206 → los 11
  tickets landeados con E2E vivo (evidencia en `PARIDAD.md`): loader reconoce `.claude/rules/`
  (Vitalia 2→7 nodos) · reindex-tras-turno + `event: map` + refetch FE (Mapa en vivo) · tarjeta de
  identidad por sesión (el chat SABE qué arnés/copia edita; loop A4 operó solo: «Causa
  diagnosticada: instalación») · sesión de reparación rotulada + deriva re-evaluada · `ctxPct`
  REPARADO (medía acumulado: 100 % espurio → 14 % real) + rotación invisible probada
  (`ClaudeSessionID` rotó, checkpoint mecánico, conversación continua) · historial B2 (Close
  archiva metadata + lector JSONL nativo cosió 9 turnos de 2 JSONLs, endpoints + picker). Reparación
  REAL de Vitalia hecha por chat: sello con rol + skill `hipaa-check`. 3 bugs destapados por E2E y
  reparados (roleFor sin índice al crear sesión · ctxPct acumulado · Cwd sin estampar). 5
  capabilities nuevas (CAP-94..98), cobertura 100 %. Deuda honesta en `PARIDAD.md` §Estado global.
- Continuaciones abiertas (eje N2: derivación LIVE · validar ~40 `vivo·nc` · homologación 2° orden · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `264 checks · pass 51 · fail 0 · error 0 · deferred 213 · n/a 0` (medido 2026-07-24, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-24
- **arch/:** 21 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 101 — 57 vivo · 40 vivo·nc · 1 parcial · 3 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. **Cableado a CI + auto-cura local (HS-21):** `estado.sh --check` en el job `go`
> rompe el merge si estas cifras quedan stale (drift-gate); y el hook `pre-commit estado-cifras`
> las **regenera solas** en el commit (`--check`→regen→`git add`). Drift histórico ya corregido.
