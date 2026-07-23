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
- [x] `decisiones.md` abierto con la primera nota del operador (pendiente de resolver, NO firmada).
- [ ] `spec.md` / `design.md` — no arrancados.
- [ ] Implementación — no arrancada (código NO se toca hasta specs firmados).
- [ ] `PARIDAD.md` — no arrancado.

## Retomar aquí

> Se actualiza al cierre de cada turno de trabajo (METODOLOGIA §10).

- **2026-07-23 — paquete creado, sin resolver nada todavía.** El operador dejó UNA nota abierta
  en `decisiones.md` (entrada «PENDIENTE-01») para retomarla en una conversación nueva, antes de
  tocar mockup: cómo ArnesIA reconcilia un arnés YA instalado dentro de un **proyecto** cargado
  (no un arnés suelto) contra su origen real de marketplace — matching, versión/desactualización,
  y el caso sin match (crear slot de mapeo a un marketplace elegido por el operador). Esto
  encadena con el objetivo de «reparar» sin pasar código a mano: el operador solo entrega
  acceso a plugin+marketplace y el usuario final instala/repara solo.
- **Próximo paso concreto:** abrir esa conversación, resolver PENDIENTE-01 (puede tocar el
  modelo de datos de Slice 0/`domain.Arnes` — evaluar si es scope de ESTE paquete o si destapa
  deuda propia de Portafolio), recién después arrancar el mockup del punto 1 del flujo.
