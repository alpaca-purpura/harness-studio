---
elemento: rule
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/memory
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/features-overview
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/context-window
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://agents.md/
    autoridad: estándar-abierto
    revisado: 2026-07-04
  - url: https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
    autoridad: oficial
    revisado: 2026-07-04
---

# Rules / memoria (CLAUDE.md · AGENTS.md · .claude/rules · imports)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Capa de instrucción **always-on**: markdown cargado en cada sesión para que
Claude lleve conocimiento persistente del proyecto/persona/org sin re-decirlo. **Es contexto,
no configuración enforced:** «para bloquear una acción pase lo que pase, usa un PreToolUse hook,
no CLAUDE.md». *(oficial: code.claude.com/docs/en/memory)*

**Jerarquía y carga (L1.2, de amplio a específico, todos aditivos):** managed policy > usuario
`~/.claude/CLAUDE.md` > proyecto `./CLAUDE.md` o `./.claude/CLAUDE.md` > local `./CLAUDE.local.md`
(gitignore). En conflicto, «la instrucción más específica suele ganar» (juicio del modelo, no
override duro). Walk de directorio root→cwd; subdirs cargan lazy al leer archivos ahí. *(oficial)*

**`@path` imports (L1.3):** relativo al archivo importador, recursivo **máx 4 hops**, salta
code spans (backticks = mención sin importar), 1ª vez pide aprobación. Los imports igual entran
al contexto — organizan, **no reducen tokens**. *(oficial)*

**`.claude/rules/` (L1.4):** un tema por archivo, recursivo; frontmatter `paths: [...]` scopea
la regla a cargar **solo al leer archivos que matchean** (no cada turno); sin scope carga al
launch con prioridad de `.claude/CLAUDE.md`. Symlinks para compartir cross-repo. *(oficial)*

**Economía de contexto (L1.5, central):** objetivo **<200 líneas** por CLAUDE.md («más largo
consume más contexto y reduce adherencia»). Cada token de regla se paga **cada turno**. El
`/context` cuantifica el costo (ej. CLAUDE.md ~1.800 tok). Post-compaction: CLAUDE.md root +
rules sin scope se re-inyectan; rules con `paths` y CLAUDE.md de subdir NO (recargan al leer).
El giro de context-engineering: preferir retrieval just-in-time (grep/glob/skills) al pre-load.
*(oficial: memory + context-window + engineering blog)*

**Prácticas (L1.6):** específico > verboso («2-space indent» > «formatea bien»; «run npm test
before commit» > «testea») · headers/bullets · **sin contradicciones** (Claude elige arbitrario)
· `file:line` > pegar código (se pone stale) · no re-decir lo que el linter ya enforca · crecer
por trigger (añadir cuando el mismo error pasa 2ª vez), no upfront · revisar como código (owner
+ PR) · procedimiento multi-paso → skill, no CLAUDE.md · constraint de un tipo de archivo →
`paths:` rule. *(oficial + maketocreate)*

**AGENTS.md (L1.7):** Claude Code **no** lo lee nativo. Patrón: CLAUDE.md con `@AGENTS.md` +
notas Claude-específicas, o symlink. AGENTS.md donado a la Agentic AI Foundation (Linux
Foundation, 2025-12), >60k proyectos, punto de convergencia vendor-neutral. *(oficial + agents.md)*

**Novedades (L1.8):** **auto memory** (v2.1.59+) `~/.claude/projects/<p>/memory/MEMORY.md`
(Claude-authored, ≤200 líneas/25KB, cross-worktree) · `.claude/rules/ paths:` first-class ·
`claudeMdExcludes` (monorepo) · `claudeMd` key en managed-settings (piso org) · hook
`InstructionsLoaded` (auditar qué cargó) · el viejo `#` quick-add regresó/roto (v2.0.74), ya no
documentado. *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

Las reglas viven en la **banda Base** del mapa (VISION A6: infraestructura compartida, no una
caja). Son «siempre en contexto» y su métrica = sesiones que las cargan × su costo (ya modelado
en it.6 como 6º tipo de primera clase).

1. **Presupuesto de contexto medido y visible:** el costo always-on de CLAUDE.md + rules sin
   scope se mide y se pinta en banda Base; superar el techo = hallazgo. ⇐ L1.5 + principio 11.
   *(esto ES el principio 11 «economía de contexto medible» hecho check)*
2. **Regla dura vs advisory:** «nunca hagas X» sobre acción destructiva/secreto **exige un hook
   PreToolUse** que la enforce; sin él es solo advisory → hallazgo de Guardia. ⇐ L1.1/L1.6.
   Conecta [[hooks]] (la regla la enuncia, el hook la garantiza).
3. **Scope por relevancia:** guía específica de un lenguaje/dir va a `paths:` rule, no al
   CLAUDE.md global (no pagar tax always-on por lo irrelevante). ⇐ L1.4/L1.5.
4. **Sin contradicciones cross-capa:** reglas de distintos niveles que se contradicen = error
   (el modelo elige arbitrario). ⇐ L1.6.
5. **AGENTS.md por import/symlink, nunca duplicado:** un arnés nuestro que convive con otras
   tools usa `@AGENTS.md`, no copia. ⇐ L1.7 (VISION ya adoptó AGENTS.md como estándar ganador).

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| rules-size-budget | CLAUDE.md + rules sin scope <~200 líneas | warn | banda Base «reglas caras: Xk siempre en contexto» | L1.5 · L2.1 |
| rules-token-budget | total always-on (managed+user+proyecto+local+rules) bajo el techo | warn | banda Base con conteo tok vs presupuesto | L1.5 · L2.1 |
| rules-contradiction | dos reglas cargadas se contradicen (indent, comando de test) | error | flag par de archivos + líneas | L1.6 · L2.4 |
| rules-pasted-code | fences de código >N líneas en vez de `file:line` | warn | conteo de fences grandes | L1.6 |
| rules-path-scoping | guía de un dir/lenguaje en CLAUDE.md global en vez de `paths:` rule | info | candidato a extraer + ahorro estimado | L1.4 · L2.3 |
| rules-hard-enforcement | «nunca/siempre» sobre acción destructiva/secreto sin hook PreToolUse | warn | banda Base «regla solo advisory, no aplicada» | L1.1 · L2.2 |
| rules-agents-dup | AGENTS.md y CLAUDE.md con contenido solapado sin `@import`/symlink | warn | score de similitud entre archivos | L1.7 · L2.5 |
| rules-import-integrity | `@path` resuelve, depth ≤4, sin ciclo | error | import roto / ciclo / depth | L1.3 |
| rules-local-gitignore | `CLAUDE.local.md` existe pero no está en `.gitignore` | warn | estado git del archivo | L1.6 |
| rules-staleness | referencia comandos/paths/versiones ausentes del repo | warn | conteo de referencias muertas | L1.6 |
| rules-procedural-leak | pasos multi-step («primero…luego…») mejor como skill | info | conteo de bloques secuenciales fuera de skills | L1.6 · L2 |
| rules-managed-gap | org sin managed `claudeMd`/`allowManagedPermissionRulesOnly` mientras cada proyecto puede aflojar | info | overlap managed vs proyecto | L1.8 |
| context-injection-native | reglas-negocio/conocimiento estático entran por `CLAUDE.md` / rules `paths:` / `@import` / SessionStart hook — **nunca** por `persistent_facts`/`activation_steps_prepend` (BMAD-ismos que CC ignora) | warn | banda Base «conocimiento inyectado por clave fantasma — no carga» | doctrina v1 §8.6 · L1.4 |

## Changelog

- 2026-07-05 · v1.1 · **Doctrina v1 (HS-07):** +1 check `context-injection-native` — el conocimiento
  estático (reglas de negocio, MOF del cliente) entra por primitivas CC nativas, no por los globs
  `persistent_facts` de BMAD que CC ignora en silencio. 12→13 checks. · barrido externo (VISION §Linaje).
- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (memory, features-overview,
  context-window, engineering context post) + agents.md (Linux Foundation). L2 amarra reglas a
  banda Base + principio 11 (presupuesto medido) + frontera regla-advisory/hook-enforced.
  12 checks. Novedades: auto memory v2.1.59, `.claude/rules/ paths:`, `claudeMdExcludes`,
  managed `claudeMd`, hook InstructionsLoaded, `#` quick-add roto. · pasada fundacional.
