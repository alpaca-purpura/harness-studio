# Medición p11 — economía de contexto medible (RF-132)

Fecha: 2026-07-08 · Método: `spec-writer` headless REAL 2× con `claude -p
--output-format json --max-turns 12 --plugin-dir <arnés>` sobre la MISMA idea
(«exportar reporte mensual de cobranzas a CSV, separador configurable»), cwd tmp
aislado por corrida. (a) = dogfood ANTES de la Fase 4 (commit `bb91ab9`: estándar
std-spec en prosa, sin plantilla/validador/hooks). (b) = dogfood CON Fase 4
(plantilla `references/plantilla-spec.md` + `scripts/validate_spec` + Guardia
`hooks/hooks.json`) y el prompt estilo `tarea()` nuevo (bloque de insumos). Modelo:
claude-fable-5 en ambas. Crudos en scratchpad `p11/` (result-a/b.json, stream-c.jsonl).

## 1. La corrida del escritor (spec-writer)

| métrica | (a) prosa | (b) plantilla+gate | Δ |
|---|---|---|---|
| subtype | success | success | — |
| turnos | 7 | 10 | +3 |
| duración | 79,9 s | 92,1 s | +15 % |
| input tokens | 15.257 | 15.394 | ≈ = |
| cache creation | 24.852 | 29.313 | +18 % |
| cache read | 144.805 | 313.589 | +117 % |
| output tokens | 3.916 | 5.060 | +29 % |
| costo (USD) | 0,990 | 1,307 | +32 % |
| `status` del artefacto | prosa libre (sin estampa) | **`done` estampado por el validador** | — |
| digest determinista | NO existe | **`spec.md.digest.md` (301 B ≈ 75 tok)** | — |

**Lectura honesta:** la corrida del ESCRITOR cuesta MÁS con plantilla+gate (+29 %
output, +3 turnos, +32 % USD). Ese sobrecosto compra: estructura determinista
(script, no LLM — D5), `status: done` estampado por código (jamás autodeclarado),
digest sidecar generado, y el gate de Guardia corriendo en cada Write. El spec (b)
además salió MÁS CORTO y más denso (1.857 B vs 2.999 B: la plantilla mata la prosa
decorativa).

## 2. El hand-off (donde p11 realmente paga)

El ahorro del diseño D7 no está en el escritor: está en cada eslabón AGUAS ABAJO,
multiplicado por cada iteración del consumidor.

| qué viaja al siguiente eslabón (builder) | bytes | ≈ tokens |
|---|---|---|
| ANTES: spec.md entero (leído por el consumidor o inyectado) | 2.999 | ≈ 750 |
| DESPUÉS: `tarea()` inyecta ruta + digest (F3) | 301 | ≈ 75 |
| **ahorro por lectura de insumo** | **−90 %** | **≈ −675 tok** |

Determinista y verificable sin LLM: `wc -c` de los artefactos reales de la corrida
(b). Con el repair-cap del conductor (≤3 iteraciones) y N cajas consumidoras, el
ahorro escala ≈ 675 × lecturas evitadas; el consumidor siempre puede abrir la ruta
si necesita el detalle (el doc no se pierde — se deja de DUPLICAR en cada prompt).

## 3. La secuencia del gate (RF-131, corrida dirigida c)

`stream-json` real (`stream-c.jsonl`): se ordenó copiar la plantilla CON
placeholders → **PostToolUse `spec-guard.sh` bloqueó el Write con
`decision:block`** y las 7 violaciones (una por placeholder) → el modelo leyó la
razón, llenó el spec para la idea de prueba → el MISMO validador pasó, estampó
`status: done` y generó el digest. 11 turnos, exit 0. El gate se probó en ambas
direcciones en headless real, no simulado.

## Veredicto

- p11 medido: **el hand-off encadenado ahorra ~90 % del contexto por insumo**; el
  escritor paga +32 % una vez por artefacto a cambio de validación determinista.
- El número que cementa la doctrina: **digest ≈ 75 tokens ≤ 200** (D7 cumplido).
- Trade-off registrado, no maquillado: si una caja NO tiene consumidores aguas
  abajo, la plantilla+gate es costo neto — coherente con D4/D5 (obligatorio solo
  arquetipos pipeline/excepcion, `abierto` exento).
