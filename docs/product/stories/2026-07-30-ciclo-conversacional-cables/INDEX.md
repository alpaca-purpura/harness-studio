# Ciclo conversacional — cables del Dock (carril C) — MVP 1 día

> Paquete del MVP 2026-07-30 (prioridad 3, barato y de alto valor demo). Decisiones
> firmadas vía aprobación del plan de sesión (`decisiones.md`).

**Qué es:** los 2 cables que faltan para que «mejoras narradas → aprobar conversando»
se sienta UNA experiencia: la tarjeta de punto de mejora abre el Dock con alcance
(hoy la propuesta queda invisible si el Dock está colapsado — `openChat()` tiene 0
callers), y el inspector habilita «Editar conversando (dock)». Stretch: watcher sobre
instalaciones → re-evaluar deriva → SSE.

## Etapas §10

| Etapa | Estado |
|---|---|
| Mockup | n/a (cero componente nuevo: cableo de superficies firmadas HS-26/capa Mejora) |
| Decisiones | ✅ FIRMADAS 🧑‍⚖️ 2026-07-30 (aprobación del plan) — `decisiones.md` |
| Spec | ✅ `spec.md` |
| Implementación | ✅ en main (e55006e) |
| PARIDAD | ⬜ navegador contra daemon real + laptop |

## Retomar aquí

C-T1a (`workspace-stage.tsx` onProponer += openChat+setScope) → C-T1b (prop
`onEditarConversando` del inspector). C-T2 watcher = SOLO si sobra el día.
