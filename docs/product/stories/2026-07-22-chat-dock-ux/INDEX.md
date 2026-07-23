# Paquete — Chat dock · legibilidad y ergonomía

> alta 2026-07-22 · origen: 6 quejas directas del operador sobre el espacio de conversación de
> la app. Disciplina §10: mockup→decisiones→spec→implementar→PARIDAD con firmas 🧑‍⚖️.

## Alcance (las 6 quejas)

1. Composer auto-crece hasta 3 líneas (CH-D1)
2. Actividad visible + detalle desplegable estilo Cursor (CH-D2)
3. Una burbuja por paso del asistente (CH-D3)
4. Markdown renderizado (CH-D4)
5. Dock redimensionable por el borde izquierdo (CH-D5)
6. Guardrail: chat de la app NO toca los arneses propios de arnesia (CH-D6)

## Estado

| etapa | estado |
|---|---|
| decisiones | ✅ **firmadas 🧑‍⚖️ 2026-07-22** («firmo, prosigue») — CH-D1..D6 (+D4b «cariño» mismo día) |
| mockup | ✅ **Gate 1 firmado 🧑‍⚖️ 2026-07-22** («firmo, implementa») — `mockups/arnesia-chat-dock-ux.html` |
| spec | ✅ AS-BUILT completo (`spec.md`, tramos A y B) |
| implementar | ✅ **COMPLETO** — tramo A (CAP-99) + tramo B (CAP-100), suite Go + fitness + verify FE + E2E vivo con claude real |
| PARIDAD | ✅ **FIRMADA 🧑‍⚖️ 2026-07-22** («DE momento todo bien… firmo»; error posterior = bugfix nuevo) |

## Retomar aquí

**PAQUETE CERRADO (ficha `ledger/HS-26.md`).** Los 6 CH-D + D4b construidos, verificados E2E
vivo con claude real y FIRMADOS (evidencia y desviaciones aceptadas → `PARIDAD.md`). Deuda
viva que NO reabre el gate: llevar el build al escritorio (`make installer` + self-update) ·
segundos en la tarjeta si el operador los pide · parseo por-tool del guardrail CH-D6 —
registrada en `BACKLOG.md`. Un error que aparezca se trabaja como bugfix en paquete nuevo.
