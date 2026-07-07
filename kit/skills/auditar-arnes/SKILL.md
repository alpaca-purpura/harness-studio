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
2. **Manifiesto**: fases y spine declarados, META de enganche (rol · proceso ·
   reporta_a · empresa) completa. Sin manifiesto = brecha error.
3. **Contratos caja por caja**: frontmatter fusionado completo (3 ejes) · capabilities
   con success verificable · `estado` es transición legal del spine · un escritor_unico
   por artefacto · gate HONESTO (¿`auto` con Gherkin que de verdad corre, o debería ser
   `none`? un gate inflado es hallazgo grave) · handoff definido.
4. **Cableado del grafo**: necesita sin productor (huérfanos) · cajas sin ruta de
   salida · ciclos no declarados como retrabajo.
5. **Firewall CC-native**: greppea frontmatter fantasma (persistent_facts /
   activation_steps_prepend / customize / sanctum) en TODOS los .md del arnés.
6. **Veredicto mecánico**: si puedes, corre `arnesia conformance --arnes <graph.l0>` y
   anexa su salida; lo que el motor marca `deferred` es gris honesto, no verde.
7. **Informe**: brechas ordenadas por severidad (error > warn), cada una con
   ruta:línea, el check del estándar que viola (cita `knowhow/<elemento>.md`) y el fix
   propuesto. Cierra con el plan de mejora en orden de ataque.

## Reglas

- Gris ≠ verde: lo no verificable se declara gris.
- No modifiques el arnés durante la auditoría; el fix llega después y por pedido.
- Nada del kit/doctrina se escribe en el arnés (② no contamina ③).
