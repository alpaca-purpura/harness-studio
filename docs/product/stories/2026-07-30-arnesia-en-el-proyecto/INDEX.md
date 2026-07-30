# ArnesIA en el proyecto — semilla `.arnesia/` (carril A) — MVP 1 día

> Paquete del MVP 2026-07-30 (prioridad 2). Decisiones firmadas vía aprobación del
> plan de sesión (`decisiones.md`).

**Qué es:** nace la política `.arnesia/` — la carpeta del PROYECTO del usuario donde
vive el process-as-code sembrado por ArnesIA (terreno stubs · product · wip ·
plantillas de proceso POR TIPO de paquete), derivada determinísticamente del modelo de
terreno FIRMADO (D0-D20 + `arnes.yaml`). Incluye `arnesia init` (+ `--check` doctor),
endpoints de forja, y la enmienda ADITIVA a la doctrine que sanciona `.arnesia/` como
zona de escritura del proceso.

## Etapas §10

| Etapa | Estado |
|---|---|
| Mockup | n/a (sin superficie FE nueva; CLI + contrato) |
| Decisiones | ✅ FIRMADAS 🧑‍⚖️ 2026-07-30 (aprobación del plan) — `decisiones.md` |
| Spec | ✅ `spec.md` |
| Implementación | ✅ en main (merge 07f9d3b) |
| PARIDAD | ⬜ Installer 3: `arnesia init` sobre proyecto real del operador |

## Retomar aquí

Orden: A-T1 contrato `contracts/semilla-arnesia.md` → A-T2 `semilla/` embebido +
`internal/adapters/forja/` → A-T3 usecase + CLI `init` + HTTP → A-T5 enmienda
doctrine (∥). Recortes y verificación en `spec.md`.
