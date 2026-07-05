# CADENCE — cómo vive el árbol de conocimiento

> **El conocimiento no es un `.md` que se lee una vez; es un árbol que crece.** Este
> documento define el mecanismo: cómo entra un hallazgo nuevo, cómo se versiona un nodo, cómo
> se mantiene la disciplina de dos capas y cómo el árbol se convierte en el ruleset que
> ArnesIA carga para pintar puntos de mejora en el mapa. Norte: [`../VISION.md`](../VISION.md)
> (principio 8 «estándar propio», principio 11 «economía de contexto medible») y
> [`../METODOLOGIA.md`](../METODOLOGIA.md) (doctrina de negocio, que **deriva** de este árbol).

## Por qué existe

El arnés importa casi tanto como el modelo, y la superficie que lo dirige (skills, hooks,
rules, subagents, commands, MCP, plugins, settings, output-styles, statusline, headless)
**cambia cada semana**: aparecen comandos nuevos (hace un año no existía `/goal`), campos de
frontmatter nuevos, eventos de hook nuevos, features que deprecan patrones viejos. Un estándar
congelado envejece en días. Por eso el estándar de ArnesIA vive como **árbol versionado y
aditivo**, no como doctrina de piedra.

Este árbol es la **capa 1** (qué recomiendan Anthropic y los mejores). La **capa 2** (nuestra
forma de trabajo) vive DENTRO de cada nodo y **está obligada a derivar de la capa 1**. Regla
dura: si nuestra forma de hacer las cosas no es la recomendada por los expertos, o la
corregimos o la marcamos como divergencia justificada — nunca se queda como divergencia muda.

## Anatomía de un nodo (`elements/<elemento>.md`)

Cada elemento de la superficie de Claude Code = un nodo = un archivo. Estructura obligatoria:

```markdown
---
elemento: <skill|hook|rule|subagent|command|mcp|plugin|settings|output-style|statusline|headless>
version: <n.m>              # sube con cada pasada que cambia el nodo (semver-lite)
updated: <YYYY-MM-DD>       # fecha de la última pasada
status: vivo | estable | en-revisión
fuentes:                    # las fuentes vivas que se re-chequean en cada cadencia
  - url: <...>
    autoridad: oficial | estándar-abierto | experto
    revisado: <YYYY-MM-DD>
---

## L1 · Estándar (oficial Anthropic + expertos)
Qué dicen los docs oficiales y los mejores de internet. Cada afirmación fechada y con
fuente. Es EVIDENCIA, no opinión nuestra. Append-only en espíritu: lo viejo se marca
`(deprecado <fecha>)`, no se borra.

## L2 · Nuestra adaptación (paradigma alpacapurpura)
Nuestras reglas para este elemento. Cada regla **cita el punto de L1 del que deriva**
(`⇐ L1: <ancla>`). Si una regla nuestra va más allá o contra L1, se marca:
- `⚠ divergencia` + justificación explícita, o
- se corrige para alinearse.
Aquí es donde el estándar genérico se moldea a nuestra fábrica de cajas (contratos,
eval-gates, rol×proceso).

## Checklist evaluable
La rúbrica as code: lo que un linter corre contra un componente real. Tabla:
| id | qué chequea | severidad | señal en el mapa | deriva de |
Cada check es **objetivamente verificable** desde el archivo o la telemetría (no vago).
Severidad: `error` (rompe el estándar) · `warn` (huele mal) · `info` (mejora posible).
La columna «señal en el mapa» es el puente a UX: qué badge/estado pinta ArnesIA.

## Changelog
Anillos de crecimiento. Una línea por pasada:
`YYYY-MM-DD · vX.Y · qué se añadió/cambió/deprecó · quién/qué fuente lo disparó`
```

## El ritual (cadencia semanal)

> **Ejecutable:** el procedimiento paso a paso vive en [`WEEKLY-UPDATE.md`](./WEEKLY-UPDATE.md).
> Se dispara de dos formas — **routine programado** (cron cloud, semanal) o el **botón
> «Actualizar estándar»** del mapa (manual, on-demand). Tier vigente: **HÍBRIDO** (auto-aplica
> solo evidencia L1; L2/checks/elementos nuevos van a propuestas en `_inbox/`).

1. **Barrido de novedades.** Research de qué salió nuevo (changelog de Claude Code, docs
   oficiales, blog de Anthropic, releases del kit, chatter de expertos). Un frente por
   familia de elementos; se puede automatizar con agentes en paralelo (como la pasada
   fundacional que creó este árbol).
2. **Triage.** Cada hallazgo cae en un nodo existente o crea uno nuevo (elemento nuevo de la
   superficie de Claude Code).
3. **Append a L1.** Se añade la afirmación fechada + fuente. Lo que quedó obsoleto se marca
   `(deprecado <fecha>)` — **no se borra** (el árbol conserva su memoria; principio 4 «aditivo
   y sin pérdida»).
4. **Revisión de L2.** ¿Sigue alineada nuestra forma con el L1 actualizado? Si un cambio de
   Anthropic vuelve nuestra regla obsoleta o divergente → corregir o re-justificar.
5. **Evolución de checks.** Traducir lo nuevo a checks evaluables (o retirar checks que ya no
   apliquen). Todo check nuevo declara su señal en el mapa.
6. **Bump + changelog.** Subir `version` y `updated` del nodo, registrar el anillo.
7. **Propagar.** Si el cambio toca reglas de negocio, actualizar [`../METODOLOGIA.md`](../METODOLOGIA.md);
   si toca la UX (un check nuevo que el mapa debe pintar), registrarlo en [`../UX.md`](../UX.md).

## Reglas del árbol (no re-negociar)

- **Aditivo, no reescritura.** Un nodo crece; lo viejo se marca deprecado, no se elimina.
- **Toda afirmación de L1 lleva fuente + fecha.** Sin fuente, no es estándar — es rumor.
- **L2 deriva de L1.** Divergencia = marcada y justificada, jamás silenciosa.
- **Todo check es objetivamente verificable** desde el archivo del componente o su telemetría.
- **Prioridad de fuentes:** (1) docs oficiales Anthropic (code.claude.com/docs,
  docs.claude.com) · (2) estándares abiertos adoptados (agentskills.io, AGENTS.md,
  marketplace.json) · (3) expertos reputados. La opinión de un experto no pisa a la doc
  oficial; la complementa.
- **Honestidad (heredada de METODOLOGIA §4):** si no verificamos algo, dice «unverified»;
  no se inventa una recomendación para rellenar.

## Hacia dónde va (el árbol ES producto)

Hoy este árbol es markdown as code en el repo. Mañana es el **ruleset que el motor de ArnesIA
carga**: los `Checklist evaluable` de cada nodo se vuelven las reglas del linter que corre
sobre cada arnés importado (METODOLOGIA §6 «conformación») y pinta puntos de mejora en el mapa
(capa Desempeño / Diagnóstico). Es la semilla concreta del «segundo cerebro / estándar
propio» (VISION debate abierto #1). Por eso la disciplina importa: cada check mal puesto hoy
es un falso positivo en el mapa mañana.
