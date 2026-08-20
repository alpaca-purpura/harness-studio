---
story_id: 2026-07-09-estado-ci-drift-gate
state: done               # idea→refining→refined→ready→developing→developed→reviewing→[done]
release: v0.7-mvp-instalable
module: conformance
cap_target: ninguno (infra/honestidad — fuera del universo R2; ver decisiones D-módulo)
chris_verify:
  signoff: true           # ✅ FIRMADO 2026-07-10 («firma y commitea directo a main»)
defer_audit: false
last_audit: 2026-07-10     # 2 auditorías adversariales (subagentes) → MERGEABLE
ledger: HS-21
---

# checkpoint — estado.sh → CI drift-gate (presente)

## Estado

`reviewing` — código IMPLEMENTADO, auto-validado (V1-V4) y auditado (subagente adversarial →
MERGEABLE). `estado.sh --check` rompe el merge si el bloque `<!--stats-->` de `checkpoint.md`
queda stale vs el estado real. Neutraliza la fecha (procedencia, no cifra). 3 superficies:
`scripts/estado.sh` (+`--check`) · `.github/workflows/ci.yml` (step en job `go`) ·
`docs/architecture/conventions/ci.md` (fila `ci-estado-drift`). Sin capability (fuera R2).

## dod_evidence (verificación REAL)

- **V1** repo sync → `estado.sh --check` **exit 0**.
- **V2** drift real (checkpoint 247 vs real 248; boundaries 17→18) → **exit 1** con unified-diff.
- **V3** mutar SOLO la fecha `medido <ISO>` → **exit 0** (sin falso positivo).
- **V4** `go build`/`go vet`/`go test ./...`/fitness → **0 FAIL**. Ruleset 247→**248** (la fila
  nueva de `ci.md` entra como check `deferred`, mecanismo CI) → checkpoint regenerado self-consistent.
- Auditoría adversarial (subagente): **MERGEABLE**; 2 hallazgos BAJA resueltos (regex ISO-anclado ·
  esta ficha + BACKLOG).
- **Δ1 resuelto (D5):** hook `pre-commit estado-cifras` auto-regenera las cifras en el commit.
  Verificado con **commit REAL**: drift 18→17 → el commit salió con trigger + checkpoint fresco.
  `pre-push` y `stage_fixed` descartados con evidencia E2E. Ruleset 248→**249** (fila git-hooks.md).

## Retomar aquí

**CERRADO (HS-21, 2026-07-10).** Gate PARIDAD 🧑‍⚖️ firmado (10 filas + 5 desviaciones). Commiteado a
`main`. Siguientes del eje N2 «honestidad automática» (→ BACKLOG): **Paquete B** derivación LIVE
(`vivo ⟺ check verde corriendo`) · **Paquete C** validar los ~40 `vivo·nc` (test por-cap).
