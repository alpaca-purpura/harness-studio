# PARIDAD — ciclo conversacional, cables del Dock

> Evidencia 2026-07-30. **Gate 🧑‍⚖️: PENDIENTE** (click-through en laptop con
> tarjeta real de mejora). Código en main: `e55006e`.

| Qué (RF/AC) | Evidencia | Estado |
|---|---|---|
| «Proponerlo en el chat» abre Dock + alcance + composer prellenado, JAMÁS auto-envía | E2E vivo (daemon sandbox + Playwright): dock montó, chip `skill · spec-writer · …`, textarea con foco al final, buzón consumido, `GET /api/sessions` = 0 turnos | ✅ |
| Chip de alcance removible/re-fijable, sin pisar composer | E2E vivo: ✕ quita, botón re-fija, texto intacto | ✅ |
| «Editar conversando» vivo solo `viewedId===arnesId` (CH-D6) | E2E vivo DOM + story `EditarConversandoCableado` | ✅ |
| Disabled honesto sin sesión (title nuevo) | story `EditarConversandoSinSesion` con title exacto pinneado | ✅ |
| «Editar fuente» intacto (Fase 2) | story asserta disabled | ✅ |
| Descartar NO tocado (ya estaba cableado — C-D1) | verificado en vivo pre-cambio | ✅ |
| Suite | 20/20 inspector · 607/607 total · verify exit 0 · gate a11y | ✅ |
| **AC-d: click-through con tarjeta REAL de mejora (exige telemetría medida) en laptop** | | ⬜ **gate** |

Gap declarado: la cadena C-T1a sobre tarjeta real no se pudo escenificar en sandbox
(sin corridas ⇒ sin tarjetas); la cadena idéntica (openChat+setScope) quedó probada
vía C-T1b + el buzón en vivo.

## Firma

- [ ] 🧑‍⚖️ click-through en laptop — fecha:
