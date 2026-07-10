---
story_id: 2026-07-09-homologacion-metodologia
state: developed          # idea→refining→refined→ready→developing→[developed]→reviewing→done
module: kit
cap_target: método-homologado (transversal)
chris_verify:
  signoff: false          # ⏳ gate humano pendiente (ver PARIDAD.md)
defer_audit: false
last_audit: 2026-07-09
---

# checkpoint — Homologación de metodología (3 repos)

## Estado

`developed` — método del plugin adoptado y migrado EN VIVO en harness-studio, todo verde.
**Falta:** gate humano de PARIDAD (`chris_verify.signoff`) + replicación a cockpit/dev-studio
(paquete siguiente) + upstream del método al plugin (nueva versión).

## dod_evidence (verificación REAL, no "compila")

- `--doctor` seam → **exit 0** (12 slots llenos).
- `/pm` bootstrap → corre contra estructura nueva (seam + closure-gate scan + checkpoint/BACKLOG).
- `cap_doctor.py` → **82 capabilities válidas**.
- `go build ./...` ✓ · `go vet ./...` ✓ · `go test ./... ` → **11 paquetes ok / 0 FAIL**.
- `go-arch-lint` (config reubicado) → **OK, sin warnings**.
- `conformance --todo` → **247 checks · pass 40 · fail 0 · deferred 207** (idéntico al baseline).
- `conformance --arnes dogfood` → **21 · pass 20 · fail 1** (idéntico al baseline).
- `estado.sh` (pipeline E2E) → conformance lee ruleset reubicado y regenera checkpoint. ✓
- **D8/Task 9 (VISION/METODOLOGIA/UX → `docs/`):** 3 `git mv` + 33 links markdown recomputados +
  schema/seam/`/pm`/README/CLAUDE cableados → `--todo` **247·40·0** · `--arnes` **21·20·1** ·
  `go build`/`go test`/`go-arch-lint`/`cap_doctor`/`--doctor` verdes (idénticos al baseline) · 0 links vivos rotos.

## Retomar aquí

Todo lo migrable en harness-studio está hecho y verde (Tasks 1–9; D8/Task 9 = VISION/METODOLOGIA/UX
→ `docs/`, HECHO). Falta SOLO la firma 🧑‍⚖️ del gate (PARIDAD.md). Próximos paquetes: (a) upstream
del método al kit `harness@prenter-marketplace` (nueva versión); (b) replicar `docs/` + seam + `/pm`
a cockpit y dev-studio; (c) forjar los arneses secundarios `/po` `/architect` `/dev-team` `/auditor`.
