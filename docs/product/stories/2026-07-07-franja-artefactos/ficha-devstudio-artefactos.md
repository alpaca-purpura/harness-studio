# Aviso ArnesIA → DevStudio — identidad de artefactos en el contrato de caja (2026-07-08)

> Entregar este texto al agente de DevStudio (ficha gemela de cortesía, patrón
> DH-18/PB-25 — compromiso HS-12 «solo aditivos»). No requiere acción de DevStudio: es
> el anuncio de campos OPCIONALES nuevos en la forma-plugin que ya consumen.

**Qué cambió (todo aditivo, paquete franja-artefactos, RF-110):**
`box.contract.schema.json` — el `contract:` de cada SKILL.md-caja — ganó 3 campos
opcionales en `entrega[]`:

| campo | tipo | qué declara |
|---|---|---|
| `path` | string | identidad ARCHIVO del artefacto (ruta relativa al arnés). Presente = artefacto-archivo stat-eable; ausente = etiqueta («código + tests») |
| `plantilla` | string | ruta (relativa a la skill escritora) de la reference-esqueleto que la caja llena (`references/plantilla-<art>.md`) |
| `refina` | string | esta entrega es una REVISIÓN del art nombrado — única vía legal a la multi-escritura (cadena lineal; la versión se DERIVA) |

**Garantías para consumidores:**
- Contratos existentes validan SIN cambios (`additionalProperties` intacto para lo
  previo; los 3 campos son opcionales). El enum-5 de `spine.categorias` NO se tocó.
- El dogfood publicado `dev-full-cycle` ya los usa (spec-writer declara
  `path: spec.md` + `plantilla: references/plantilla-spec.md`) y suma su Guardia real
  (`hooks/hooks.json` — el loader la reconoce por la fila `hook` de nomenclatura §3,
  sin cambio de contrato: la fila existía desde v1).
- Los artefactos se PROYECTAN de `necesita[]/entrega[]` — jamás son nodos del L0 ni
  descriptor aparte (misma postura I-77 fichada en HS-12/P5). Si DevStudio quiere
  pintar hand-offs, la derivación de referencia vive en
  `web/src/entities/arnes/model/artefactos.ts` (pura, portable).
- Convención acompañante dentro de la skill (no es contrato, es layout): plantillas en
  `skills/<caja>/references/` · validadores en `skills/<caja>/scripts/` — el validador
  estampa `status: done` y genera `<art>.digest.md` (≤200 tokens) para el encadenado.

Detalle as-code: `docs/architecture/contracts/schema/box.contract.schema.json` ($comment D3/D9) ·
checks nuevos en `arnesia conformance --arnes`: `sin-huerfanos` · `dead-end` ·
`ruta-a-existe` · `art-identidad-coherente` · `refina-coherente` · `art-es-path` (warn).
