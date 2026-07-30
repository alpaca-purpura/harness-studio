# Volverlo de arnesia y publicar (B1 + B2) — MVP 1 día

> Paquete PRIORIDAD 1 del MVP 2026-07-30. Corte y decisiones aprobados por el operador
> vía plan de sesión (ver `decisiones.md`). Plan completo del día:
> `~/.claude/plans/playful-seeking-hippo.md` (extra-repo) + `docs/product/checkpoint.md`.

**Qué es:** cierra el ciclo extraer→sellar→publicar. B1: `Identificar` emite un sello
`arnes.l0.json` VÁLIDO contra `graph.l0.schema.json` (hoy escribe uno inválido — bug) y
nace la política de versionado de arnés as-code. B2: write-side real del formato
prenter-marketplace (publish con gate de conformance, pull-antes-de-push, tag-tras-push).

## Etapas §10

| Etapa | Estado |
|---|---|
| Mockup | ✅ resuelto como superset del Storybook vigente, sin html nuevo (B-D4) |
| Decisiones | ✅ FIRMADAS 🧑‍⚖️ 2026-07-30 (aprobación del plan) — `decisiones.md` |
| Spec | ✅ `spec.md` (spec-lite sancionada por la firma del corte) |
| Implementación | 🔨 EN CURSO |
| PARIDAD | ⬜ al cierre del bloque (installer + laptop) |

## Retomar aquí

Implementación en curso: B1 primero (backend → HTTP/openapi → FE form), después B2
(dominio/puerto → adapter git → usecase → HTTP → FE botón). Innegociables y orden de
recorte en `spec.md` §Recortes. Verificación: `verificando-binario-instalado` antes de
cada installer; E2E real desde la laptop del operador.
