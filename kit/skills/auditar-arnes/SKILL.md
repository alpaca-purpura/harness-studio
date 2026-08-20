---
name: auditar-arnes
description: Auditar el arnés del cwd contra la doctrina ArnesIA — nomenclatura, contratos fusionados, spine, honestidad de gates — y proponer el plan de mejora (usa cuando el usuario pida revisar, auditar o encontrar puntos de mejora del arnés).
---

# Auditar el arnés

Auditas el arnés del cwd (cuerpo ③) contra el estándar. Tu salida es un informe de
brechas priorizado + plan de mejora — NO cambies nada sin que el usuario lo pida.

## Proceso

1. **Reconocimiento** (nomenclatura v1): ¿hay `arnes.l0.json` en la raíz? ¿las skills
   viven en `skills/<id>/SKILL.md` (directorio)? ¿qué archivos NO reconoce la
   nomenclatura? Todo lo no reconocido se REPORTA visible, jamás se ignora.
2. **Abre el nodo del estándar de CADA clase presente en el arnés** — no solo
   `knowhow/skills.md`. Son 12 y cada uno trae su checklist evaluable: `skills` · `hooks`
   · `rules` · `subagents` · `commands` · `mcp` · `plugins` · `settings-permissions` ·
   `output-styles` · `statusline` · `headless-sdk` · `harness-profile`. Auditar una clase
   sin abrir su nodo es opinar, no auditar: los checks que no leíste no los podés
   reportar. **Declara qué nodos abriste** — esa lista define el alcance real de la
   auditoría, y lo que quedó fuera es un hueco del informe, no del arnés.
3. **Manifiesto**: fases y spine declarados, META de enganche (rol · proceso ·
   reporta_a · empresa) completa. Sin manifiesto = brecha error.
4. **Contratos caja por caja**: frontmatter fusionado completo (3 ejes) · capabilities
   con success verificable · `estado` es transición legal del spine · un escritor_unico
   por artefacto · gate HONESTO (¿`auto` con Gherkin que de verdad corre, o debería ser
   `none`? un gate inflado es hallazgo grave) · handoff definido.
5. **Cableado del grafo**: necesita sin productor (huérfanos) · cajas sin ruta de
   salida · ciclos no declarados como retrabajo.
6. **Firewall CC-native**: greppea frontmatter fantasma (persistent_facts /
   activation_steps_prepend / customize / sanctum) en TODOS los .md del arnés.
7. **Veredicto mecánico**: si puedes, corre `arnesia conformance --arnes <graph.l0>` y
   anexa su salida; lo que el motor marca `deferred` es gris honesto, no verde.
8. **Informe**: brechas ordenadas por severidad (error > warn), cada una con
   ruta:línea, **el `id` textual del check que viola** (`skill-gate-honest`,
   `no-phantom-frontmatter`, `hook-exit2-block`…) copiado de la tabla de su nodo, y el fix
   propuesto. Un hallazgo sin id de check es una opinión: o lo anclás al estándar o no
   entra al informe. Y el id se COPIA del nodo, no se recuerda de memoria — citar un check
   que no existe es la misma falta que el gate inflado. Cierra con el plan de mejora en
   orden de ataque.

## Reglas

- Gris ≠ verde: lo no verificable se declara gris.
- No modifiques el arnés durante la auditoría; el fix llega después y por pedido.
- Nada del kit/doctrina se escribe en el arnés (② no contamina ③).
