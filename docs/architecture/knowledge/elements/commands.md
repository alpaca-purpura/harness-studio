---
elemento: command
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/skills
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/commands
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://github.com/anthropics/claude-code/issues/17578
    autoridad: experto
    revisado: 2026-07-04
  - url: https://batsov.com/articles/2026/03/11/essential-claude-code-skills-and-commands/
    autoridad: experto
    revisado: 2026-07-04
---

# Slash commands (custom + built-in)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Archivo markdown que se vuelve invocación `/name` — built-in (lógica fija en
el CLI) o custom (plantilla de prompt). **Merge clave (v2.1.3):** custom commands y skills se
unificaron en un sistema — `.claude/commands/deploy.md` y `.claude/skills/deploy/SKILL.md`
producen ambos `/deploy` con el mismo comportamiento. `.claude/commands/` sigue por compat, pero
Anthropic recomienda migrar a `.claude/skills/` (archivos de apoyo, frontmatter rico, control de
invocación). El doc `/slash-commands` redirige a `/skills`. *(oficial: skills; ver [[skills]])*

**Ubicación (L1.2):** legacy `.claude/commands/<n>.md` → `/name`; actual
`.claude/skills/<n>/SKILL.md` → `/name`; plugin → `/plugin:name`. Colisión: enterprise > personal
> proyecto; **una skill gana a un command del mismo nombre**; skill puede pisar un built-in
homónimo. Subdirs anidados → nombre calificado (`/apps/web:deploy`). *(oficial)*

**Frontmatter (L1.3, todos opcionales, ver [[skills]]):** `argument-hint`, `arguments`,
`description`, `allowed-tools`, `disallowed-tools`, `model`, `disable-model-invocation`,
`user-invocable`, `context: fork`, `agent`, `hooks`, `paths`, `shell`. *(oficial)*

**Placeholders (L1.4):** `$ARGUMENTS` (todo; si ausente, se anexa `ARGUMENTS: <valor>`) ·
`$ARGUMENTS[N]`/`$N` (indexado 0-based) · `$name` (con `arguments: [issue,branch]`) · quoting
shell para multi-palabra · `\$` escapa literal · `${CLAUDE_SKILL_DIR}`/`${CLAUDE_PROJECT_DIR}`
(v2.1.196) · **stacking** `/code-review /fix-issue 123` (hasta 6, v2.1.199). **`` !`cmd` ``** corre
ANTES de que Claude vea el archivo (preprocesado, reemplaza inline; solo a inicio de línea/tras
espacio); fenced ` ```! ` multi-línea; apagar con `disableSkillShellExecution`. *(oficial)*

**Prácticas (L1.5):** procedimiento multi-paso → skill/command (no CLAUDE.md, que carga cada
turno) · subagente cuando ensuciaría el hilo, skill cuando quieres verlo inline, **hook** cuando
debe ser determinista · `disable-model-invocation: true` en command con efecto secundario
(`/commit`,`/deploy`) · declarar `argument-hint` · `allowed-tools` acotado parametrizado
(`Bash(git commit *)`, no bare `Bash`) · cuerpo <500 líneas · evaluar con baseline. Ejemplo de
scope estrecho: `/simplify` (solo limpieza) separado de `/code-review` (correctitud+limpieza). *(oficial + batsov)*

**Anti-patrones (L1.6):** command que debería ser skill (crece con scripts/plantillas) · sin
`argument-hint` cuando usa `$ARGUMENTS` · `allowed-tools` sobre-amplio · efecto secundario sin
`disable-model-invocation` · duplicar built-in (`/compact`,`/clear`) · YAML malformado (body carga
pero `description` vacía → nunca auto-matchea) · ruta hardcodeada en `!`inject`` en vez de
`${CLAUDE_SKILL_DIR}` · **inyección de shell** vía `` !`cmd $ARGUMENTS` `` sin quoting/validación. *(oficial)*

**Novedades / built-ins (L1.7):** el merge (v2.1.3) es el cambio estructural mayor (issue #17578:
docs aún con tabla «skills vs commands» contradictoria → gap de consistencia a trackear). **Set
built-in actual** incluye recientes clave: **`/goal`** (nuevo el último año — fija una condición
de persistencia que Claude persigue entre turnos; `clear`/`stop`/`off` para quitar) · `/fork`
(v2.1.161) · `/cd` (v2.1.169) · `/simplify` (v2.1.154) · `/verify`·`/run`·`/run-skill-generator`
(v2.1.145) · `/reload-skills` (v2.1.152) · `/dataviz` (v2.1.198) · `/code-review` con modo `ultra`
· `/effort` con `ultracode` · `/loop`·`/workflows`·`/deep-research`. **Removidos:** `/pr-comments`
(v2.1.91), `/vim` (v2.1.92). *(oficial: commands)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Tras el merge, para nosotros **un command custom ES una skill** — su estándar de estructura vive
en [[skills]]. Este nodo cubre lo específico de comando: **argumentos, invocación por el usuario,
y el uso de built-ins nuevos como palancas del proceso.**

1. **Command custom nace como skill:** en un arnés nuestro no se crea `.claude/commands/*.md`
   nuevo; se crea `.claude/skills/<n>/` (desbloquea contrato de caja si aplica). ⇐ L1.1/L1.2.
   Legacy conviviendo con skills = hallazgo de migración.
2. **Efecto secundario ⇒ `disable-model-invocation` + guardia:** `/commit`,`/deploy`,`/publish`
   solo por el usuario, nunca auto-invocados. ⇐ L1.5. Coherente con [[skills]] L2.4 y [[settings-permissions]].
3. **Built-ins nuevos como palancas del proceso, con lógica clara:** el operador pidió que
   comandos útiles nuevos (ej. `/goal`) se seteen «con lógica muy clara». Un arnés nuestro
   documenta CUÁNDO usa `/goal` (persistencia de objetivo por fase), `/verify` (gate de una caja),
   `/code-review ultra` (Calidad). ⇐ L1.7 + principio 1 (proceso implícito). *(esto es material
   de cadencia: cada built-in nuevo relevante se evalúa para adopción en el arnés estándar)*
4. **Argumentos declarados:** todo command con args lleva `argument-hint`; interpolar
   `$ARGUMENTS` a shell exige quoting. ⇐ L1.5/L1.6.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| cmd-argument-hint | usa `$ARGUMENTS`/`$N`/`$name` pero sin `argument-hint` | info | badge «comando sin argument-hint» | L1.5 |
| cmd-description | tiene `description` (no solo fallback de 1er párrafo) | warn | badge «sin description — auto-discovery débil» | L1.6 |
| cmd-side-effect-lock | verbos deploy/commit/push/delete sin `disable-model-invocation` | error | badge «efecto secundario invocable por el modelo» | L1.5 · L2.2 |
| cmd-tools-scoped | `allowed-tools` parametrizado, no bare `Bash`/`Bash(*)` | error | badge «allowed-tools sin acotar» | L1.5 |
| cmd-no-dup-builtin | nombre no choca con built-in (`compact`,`clear`,`model`…) | warn | badge «nombre choca con built-in» | L1.6 |
| cmd-migrate-to-skill | `.claude/commands/*.md` que se beneficia de skill (scripts/plantillas) | info | badge «candidato a migrar a .claude/skills/» | L1.1 · L2.1 |
| cmd-hardcoded-path | `` !`…` ``/script con ruta absoluta en vez de `${CLAUDE_SKILL_DIR}` | warn | badge «ruta hardcodeada, no portable» | L1.6 |
| cmd-shell-injection | `` !`cmd $ARGUMENTS` `` interpola args crudos sin quoting/validar | error | banda Guardia «posible inyección de shell vía \$ARGUMENTS» | L1.6 · L2.4 |
| cmd-body-size | cuerpo >500 líneas sin dividir | warn | badge «cuerpo >500 líneas» | L1.5 |
| cmd-builtin-usage-declared | built-ins de proceso (`/goal`,`/verify`,`/code-review`) con uso documentado en el arnés | info | badge «built-in nuevo sin lógica de uso declarada» | L2.3 |
| cmd-legacy-note | `.claude/commands/` legacy conviviendo con skills sin razón declarada | info | badge «commands/ legacy sin nota» | L1.7 · L2.1 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (skills, commands, anchor blog) +
  issue #17578 + batsov. **Registrado el merge command↔skill (v2.1.3): command custom = skill**
  → estructura en [[skills]]. L2 amarra built-ins nuevos (`/goal` et al.) como palancas del
  proceso con lógica declarada (pedido directo del operador). 11 checks. Set built-in actual +
  removidos (`/pr-comments`, `/vim`). · pasada fundacional.
