# Botón «Correr» de una caja — Inspector (paquete de trabajo)

> Ficha HS-11 (deuda viva, barrido 2026-07-23) · Origen: backend async YA vivo
> (`POST …/boxes/{id}/run` → 202+run_id · `GET …/runs/{runId}` · commit `f7a48f6`, `CAP-55`);
> `inspector.tsx` («Corridas donde actuó») solo tiene prosa placeholder, sin botón real.
> Disciplina METODOLOGIA §10: mockup → decisiones → spec → 🧑‍⚖️ → código → PARIDAD.

## Rumbo firmado (2026-07-23, respuestas del operador)

Paquete **completo** (no ticket chico) · mecanismo **SSE `event: run`** (resync vía GET solo si
hace falta) · error 409 **inline en la Section**, sin toast. Detalle en `decisiones.md` (D1-D4).

## Flujo y gates

1. **Mockup** (`mockup-boton-correr.html`, publicado como Artifact) — 6 casuísticas de la tab
   Corridas sobre tokens PRENTER reales + clases `inspector.css` verbatim. Iterar → 🧑‍⚖️ firma.
2. **Decisiones** (`decisiones.md`) — D1-D4, ya escritas.
3. **Spec** (`spec.md`) → 🧑‍⚖️ firma del paquete → implementación autorizada.
4. **Implementación** — tipos TS + client methods + listener SSE `run` + botón real en
   `Corridas()` + stories + capability FE `docs/product/capabilities/fe-mapa/`.
5. **PARIDAD** (`PARIDAD.md`) → gate final.

## Estado

- [x] investigación real del backend + FE existente (Explore, 2026-07-23)
- [x] rumbo firmado (alcance · mecanismo · error 409)
- [x] mockup publicado — Artifact: https://claude.ai/code/artifact/85ab09b6-ee47-4856-9ac8-f0be6bd4f235
- [ ] 🧑‍⚖️ firma del mockup — **PAUSADO 2026-07-23** (ver abajo)
- [ ] spec.md
- [ ] implementación + stories + tests
- [ ] capability FE registrada
- [ ] PARIDAD

## Retomar aquí

**PAUSADO a pedido del operador (2026-07-23)**, no descartado. El operador preguntó qué hace
funcionalmente el botón antes de firmar («explicame para qué sirve, revisa la visión, cómo
ayuda a crear/editar/mejorar arneses») — se le dio la explicación completa (grounded en
vision.md A1-A3 «caja = un paso del proceso, dueña de una transición» + `run_service.go` real:
Correr invoca el conductor T3, Claude Code headless corriendo DE VERDAD contra el repo real del
arnés, con rol/permisos derivados de la META, no una simulación). Tras leer la explicación, el
operador eligió **pausar este ítem** y seguir con el resto del barrido de deuda viva, sin pedir
cambios de alcance ni rechazar el mockup. Retomar: releer la explicación de arriba (o pedir la
misma a la sesión siguiente) y decidir firma/ajuste/descarte antes de tocar `spec.md`. Nada de
lo construido (mockup + decisiones D1-D4) se invalida por la pausa.
