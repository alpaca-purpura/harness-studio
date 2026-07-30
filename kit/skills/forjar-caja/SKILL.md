---
name: forjar-caja
description: Crear una caja de proceso nueva en el arnés actual — skill + contrato fusionado completo conforme a la doctrina ArnesIA (usa cuando el usuario pida añadir una caja, una fase-frente o una skill de proceso al arnés).
---

# Forjar una caja

Creas UNA caja de proceso en el arnés del cwd, conforme al contrato fusionado
(doctrine.md tiene el resumen; `knowhow/skills.md` la checklist completa).

## Proceso

1. **Lee el manifiesto** `arnes.l0.json` de la raíz: fases y spine declarados. La caja
   nueva vive en UNA fase y posee UNA transición del spine. Si la fase o la transición
   no existen, detente y pregunta — jamás inventes estados.
2. **Grill corto al usuario** (si falta): qué convierte en qué (necesita → entrega),
   quién más escribe ese artefacto (escritor_unico), cuándo se considera hecho (gate).
3. **Clasifica los 3 ejes** y dilo explícito: `arquetipo` (¿pipeline verificable,
   excepción acotada, abierto generativo… o `no-arnesar`? — si es puro juicio sin
   descomposición contratable, NO la forjes: regístralo como no-arnesar y explica) y
   `perfil_harness` (T1 pasada única · T2 multi-paso interactivo · T3 loop desatendido
   del conductor).
4. **Escribe `skills/<id>/SKILL.md`** (directorio, no archivo plano): frontmatter =
   contrato fusionado completo (why · capabilities con success verificable · constraints
   · non_goals · clase/arquetipo/perfil_harness · caja/fase/estado · necesita/entrega/
   ruta · gate honesto · handoff). Cuerpo = instrucciones operativas de la caja (cómo
   trabaja, no qué promete — lo que promete ya está en el frontmatter).
5. **Identidad + plantilla del artefacto (cajas `pipeline`/`excepcion` — franja-artefactos
   D3/D4/D5)**: la entrega primaria declara `path:` (artefacto-archivo; sin path es
   etiqueta y el check `art-es-path` lo marcará warn). Si la caja PRODUCE un documento
   propio, materializa el trío: `skills/<id>/references/plantilla-<art>.md` (esqueleto con
   frontmatter `status:`/`inputs:` + placeholders de doble-llave) + `skills/<id>/scripts/
   validate_<art>` (valida estructura con errores verbosos; estampa `status: done` SOLO al
   pasar y genera `<art>.digest.md` determinista ≤200 tokens) + `entrega[].plantilla:` con
   la ruta de la reference. Los inputs que llegan del mundo (`de: usuario|terceros:*`)
   JAMÁS llevan plantilla — se ADMITEN por script en la skill consumidora (D10). `abierto`
   está exento (§8.1). Modelo de referencia: `spec-writer` del arnés dev-full-cycle.
6. **Gate honesto**: si no existe eval ejecutable hoy, `gate.tipo: none` con `detalle`
   de qué evidencia lo sustituye. Gherkin solo si de verdad corre (`tipo: auto`) — y si
   hay validador de plantilla, el Gherkin lo cita (es el mismo script que la Guardia
   engancha en PostToolUse/Stop, D6).
7. **Cablea los edges**: `necesita.de` con origen explícito (`base:`/`caja:`/…) y
   `ruta.a` hacia la caja siguiente o `humano`. Un artefacto = UN escritor — la única
   multi-escritura legal es la cadena `refina` (revisión lineal, D9); rework al mismo
   escritor va por `ruta[]`, no por refina.
8. **Verifica**: corre `arnesia conformance --arnes <graph>` si hay grafo exportado, o
   revisa contra `knowhow/skills.md` (checklist L2) y declara qué checks quedan grises.

## Prohibido

- Frontmatter fantasma (persistent_facts / customize / sanctum / activation_steps_prepend).
- Estados fuera del spine · dos escritores de un artefacto · gate inventado.
- Escribir CUALQUIER archivo del kit/doctrina dentro del arnés (② no contamina ③). La
  prohibición es COPIAR kit/doctrina: `.arnesia/` del proyecto es zona LEGAL de artefactos
  del proceso (contrato `semilla-arnesia.md`, enmienda A-D3).
