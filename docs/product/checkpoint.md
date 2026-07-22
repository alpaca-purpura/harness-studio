# checkpoint — ArnesIA (presente)

> vig: activo · revisar: semanal. Las cifras se GENERAN con `arnesia conformance --todo` (D2/
> RF-178) — **no editar a mano**. El resto (fase, paquete activo) se toca al cerrar cada turno.

## Fase del gran plan

**Fase 5 (Implementación) EN CURSO.** Fases 1-4 ✓ (Visión · UX · Arquitectura · Specs).
Última ficha cerrada: **HS-25** (cierre de 4 gates humanos de PARIDAD — rebrand PRENTER · Portafolio
Slice 1-FE · Slice 2 sello/Identificar · shell-topbar-selector-arnes, 2026-07-22). Índice de historia →
`LEDGER.md` → `ledger/HS-NN.md`.

## Paquete de trabajo activo

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
- **Retomar aquí:** el operador señaló que **el chat (chat-cc-funcional, ya firmado HS-20) no está
  funcionando bien** y que no ve una historia de usuario para "arreglar un arnés conversando" — la
  historia SÍ existe (`stories/2026-07-08-chat-cc-funcional/INDEX.md`, "modificación de arneses
  existentes"), lo que falta es diagnóstico concreto. Investigación de esta sesión: el mecanismo de
  ALCANCE (con qué arnés/cwd abre el chat) recién se terminó de cerrar HOY (picker de
  shell-topbar-selector-arnes) y **GAP-1** (instalaciones `referenciada-cc` con CERO cwd registrado,
  documentado en Slice 1) puede hacer que chatear contra una instalación no escriba donde se espera.
  Ninguno de los dos ítems que el operador nombró (re-key `(home,id,scope)` · Fase 1 forja-ciclo-vivo
  item 3 "Chat forja/edita") es el bloqueo directo — re-key es plomería de identidad del Portafolio sin
  relación directa con el chat, y forja-ciclo-vivo item 3 es un flujo GUIADO distinto (init/doctor/
  loop-forward) pausado desde F-D6, no lo que hoy está roto. Próximo paso recomendado: diagnóstico en
  vivo del chat contra un arnés real (confirmar si el bug es GAP-1/cwd, el spike `control_response`
  pendiente, u otra cosa) antes de invertir en cualquiera de los dos backlogs grandes — ver conversación
  de esta sesión.
- Continuaciones abiertas (eje N2: derivación LIVE · validar ~40 `vivo·nc` · homologación 2° orden · deuda viva) → `docs/product/BACKLOG.md`.

## Cifras vivas

<!--stats: `scripts/estado.sh` regenera TODO este bloque desde conformance/árbol; no editar a mano -->
- **ruleset `--todo`:** `257 checks · pass 48 · fail 0 · error 0 · deferred 209 · n/a 0` (medido 2026-07-16, `go run ./cmd/arnesia conformance --todo`)
- **dogfood `--arnes`:** `21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0` (warn honesto `art-es-path`, el diente no se silencia) — medido 2026-07-16
- **arch/:** 18 boundaries (`codigo-traza-a-capability` **enforced**: R1/R2/R4 pasan)
- **docs/architecture/knowledge/:** 12 nodos · 138 checks
- **capabilities (SSoT):** 93 — 50 vivo · 39 vivo·nc · 1 parcial · 3 stub · **cobertura 100%** (0 huérfanos, 0 punteros colgantes)
<!--/stats-->

> Nota: `scripts/estado.sh` regenera **todo** el bloque desde conformance/árbol (RF-178 + HS-20):
> ruleset · `--arnes` · arch boundaries · knowledge nodos·checks · distribución capabilities — ya
> nada se teclea. **Cableado a CI + auto-cura local (HS-21):** `estado.sh --check` en el job `go`
> rompe el merge si estas cifras quedan stale (drift-gate); y el hook `pre-commit estado-cifras`
> las **regenera solas** en el commit (`--check`→regen→`git add`). Drift histórico ya corregido.
