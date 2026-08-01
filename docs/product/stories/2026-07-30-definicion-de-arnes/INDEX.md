# Definición canónica de arnés + cadena de proceso — paquete de doctrina

> Nace 2026-07-30 de la pregunta fundacional del operador: «¿qué es un arnés? ¿un
> plugin puede tener múltiples arneses? ¿feature/bugfix/spike son arneses distintos?»
> (transcripta en `chris-input.md`). Es un paquete de DOCTRINA: el entregable es la
> definición firmada + su bajada as-code, no código de app.

**Qué es:** cementar QUÉ ES UN ARNÉS por definición para ArnesIA — el corte
(rol × proceso vs end-to-end vs tipo-de-trabajo), la relación arnés↔plugin, y la
**cadena** (el proceso end-to-end que atraviesa varios arneses) como entidad
declarada. Primero dev como dogfood, luego generaliza a otros rubros (p7).

## Etapas §10

| Etapa | Estado |
|---|---|
| Decisiones | ✅ **FIRMADAS 🧑‍⚖️ 2026-08-01** — DEF-D1..D3 + directivas DEF-D5/D6 · DEF-D4 hallazgo gentle-ai · registro AUD-1..9 (auditoría de la spec) |
| Spec | ✅ **v2 FIRMADA 🧑‍⚖️ 2026-08-01** («ok firmo», /goal del operador) — [`spec.md`](./spec.md) (crear·visualizar·mantener + MA-L1..L7 + MA-E1..E13 + carril D del MVP; hallazgos AUD-1..9 en `decisiones.md`) |
| Mockup | ✅ **MA-T6 GATE FIRMADO 🧑‍⚖️ 2026-08-01** (misma firma) — [`mockup-mapa-actividades.html`](./mockup-mapa-actividades.html) (re-derivado a la superficie VIGENTE; fila en `mockups/INDEX.md`) — FE desbloqueado |
| Implementación | ✅ **MA-T1a..T7 EJECUTADOS 2026-08-01** — seam al indexar + saneo + `developer-vitalia` 0.1.0 publicado y VIVO en el Mapa de la app instalada (chips N0 · foco N1 · E5/E6/E7/E13 reales) |
| PARIDAD | ✍️ [`PARIDAD.md`](./PARIDAD.md) con evidencia por ticket (§8 completo, 4 hallazgos de dogfood D/E/F/peek visibles) — **gate 🧑‍⚖️ del operador pendiente** |

## Retomar aquí

1. **Resolver las deudas de dogfood D · E · F + gotcha peek COMO PARTE DE ESTA
   STORY** → **[`handoff-deudas-dogfood.md`](./handoff-deudas-dogfood.md)**
   (estado vivo, causas verificadas con file:line, orden de ataque F→peek→D→E,
   opciones de diseño y las 3 preguntas para el operador ANTES de codear D/E).
2. **Gate 🧑‍⚖️ de PARIDAD del operador** (`PARIDAD.md`): chips N0 + foco N1 en la
   app instalada (sesión propia de `developer-vitalia`). Independiente de 1.
2. Sub-especificaciones que la spec dejó nombradas y NO resueltas (van naciendo
   con sus tickets): hoja canónica de la definición en `docs/` · schema
   `proceso/<id>.yaml` (2 sub-preguntas DEF-D3) · afilar META `proceso`→referencia.
3. Contexto completo del debate: `debate-definicion.md` (§4b jerarquía · §5 cadena).
