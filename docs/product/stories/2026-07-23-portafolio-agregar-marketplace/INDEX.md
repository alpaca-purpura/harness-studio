# Portafolio · Agregar de marketplace → clonar + mejorar (paquete de trabajo)

> Ficha del programa «Portafolio · ciclo de vida del arnés» ([`BACKLOG.md`](../../BACKLOG.md) outcome,
> item 2). Origen: `bloqueo(1)` cae — Slice 0 «Cimientos» (HS-23) y Slice 1 «FE» (HS-25) ambos
> FIRMADOS 🧑‍⚖️. Fuente de specs a nivel de reglas (aún no promovidas a este paquete):
> [`../2026-07-10-spike-carga-arneses/spec-funcional.md`](../2026-07-10-spike-carga-arneses/spec-funcional.md)
> §7.2 + [`decisiones.md`](../2026-07-10-spike-carga-arneses/decisiones.md) S-D2/S-D6.

## Flujo y gates (METODOLOGIA §10)

1. **Mockup** — forkear el baseline vigente ([`mockups/INDEX.md`](../../../../mockups/INDEX.md)):
   el wizard del Portafolio ya construido en Slice 1 (`portafolio-wizard.stories.tsx`, rama
   «Proyecto») recibe la rama hermana «Marketplace» (git url → validar `marketplace.json` →
   listar → elegir → clonar a `<checkouts>/<home-slug>/<id>/`). Superset, no reinventar.
2. **Decisiones** (`decisiones.md`) — una entrada por decisión conversada, EN EL MISMO TURNO.
3. **Specs** (`spec.md` + `design.md`) — 🧑‍⚖️ firma del paquete.
4. **Implementación** contra el spec firmado.
5. **Paridad** (`PARIDAD.md`) — 🧑‍⚖️ gate final.

## Estado

- [ ] Mockup — no arrancado.
- [x] `decisiones.md` — PENDIENTE-01 **FIRMADA** (2026-07-23): registro extiende Portafolio ·
  matching manifiesto+hash · sin-match manual con sugerencia · extiende outcome 5.
- [ ] `spec.md` / `design.md` — no arrancados.
- [ ] Implementación — no arrancada (código NO se toca hasta specs firmados).
- [ ] `PARIDAD.md` — no arrancado.

## Retomar aquí

> Se actualiza al cierre de cada turno de trabajo (METODOLOGIA §10).

- **2026-07-23 — PENDIENTE-01 resuelta, desbloqueado el mockup.** Las 4 sub-preguntas (dónde vive
  el registro de marketplaces · mecanismo de matching · caso sin-match · relación con outcome 5)
  quedaron firmadas en `decisiones.md`. Modelo: registro extiende Portafolio/Slice 0 (lista
  `marketplaces_conocidos`, no entidad nueva) · matching manifiesto-primero+hash-fallback (reusa
  S0-D14) · sin-match siempre confirma el operador (anti-drift) · el flujo de reparar-sin-código
  extiende el outcome 5 ya en `BACKLOG.md`, no es un outcome nuevo.
- **Próximo paso concreto:** arrancar el mockup (punto 1 del flujo) forkeando el wizard de
  Slice 1 (`portafolio-wizard.stories.tsx`), rama «Marketplace», con este modelo como insumo.
