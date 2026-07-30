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
| AC-d: click-through con tarjeta REAL de mejora | **Auto-verificado 2026-07-30** (daemon del `.deb` v0.6.0 + Playwright/Chromium): telemetría inyectada por el CAMINO REAL (`POST /v1/logs` OTLP/JSON, shape del golden de CC 2.1.220) → detector **B4 gasto-concentrado** disparó de verdad (93 % en `hipaa-check`, USD 0,11) → click en «Proponerlo en el chat»: Dock abre · chip `skill hipaa-check` con fuente_path real + caja resaltada en canvas · composer prellenado con foco al final (caret 59/59) · **0 turnos enviados** · Descartar quita la tarjeta y PERSISTE en el daemon. Capturas + script en [`verificacion-tarjeta-real/`](./verificacion-tarjeta-real/) | ✅ |
| **Residuo humano: verlo en TU laptop + firma** | | ⬜ **gate** |

Gap original («sin corridas ⇒ sin tarjetas») CERRADO por la inyección OTLP de arriba —
cero datos fabricados a mano: el detector computó sobre ingesta real del receptor.
Nota fiel del reporte: `GET /api/telemetria/arneses/{clave}/mejoras` es la ruta real
(el spec citaba `/api/harnesses/{id}/mejoras`, que no existe).

## Firma

- [ ] 🧑‍⚖️ click-through en laptop — fecha:
