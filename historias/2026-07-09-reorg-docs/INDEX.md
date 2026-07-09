# Reorg de docs — árbol de 3 ejes + hojas atómicas (paquete de trabajo)

> vig: activo · revisar: 2026-08-01
> Origen: el operador pide que `CLAUDE.md` deje de ser repositorio de backlog/pendientes y
> sea **puntero puro** (root del árbol); detalle en **hojas atómicas** con TTL, grep-ables
> desde un índice; separar por eje temporal (**lo-hecho ≠ lo-que-viene ≠ ahora**) para no
> cargar todo en memoria. Objetivo: máximo ahorro de tokens por sesión.
> Adaptación (METODOLOGIA §10): paquete meta-doc, CERO superficie UI — no aplica
> `mockup-*.html`/`design.md`. Flujo: auditoría → decisiones → spec + drafts → 🧑‍⚖️ firma
> → ejecución de la migración.

## Rumbo firmado

- **Alcance = diseño primero** (2026-07-09): se redacta la propuesta + drafts ANTES de tocar
  docs firmados. La migración se ejecuta tras firma. → `decisiones.md` D1.
- **Cifras vivas = generadas por conformance** (2026-07-09): las cuentas (checks/pass/
  deferred/boundaries) salen de `arnesia conformance`, nunca a mano. → D2. Hecho técnico
  que lo habilita: el comando YA emite la línea (`cmd/arnesia/conformance.go:142`).

## Entregables de este paquete (diseño)

- [x] `decisiones.md` — 4 firmas (alcance · cifras · 4º eje CAPABILITIES · unidad) + propuestas
- [x] `spec.md` — árbol de **4 ejes**, taxonomía de hojas, TTL, cifras generadas, RF-170..182
- [x] `CLAUDE-nuevo.md` — draft del router puro (~1.2k tok), para comparar lado a lado
- [x] `BACKLOG.md` — draft con los pendientes REALES de hoy (consolidados de LEDGER+INDEX)
- [x] `CAPABILITIES.md` — **SSoT funcional: 82 capabilities DERIVADOS DEL CÓDIGO** (barrido de
      7 subagentes sobre cmd/internal/web/src-tauri) con punteros `file#Símbolo` + estado honesto
- [x] `doctrina-capabilities.md` — doctrina de enforcement (R1..R4, hook, rollout warn→block)
- [x] `migracion.md` — qué línea/bloque va a qué hoja; orden de ejecución reversible
- [ ] 🧑‍⚖️ **gate: el operador firma el árbol + la doctrina** → recién ahí se ejecuta la migración
      y se construye el validador/hook (pasa de warn a block cuando cobertura=100%)

## Retomar aquí

- **MIGRACIÓN EJECUTADA (2026-07-09).** Firmado (árbol 4 ejes + doctrina + granularidad 82 +
  separar stack + punteros a doctrina/capabilities en el router). Landeado en raíz del repo:
  - `CLAUDE.md` → **router puro 835 tok** (−81% vs 4472) con punteros a CAPABILITIES · ESTADO ·
    BACKLOG · LEDGER · STACK · doctrina de desarrollo (§10) · doctrina de capabilities.
  - `CAPABILITIES.md` (raíz) = **SSoT funcional, 82 caps derivados del código**, punteros `file#Símbolo`.
  - `BACKLOG.md` · `ESTADO.md` (cifras reales `247·38·209`) · `STACK.md` (→ arch/, drift SQLite corregido).
  - `LEDGER.md` partido → `ledger/HS-01..18.md` (16 fichas + HS-18 de esta reorg); LEDGER = índice 23k.
  - `arch/boundaries/codigo-traza-a-capability.md` (`proposed`, 5 checks defer honesto, +5 al ruleset).
  - Verificado: `conformance --todo` 247 checks parsea sin romper, boundary defer (no pass fabricado).
- **Próximo paso (BACKLOG):** construir el validador `capability-trace` (R1/R2) + job lefthook +
  hook CC → graduar el boundary de `proposed` a `enforced` (warn→block). Cablear cifras generadas
  a `ESTADO.md` (RF-178). Validar los `sin-check`/STUB. Resolver dead-code candidatos.
- **Sin commitear aún** (working tree con cambios de otras sesiones — commit por archivo, nunca `-A`).
