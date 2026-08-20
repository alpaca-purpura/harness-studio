# Harness backlog — carril L1 del CIL (proceso y herramientas)

> **Qué entra acá:** fricción del PROCESO y de las HERRAMIENTAS con las que trabajamos —
> el cockpit, los scripts, el kit, el flujo de release. Lo que es deuda del PRODUCTO
> (código y runtime de ArnesIA) va a [`tech-debt.md`](./tech-debt.md), carril L3.
>
> **Cómo se lee:** el cockpit lo parsea y lo pinta en las vistas Harness y CIL
> (`tools/cockpit/go/handlers_lifecycle.go::parseLifecycleTable`). El formato de la tabla
> **es el contrato**: 6 columnas, en este orden, y el `id` tiene que matchear `HB-\d+`.
>
> **Severidad:** 🔴 silent-killer (falla sin avisar) · 🟡 quick-win · 🔵 decision (necesita
> que alguien decida) · 🟣 wave (ola de trabajo).
>
> **Estado:** `reported` → `triaged` → `ratified` → `applied` → `verified`, más `deferred`
> para lo que se posterga a propósito. Un ítem que nadie miró se queda en `reported` — eso
> es información, no un pendiente de higiene.

| id | fecha | sev | item | estado | ref |
|---|---|---|---|---|---|
| HB-1 | 2026-08-19 | 🟡 | El cockpit borra los comentarios del frontmatter al transicionar una story: `stringifyFrontmatter` re-serializa con yaml.v3, que no los preserva. La línea de evidencia que siembra `backfill_checkpoints.py` desaparece en el primer drag; el marcador `backfilled: true` sí sobrevive | reported | `tools/cockpit/go/frontmatter.go` |
| HB-2 | 2026-08-19 | 🔵 | Los carriles del CIL no existían en este repo, así que las vistas Harness y Learnings del cockpit estaban apagadas. Este archivo y `tech-debt.md` los abren | applied | `stories/2026-08-18-cockpit-y-doctrina/decisiones.md` § CD-8 |
| HB-3 | 2026-08-19 | 🟡 | El primer render del board pide `?sistema=main` (el `FALLBACK_SISTEMA` del provider) y deja tres 400 en la consola antes de resolver `platform`. Ruidoso, sin efecto funcional | reported | `tools/cockpit/ui/components/providers/SistemaProvider.tsx` |
| HB-4 | 2026-08-19 | 🔵 | El cockpit vendored puede driftear del upstream de vitalia-app. Mitigado con la tabla de diffs y `vendored_arnesia_test.go`, que falla si un merge borra una adaptación — pero nadie chequea el upstream periódicamente | ratified | `tools/cockpit/README-vendored.md` |
| HB-5 | 2026-08-13 | 🟣 | Suite Go COMPLETA en verde en Windows: faltan shims `gh`/`git` sh, chmod/bits como fault-injection, JSON que embebe backslash, literales `/` en domain/usecase/portafolio/stt | triaged | `stories/2026-08-13-compilacion-windows/decisiones.md` § CW-D5 |
| HB-6 | 2026-08-13 | 🔵 | `bump.py` portable. `bump.sh` ya corre en Git Bash con el shim de python, así que el bump YA funciona en Windows; el `.py` quedaría solo para no depender de bash | deferred | `docs/product/BACKLOG.md` § Port Windows |
| HB-7 | 2026-08-13 | 🔴 | `go-arch-lint` reporta violaciones FALSAS con paths backslash y `estado.sh` no termina en Windows. Hoy se saltean con skip honesto, pero eso significa que las cifras del repo solo se regeneran en Linux/CI | triaged | `docs/product/BACKLOG.md` § Port Windows |
| HB-8 | 2026-08-19 | 🟡 | `TestPresupuestoDeBinario` agota su timeout de 600 s en Windows: compila desde `git archive` en un dir sin caché de build. El presupuesto se midió a mano replicando su método; el test sigue siendo el juez y corre en CI | reported | `docs/architecture/fitness/telemetria_test.go` |
| HB-9 | 2026-08-19 | 🔵 | El provisioning de `~/.arnesia` es PEREZOSO: se dispara al spawnear una sesión, no al arrancar el daemon. Tras instalar una versión nueva, el kit en disco queda viejo hasta el primer turno de chat. Se auto-repara por huella, pero nadie lo dice | reported | `internal/usecase/session_service.go` |
| HB-10 | 2026-08-19 | 🟡 | El contador L2 del CIL dice 4 y la vista Learnings lista 3: `countMarkdownFiles` cuenta todos los `.md` incluido el `README.md` que `listLearningsV2` excluye a propósito. Cifra que no coincide con lo que se ve. Se registra en vez de parchear para no driftear del upstream por algo cosmético | reported | `tools/cockpit/go/handlers_lifecycle.go::handleCIL` |
| HB-11 | 2026-08-19 | 🔵 | El eval de `forjar-arnes` mide una caja **T2** (multi-paso interactivo) con sesiones `claude -p` de UN turno, así que los prompts tienen que entregar todo cerrado: sin `reporta_a` la sesión se frena a preguntarlo, que es lo correcto. Medir el grill en sí pide un runner multi-turno con respuestas guionadas | reported | `evals/README.md` § Los prompts entregan TODO cerrado |
| HB-12 | 2026-08-19 | 🔴 | **El eval encontró que `forjar-arnes` no decía la FORMA del manifiesto**: la sesión escribía `fases` como objetos, `marketplace` como objeto y un `fase` extra en cada transición — 9 errores de schema, todos por adivinar la estructura. Corregido con un esqueleto JSON exacto (validado contra `graph.l0.schema.json`) + los 3 errores típicos nombrados | applied | `kit/skills/forjar-arnes/SKILL.md` paso 3 |
| HB-13 | 2026-08-19 | 🟡 | El eval encontró que la forja se frenaba a pedir autorización para `.claude/settings.json`: la skill no decía que un arnés nace en **forma-plugin** (`hooks/hooks.json`), y esa ruta está protegida por Claude Code. Corregido declarando el layout completo | applied | `kit/skills/forjar-arnes/SKILL.md` paso 4 |
