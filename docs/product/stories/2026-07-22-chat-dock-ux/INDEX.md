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
| decisiones | ✅ **firmadas 🧑‍⚖️ 2026-07-22** («firmo, prosigue») — CH-D1..D6 |
| mockup | 📋 propuesta publicada (`mockups/arnesia-chat-dock-ux.html`, CH-D2/D3/D4) — **Gate 1 PENDIENTE 🧑‍⚖️** |
| spec | ✅ tramo A (CH-D1/D5/D6) as-built en `spec.md` · tramo B tras Gate 1 |
| implementar | ✅ tramo A construido + verificado (suite Go + fitness + E2E vivo playwright) · tramo B ⏳ |
| PARIDAD | ⏳ |

## Retomar aquí

**Tramo A (CH-D1 composer 3 líneas · CH-D5 resize+persistencia · CH-D6 gate paquete cerrado
CAP-99) CONSTRUIDO y verificado** — ver `spec.md` (as-built). **Siguiente: Gate 1 🧑‍⚖️ del
mockup** `mockups/arnesia-chat-dock-ux.html` (actividad desplegable · burbuja por paso ·
markdown); firmado → spec del tramo B: daemon deja de descartar thinking/tool_use
(`conductor.go:515-520`), corte de burbuja por boundary (`session_service.go:428-471` +
`sessions-store.ts:282-315`), renderer markdown (dependencia nueva, elegir). PARIDAD del
tramo A se firma junto con el paquete.
