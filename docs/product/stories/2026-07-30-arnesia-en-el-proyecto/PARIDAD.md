# PARIDAD — arnesia en el proyecto (semilla `.arnesia/`)

> Evidencia 2026-07-30. **Gate 🧑‍⚖️: PENDIENTE** (init sobre proyecto REAL del
> operador). Código en main: A-T1 `0b31add` · A-T2..T5 merge `07f9d3b` (bdc1a1b).

| Qué (RF/AC) | Evidencia | Estado |
|---|---|---|
| Contrato `semilla-arnesia.md` + árbol fuente `semilla/` (yaml graduado byte-idéntico + 14 plantillas que parsean) | A-T1; tensiones del modelo → A-D5, no silenciadas | ✅ |
| `arnesia init` siembra el árbol del contrato | E2E: 19 creados · salud sana · exit 0 (worktree y post-merge) | ✅ |
| Idempotencia: re-run jamás pisa | E2E: 0 creados / 19 `ya-existia`; test de contenido intacto | ✅ |
| Doctor honesto con exit codes | E2E: borrar `01-spec.md` ⇒ `incompleta` + faltante listado + exit 1 | ✅ |
| Endpoints `/api/forja/semillas` (+`/chequeos`), campo `path` | E2E daemon sandbox: chequeo reporta incompleta; siembra repara SOLO lo ausente; path protegido ⇒ 400 | ✅ |
| Enmienda doctrine ADITIVA, runtime intacto | test frontera: `<proyecto>/.arnesia/` NO matchea `ProtegerPaqueteCerrado`; kit/doctrine + forjar-caja + metodología §9 con nota fechada | ✅ |
| Capabilities módulo `forja/` (CAP-149/150) + arch-lint componente `forja` | fitness ok · cap_doctor 152 válidas · go-arch-lint OK | ✅ |
| AC-e: init con el binario INSTALABLE sobre proyecto real | **Auto-verificado 2026-07-30 con el payload del `.deb` v0.6.0**: init sobre COPIA de `dogfood/dev-full-cycle` (proyecto real, aditivo puro) → sana · re-run 0 creados/todo ya-existía · borrar `01-spec.md` ⇒ `incompleta` + faltante listado + **exit 1** | ✅ |
| **Residuo humano: init sobre TU proyecto en TU laptop + firma** | | ⬜ **gate** |

Desviaciones/gaps declarados: doctor sin verificación de hashes (deriva de semilla =
Fase 2) · `--arnes-yaml` del proyecto = Fase 2 · nodos `.arnesia/` en Mapa = Fase 2 ·
spec corregido `{dir}`→`{path}` (coherencia con /portafolio/escaneos) · colisión
cap_num del merge reparada (149/150).

## Firma

- [ ] 🧑‍⚖️ init en proyecto real — fecha:
