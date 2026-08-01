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
| Implementación | ⬜ MA-T1..T5/T7 (`spec.md` §7) |
| PARIDAD | ⬜ |

## Retomar aquí

1. **Dos gates 🧑‍⚖️ del operador, juntos:** firma de `spec.md` v2 (auditada
   AUD-1..9) + gate del mockup `mockup-mapa-actividades.html` (MA-T6, abrir en
   navegador). Con firmas → MA-T1a/T1b (dato+seam) → MA-T2 (saneo REAL: dup
   legacy + vitalia-app sin-home) → MA-T3 (**forja `developer-vitalia`**: tabla
   de clasificación con 6 ambiguos A DECIDIR en el gate + 4 actividades + sello +
   publicar B2) → recién después MA-T4/T5/T7 (FE).
2. Sub-especificaciones que la spec dejó nombradas y NO resueltas (van naciendo
   con sus tickets): hoja canónica de la definición en `docs/` · schema
   `proceso/<id>.yaml` (2 sub-preguntas DEF-D3) · afilar META `proceso`→referencia.
3. Contexto completo del debate: `debate-definicion.md` (§4b jerarquía · §5 cadena).
