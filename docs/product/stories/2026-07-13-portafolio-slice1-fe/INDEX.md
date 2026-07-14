# Portafolio · Slice 1 «FE Portafolio» — paquete de build

> `tipo: build` (FE sobre los cimientos FIRMADOS de Slice 0, HS-23). Deriva del spike
> [`../2026-07-10-spike-carga-arneses/`](../2026-07-10-spike-carga-arneses/INDEX.md) (spec §10 +
> casuística §I + revisión adversaria G1-G9) + BACKLOG «Programa Portafolio» item 1 (DESBLOQUEADO).
> Roles: **Fable 5 = arquitecto** (plan + decisiones técnicas, 2026-07-13) · **Sonnet 5 = constructor**
> (ejecuta tickets T1-T8) · **operador = firma** (gate PARIDAD).

## Qué se construye (1 línea)

Las 3 superficies del Portafolio (Lista/lente-empresa · Wizard-Proyecto/carpeta-local · Drawer READ)
cableadas al HTTP REAL de Slice 0, con defaults honestos (deriva-no-evaluable · update no-verificado ·
marketplace disabled), los fixes G1-G9 uno a uno, «Abrir en Mapa» (observación read-only, reusa
CAP-58/61) + Desvincular con confirmación — todo porteado a Storybook (SSoT, story = test, a11y = test).

## Archivos

- [`plan-implementacion.md`](./plan-implementacion.md) — **EL plan**: goal verificable (§0), diseño
  técnico con contratos exactos Go+TS (§2), tickets T1-T8 con tests/stories nombrados y mapa G1-G9→ticket
  (§3), gate local (§M), prohibiciones (§P).
- [`decisiones.md`](./decisiones.md) — S1-D1..D15: decisiones del arquitecto que cierran lo que la spec
  dejó abierto para el FE (endpoint de observación por el gap `checkProtected` · colisión bare-id visible ·
  `Registries` poblado · regla de salud G9 · placement FSD · proyecto vitest `unit` · toolbar mínima ·
  tooltips S5 · frescura del dato) + **3 gaps reales** encontrados verificando contra el código de Slice 0.
- `paridad.md` — la escribe el constructor al cerrar T8: tabla spec §10 + G1-G9 ↔ realidad + evidencia
  E2E viva (browser + curl contra el daemon real). **NO existe aún.**

## Retomar aquí

> **Estado (2026-07-13): PLAN ESCRITO, build NO arrancado.** El arquitecto (Fable 5) dejó plan +
> decisiones listos; el gate de Slice 0 está firmado (HS-23) así que este slice está DESBLOQUEADO.
> **Siguiente paso:** (1) el operador revisa `decisiones.md` — en particular los 3 gaps flaggeados
> (registro-protegido → S1-D1 · colisión bare-id → S1-D2 · `Registries` sin poblar → S1-D3) — y da la
> orden de build; (2) Sonnet 5 ejecuta T1→T8 EN ORDEN, un commit por ticket, gate §M verde antes de cada
> commit, desviaciones documentadas como S1-D nuevas EN EL MISMO TURNO; (3) al cerrar T8: `paridad.md`
> con evidencia real → gate humano 🧑‍⚖️ del operador (la firma NO se simula).
