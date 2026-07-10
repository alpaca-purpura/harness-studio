---
elemento: output-style
version: 1.0
updated: 2026-07-04
status: vivo
fuentes:
  - url: https://code.claude.com/docs/en/output-styles
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://code.claude.com/docs/en/prompt-caching
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://claude.com/blog/steering-claude-code-skills-hooks-rules-subagents-and-more
    autoridad: oficial
    revisado: 2026-07-04
  - url: https://github.com/anthropics/claude-code/issues/6450
    autoridad: experto
    revisado: 2026-07-04
---

# Output styles (estilos de system prompt)

## L1 · Estándar (oficial Anthropic + expertos)

**Qué es (L1.1).** Archivos Markdown que inyectan instrucciones **directamente en el system
prompt** de Claude Code — cambian rol/tono/formato, no lo que sabe del proyecto. Un custom style
puede **reemplazar** las instrucciones SWE built-in enteras (convertir el agente de coder a
asistente de escritura) salvo que se le diga mantenerlas. Aplican a cada turno de la sesión, se
fijan al inicio (**sin switch mid-sesión**). *(oficial: output-styles)*

**Ubicación y frontmatter (L1.2):** `~/.claude/output-styles`, `.claude/output-styles`, managed,
plugin `output-styles/`. Campos: `name`, `description` (picker de `/config`),
`keep-coding-instructions` (default `false` — mantiene las SWE built-in),
`force-for-plugin` (plugin auto-aplica su estilo). Built-in: **Default · Proactive ·
Explanatory · Learning** (el fast.io con «Concise/Technical» NO casa docs → no verificado). *(oficial)*

**Mecanismo (L1.3):** se anexa al final del system prompt + recordatorios en conversación.
Selección por `/config` → persiste en `.claude/settings.local.json`. El comando standalone
`/output-style` **deprecado v2.1.73, removido v2.1.91** → usar `/config`/`outputStyle`. Cambio
aplica tras `/clear` o nueva sesión (interactúa con prompt caching). *(oficial: output-styles + prompt-caching)*

**Jerarquía de peso (L1.4, del anchor blog):** los output styles cargan el **mayor peso de
instrucción** de toda la superficie porque viven en el system prompt y **nunca se compactan**
(styles > CLAUDE.md/rules > skills > hooks-external). Para adiciones one-off preferir
`--append-system-prompt` (añade sin remover defaults). *(oficial: anchor blog)*

**Prácticas (L1.5):** proyecto/convenciones → CLAUDE.md, NO output style · `keep-coding-instructions:
true` cuando cambias el CÓMO comunica pero sigue haciendo ingeniería (omitir solo si no hace SWE
del todo) · revisar built-in antes de crear custom · especificar **estructura, no frases
literales** (frágil entre versiones) · subagentes **NO heredan** el output style → codificar
tono/formato en la definición del subagente. *(oficial + comunidad)*

**Límite conocido (L1.6, clave):** los estilos son **steering probabilístico, no enforcement**.
Issue #6450: un estilo «Professional» (sin emoji/euforia) fue pisado consistentemente por el
entrenamiento base; Anthropic: los patrones base «pisan tus instrucciones de estilo porque son
más fundamentales». Cerrado «not planned» → **la supresión de tono no está garantizada** sin
respaldo de hook. *(experto: issue #6450)*

**Novedades (L1.7):** superficie pequeña; churn = consolidación de comando (`/output-style`
removido v2.1.91 → `/config`) · resolución de dir anidado (closest-to-cwd gana, v2.1.178) ·
plugins con `force-for-plugin`. *(oficial)*

## L2 · Nuestra adaptación (paradigma alpacapurpura)

El output style es la superficie **menos usada** en un arnés nuestro por diseño: el rol×proceso
se encarna en skills/rules/hooks (VISION principio 1), no en persona. Cuando existe, tiene que
justificarse.

1. **Convenciones nunca en el estilo:** política de repo, comandos de validación, hechos del
   proyecto van a CLAUDE.md ([[rules]]), no al output style. ⇐ L1.5. Check de solape.
2. **Coding harness mantiene sus instrucciones SWE:** un arnés de dev con custom style declara
   `keep-coding-instructions: true` o pierde scoping/verificación. ⇐ L1.5.
3. **Estilo no contradice reglas:** persona/autonomía del estilo no puede chocar con banda Base
   («actúa libre» vs «siempre confirma destructivo»). ⇐ L1.5 (cross-capa con [[rules]]/[[settings-permissions]]).
4. **Supresión de tono ⇒ respaldar con hook:** si el estilo intenta forzar tono (sin emoji,
   sin euforia), sabemos que no está garantizado → si importa, un [[hooks]] lo enforca. ⇐ L1.6.
5. **Subagentes llevan su propio contrato:** no asumir herencia de estilo en la maquinaria. ⇐ L1.5 + [[subagents]].

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | deriva de |
|----|-------------|-----------|------------------|-----------|
| style-missing-desc | frontmatter sin `description` (queda sin etiqueta en picker) | info | badge «sin descripción en picker» | L1.2 |
| style-loses-coding | body claramente de coding pero `keep-coding-instructions` unset/false | warn | badge «pierde verificación/scoping de código» | L1.5 · L2.2 |
| style-duplicates-rules | body con política de repo (comandos, ownership) que va en CLAUDE.md | warn | badge «política de repo en style — mover a CLAUDE.md» | L1.5 · L2.1 |
| style-vs-rules-conflict | persona/tono/autonomía contradice CLAUDE.md o permisos | warn | badge «estilo en conflicto con reglas» | L1.5 · L2.3 |
| style-orphaned | archivo de estilo nunca referenciado por `outputStyle` ni `force-for-plugin` | info | badge «estilo huérfano, nunca seleccionado» | L1.3 |
| style-deprecated-cmd | scripts/docs del arnés usan `/output-style` (removido v2.1.91) | error | badge «comando obsoleto /output-style» | L1.3 |
| style-rigid-phrasing | dicta frases literales en vez de reglas de estructura | info | badge «frases rígidas — frágil entre versiones» | L1.5 |
| style-tone-suppression | intenta suprimir tono base sin respaldo de hook | warn | badge «supresión de tono no garantizada (issue #6450)» | L1.6 · L2.4 |

## Changelog

- 2026-07-04 · v1.0 · Nodo fundacional. L1 de docs oficiales (output-styles, prompt-caching) +
  anchor blog (jerarquía de peso: styles nunca se compactan) + issue #6450 (límite: enforcement
  probabilístico). L2 amarra output style = superficie menos usada por diseño (el rol vive en
  skills/rules), cross-check con reglas/permisos, supresión de tono ⇒ hook. 8 checks. Novedades:
  `/output-style` removido v2.1.91, `force-for-plugin`, resolución dir anidado. · pasada fundacional.
