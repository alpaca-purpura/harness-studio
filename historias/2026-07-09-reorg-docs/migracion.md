# Migración — qué línea va a dónde (plan de ejecución, post-firma)

> vig: activo · revisar: 2026-08-01 · NO ejecutar hasta que el operador firme el árbol.
> Todo es reversible por git. Un commit por paso para poder revertir granular.

## Mapa origen → destino

| Origen (hoy) | Destino |
|---|---|
| `CLAUDE.md` bloques narrativos (HS-06/11/14/17…, cifras, «LANDEADO») | se BORRAN de CLAUDE.md; el dato ya vive en `ledger/` |
| `CLAUDE.md` punteros (VISION/METODOLOGIA/knowledge/arch/UX + reglas git/§10) | se conservan → `CLAUDE-nuevo.md` (ya draft) |
| `CLAUDE.md` cifras (240 checks/28 pass…) | `ESTADO.md`, bloque generado (RF-178) |
| `LEDGER.md` fichas prosa `### HS-NN …` (líneas 11-994) | `ledger/HS-NN.md` (una hoja c/u, cabecera `archivado`) |
| `LEDGER.md` tabla Log (líneas 997-1020) | se queda en `LEDGER.md` = el índice grep-able |
| `LEDGER.md` campos `Siguiente:`/`Deuda:` dentro de fichas | se ELIMINAN de las fichas → `BACKLOG.md` |
| 106 RFs dispersos en `historias/*/spec.md` + `UX.md:677` inventario (stale) | se mapean a `CAP-NN` → `CAPABILITIES.md` (saldo validado) |
| 7× `historias/*/INDEX.md` sección «Retomar aquí» | `historias/*/RETOMAR.md` (opcional, barato) |

## Orden de ejecución (reversible, 1 commit/paso)

1. **Crear las hojas nuevas sin borrar nada** (aditivo, cero riesgo):
   `BACKLOG.md` (de este draft, reconciliando `verificar`) · `ESTADO.md` (fase + paquete
   activo + bloque de cifras generado) · `CAPABILITIES.md` (de este draft; linkear checks
   reales por grep de test names, resolver `(por linkear)`) · `ledger/HS-NN.md` (split fichas).
   → en este punto todo existe por duplicado; nada se perdió.
   Nota CAPABILITIES: es DERIVADO — el registro se puebla mapeando los 106 RFs a `CAP-NN` y
   linkeando cada uno a su test/check; el `estado` sale de conformance (RF-181), no a mano.
2. **Verificar paridad:** `grep` de cada HS-NN encuentra su hoja; `BACKLOG` cubre todo
   `Siguiente`/`Deuda`; `ESTADO` cifras = salida real de `conformance --todo`.
3. **Adelgazar `LEDGER.md`** a solo la tabla-índice (quitar las fichas prosa ya migradas y
   los campos `Siguiente`/`Deuda`).
4. **Reemplazar `CLAUDE.md`** por el router (`CLAUDE-nuevo.md`).
5. **(Opcional) partir los `INDEX.md`** en `RETOMAR.md`.
6. **Actualizar `MEMORY.md`** (índice de memoria): apuntar el patrón nuevo; y `arch/CADENCE.md`
   con el mecanismo de TTL/revisión de hojas.
7. **Medir:** `wc -c CLAUDE.md BACKLOG.md ESTADO.md` → confirmar ≤ ~4k tok combinados vs
   baseline 30k (spec §6). Escribir el número en `PARIDAD.md` del paquete.

## Riesgos y mitigación

- **Punteros rotos:** otros docs citan `LEDGER.md#HS-NN` en prosa. Mitigación: `grep -rn
  "HS-[0-9]" *.md arch/ knowledge/ historias/` antes del paso 3; re-apuntar a `ledger/HS-NN.md`.
- **Pérdida de contexto en fichas:** las fichas son historia firmada — se COPIAN íntegras a
  `ledger/`, no se resumen. El split es mecánico (por `### HS-NN`), no editorial.
- **CLAUDE.md es cargado por el harness:** el router debe seguir teniendo las reglas duras
  (git, §10) porque son OVERRIDE de comportamiento. Ya están en el draft.
