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
| PARIDAD | ⏳ **gate humano 🧑‍⚖️ pendiente** (click-through del operador) |

## Retomar aquí

**Los 6 CH-D CONSTRUIDOS y verificados E2E vivo** (ver `spec.md` as-built): composer 3
líneas · resize+persistencia · gate paquete-cerrado (CAP-99) · actividad desplegable ·
burbuja por paso · markdown con chips/copiar/sintaxis (CAP-100). Evidencia E2E: turno real
contra vitalia partido en burbujas + tarjetas, screenshots `chat-dock-tramoB.png` /
`chat-dock-actividad.png` (raíz, no versionados). Falta SOLO el gate humano de PARIDAD:
click-through del operador en la app (composer · resize · tarjeta desplegable · chips
copiar · deny del paquete cerrado). Desviación honesta: tarjeta sin segundos («N pasos»).
