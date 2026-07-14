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

> **Estado (2026-07-14): BUILD COMPLETO T1-T8, gate humano 🧑‍⚖️ PENDIENTE.** T1-T7 commiteados a
> `main` (`7d27a7d`,`cefccb7`,`6ed3fa8`,`d4c9fd6`,`f8fe840`,`7413390`,`dfa82b5`). T8 (capabilities
> completas as-code · mockup `arnesia-portafolio.html` corregido G1-G9 + `mockups/INDEX.md`
> re-estampado · `paridad.md` con E2E vivo real · cierre documental) ejecutado en esta sesión y
> **NO commiteado** por instrucción explícita del orquestador («NO COMMITEES NI HAGAS PUSH») — queda en
> el working tree para que el operador lo commitee junto a (o después de) la firma.
> **Siguiente paso:** el operador lee [`paridad.md`](./paridad.md) completa (goal §0 · tabla G1-G9 ·
> E2E vivo · 22 decisiones S1-D · 8 desviaciones) y firma el gate humano (checkbox de la sección
> `## Firma`, HOY sin marcar) — la firma NO se simula acá. Al firmar: commitear T8 + actualizar
> `checkpoint.md`/`BACKLOG.md` (ítem 1 → cerrado) + `ledger/HS-24.md`.
