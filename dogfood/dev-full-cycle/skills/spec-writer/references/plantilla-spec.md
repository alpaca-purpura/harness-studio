---
status: draft
inputs:
  - {{ruta-del-insumo}}
creado: {{YYYY-MM-DD}}
actualizado: {{YYYY-MM-DD}}
---

# spec — {{título de la unidad de trabajo}}

## Contenido

- [why](#why)
- [Capacidades](#capacidades)
- [non_goals](#non_goals)

## why

{{una línea: la intención inmutable de la unidad de trabajo}}

## Capacidades

| CAP | qué | success |
|-----|-----|---------|
| CAP-01 | {{qué logra — WHAT, no HOW}} | {{señal concreta y verificable}} |

## non_goals

- {{lo que esta unidad NO hace}}

<!--
Esqueleto de spec.md (franja-artefactos D4/RF-130, estándar base:std-spec).
Cómo llenarlo (D5 — script para estructura, LLM para juicio):
- Copia este archivo a spec.md, reemplaza TODO placeholder de doble-llave y borra
  este comentario — el validador (scripts/validate_spec) rechaza cualquier residuo.
- `status:` lo estampa scripts/validate_spec al pasar (done) — JAMÁS lo escribas a mano.
- `inputs:` = rutas reales de los insumos consumidos (D7).
- Cada CAP-NN lleva success verificable; sin capability sin success (std-spec).
- Filas CAP adicionales: CAP-02, CAP-03… ids estables entre versiones.
-->
