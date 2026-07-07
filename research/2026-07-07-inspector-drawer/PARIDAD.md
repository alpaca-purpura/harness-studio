# PARIDAD — mockup v6 ↔ app (se llena fila por fila al implementar)

> Contrato de «lo que ves en el mockup ES lo que hace la app». Gate final del paquete:
> todas las filas ✅ + click-through lado a lado (app real vs artifact) con consola
> limpia y screenshots revisados.

| RF | Elemento (mockup v6) | Componente real | Story=test | Estado |
|---|---|---|---|---|
| RF-80 | ⤢ ampliar (524) | inspector.tsx | — | ⬜ |
| RF-81 | ✕ cierra ≠ ⤡ colapsa (552-568) | inspector.tsx + página | — | ⬜ |
| RF-82 | tabs 3-up (527-531) | inspector.tsx | — | ⬜ |
| RF-83 | expandido: sticky+rejilla+920 (152-167) | inspector.tsx / map.css | — | ⬜ |
| RF-84 | estado vacío (541-550) | inspector.tsx / página | — | ⬜ |
| RF-85 | «i» secciones + SEC_TIP (309-326) | entities/arnes/model/doctrina.ts | — | ⬜ |
| RF-86 | campos punteados + DEF_CAMPO/VALOR (327-362) | doctrina.ts + inspector.tsx | — | ⬜ |
| RF-88 | a11y tooltips/contraste | ambos | — | ⬜ |
| RF-89 | Viene de (364-371) | selectors.ts + inspector.tsx | — | ⬜ |
| RF-90 | chips navegables + cond (373-378) | inspector.tsx + onSelect | — | ⬜ |
| RF-91 | Hallazgos (379-388) | inspector.tsx + api conformance | — | ⬜ |
| RF-92 | botonera staged (501-508) | inspector.tsx | — | ⬜ |
| RF-93 | tab Contenido + endpoint fuente (410-433) | inspector.tsx + Go router/usecase | — | ⬜ |
| RF-94 | no-reconocido RAW | ambos | — | ⬜ |
| RF-95 | acciones staged Contenido | inspector.tsx | — | ⬜ |
| RF-96 | tab Corridas honesta (434-443) | inspector.tsx | — | ⬜ |

Leyenda: ⬜ pendiente · 🔶 implementado sin verificar · ✅ verificado en paridad.
