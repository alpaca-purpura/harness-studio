<!-- DO-NOT-EDIT · snapshot vivo AUTO-GENERADO desde arnes.yaml `gestion_trabajo` + docs/wip/activo|done.
     Hoy sembrado a mano (golden/fixture Slice 1a); el forjador (paso 2) lo regenera. No teclear cifras. -->
# WIP — trabajo vivo (instancia · Work)

> vig: activo · 4º territorio, naturaleza **INSTANCIA** (D15). Paquetes que se pueblan CONTRA el schema de
> `gestion-trabajo`; cada uno lleva **estado** + **criterio de cierre** y al terminar pasa a `done/` (anti-cajón).
> Al cerrar, ratifica su(s) **capability** → constituye el Producto (D16).

## Tipos de paquete (de `arnes.yaml` gestion_trabajo)

| tipo | jerarquía | estados | cierre | appetite |
|---|---|---|---|---|
| **historia** | outcome→historia→sub-tarea | idea · refinando · listo · en-curso · en-revisión · done | ratifica_capability | escalable (Shape-Up) |
| **spike** | spike | abierto · investigando · cerrado | documenta_decision (→ LEDGER) | — |

WIP-caps (flow, no sprints): historia → `refinando: 3` · `en-curso: 2`.

## Tablero (snapshot)

- **activo:** 0 paquetes en `docs/wip/activo/` · **done:** 0 en `docs/wip/done/`.

> **Nota de migración (honesta):** el WIP de-facto de ArnesIA vive hoy en `docs/product/stories/`
> (paquetes de trabajo). La migración `stories/` → `docs/wip/activo|done/` bajo este schema es futura
> (no está en el Slice 1a). Este árbol arranca vacío: es el **esqueleto** que el forjador monta.
