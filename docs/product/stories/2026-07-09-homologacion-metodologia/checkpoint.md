---
story_id: 2026-07-09-homologacion-metodologia
state: done               # idea→refining→refined→ready→developing→developed→reviewing→[done]
module: kit
cap_target: método-homologado (transversal)
chris_verify:
  signoff: true           # ✅ gate firmado 2026-07-09 (orden del operador «cerrá coherente») tras corregir el drift de links (ver PARIDAD.md)
defer_audit: false
last_audit: 2026-07-09
ledger: HS-19
---

# checkpoint — Homologación de metodología (3 repos)

## Estado

`done` — método del plugin adoptado y migrado EN VIVO en harness-studio; gate PARIDAD firmado.
**Cierre (HS-19):** la verificación en vivo del gate destapó que el claim «0 links vivos rotos» de
Task 9 era FALSO — la migración recomputó links incompleta: **65 links markdown rotos en 13 docs
vivos** (`docs/process/metodologia.md` 36 · `docs/architecture/INDEX.md` 9 · …). Causa: reescritura
`arch/`→`docs/architecture/` con prefijo `./` relativo-a-raíz (no relativo al archivo movido) +
renames no propagados (ESTADO→checkpoint · CAPABILITIES→capabilities/ · LEDGER/BACKLOG/STACK). **Los
65 corregidos + ~11 punteros stale sincronizados; re-verificado 0 links vivos rotos.** 30 links en
snapshots históricos (`stories/`·`research/`·`ledger/`) se dejan a propósito (desviación #7).
**Sigue pendiente en OTRO paquete:** replicación a cockpit/dev-studio + upstream del método al plugin.

## dod_evidence (verificación REAL, no "compila")

- `--doctor` seam → **exit 0** (12 slots llenos).
- `/pm` bootstrap → corre contra estructura nueva (seam + closure-gate scan + checkpoint/BACKLOG).
- `cap_doctor.py` → **82 capabilities válidas**.
- `go build ./...` ✓ · `go vet ./...` ✓ · `go test ./... ` → **11 paquetes ok / 0 FAIL**.
- `go-arch-lint` (config reubicado) → **OK, sin warnings**.
- `conformance --todo` → **247 checks · pass 40 · fail 0 · deferred 207** (idéntico al baseline).
- `conformance --arnes dogfood` → **21 · pass 20 · fail 1** (idéntico al baseline).
- `estado.sh` (pipeline E2E) → conformance lee ruleset reubicado y regenera checkpoint. ✓
- **D8/Task 9 (VISION/METODOLOGIA/UX → `docs/`):** 3 `git mv` + links recomputados +
  schema/seam/`/pm`/README/CLAUDE cableados → `--todo` **247·40·0** · `--arnes` **21·20·1** ·
  `go build`/`go test`/`go-arch-lint`/`cap_doctor`/`--doctor` verdes (idénticos al baseline).
  ⚠ CORRECCIÓN HS-19: el «0 links vivos rotos» de Task 9 no era cierto — ver Cierre abajo.
- **Cierre (HS-19, gate en vivo):** re-corrí TODO contra el binario/repo real (no confié en el
  checkpoint). `--doctor` 0 · `cap_doctor` 82 · `go build/vet/test` 11 ok/0 FAIL · `--todo`
  **247·40·0·207** (verificado 2 vías: texto + JSON) · `--arnes dogfood/dev-full-cycle.graph.json`
  **21·20·1** · rutas viejas TODAS ausentes en raíz (solo `CLAUDE.md`+`README.md`+`docs/`). Único
  hueco: **65 links vivos rotos** → corregidos con fixer determinista (0 unresolved) + 11 punteros
  stale sincronizados + doctrina `codigo-traza-a-capability` v1.2 (SSoT `CAPABILITIES.md`→árbol YAML,
  solo prosa, cifras intactas). Re-verificado post-fix: **0 links vivos rotos**, cifras sin regresión.

## Retomar aquí

**CERRADO (HS-19).** Homologación en harness-studio COMPLETA, verde y coherente (Tasks 1–9 + fix del
drift de links del cierre). Gate PARIDAD 🧑‍⚖️ firmado. Próximos paquetes (en `BACKLOG.md`): (a) upstream
del método al kit `harness@prenter-marketplace` (nueva versión); (b) replicar `docs/` + seam + `/pm`
a cockpit y dev-studio; (c) forjar los arneses secundarios `/po` `/architect` `/dev-team` `/auditor`.
