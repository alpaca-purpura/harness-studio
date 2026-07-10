# PARIDAD — estado.sh → CI drift-gate (HS-21)

> Gate humano final 🧑‍⚖️. Cada fila = entregable ↔ verificación REAL ejecutada. `✅` = hecho +
> verificado en vivo. **FIRMADO 2026-07-10** por orden del operador («firma y commitea directo a main»).

## Matriz entregable ↔ verificación

| # | Entregable | Verificación REAL | Estado |
|---|---|---|---|
| 1 | `estado.sh --check` (regenera el bloque en memoria, normaliza fecha, compara vs commiteado, exit 1 drift / exit 0 sync, **NO escribe**) | **V1** repo en sync → **exit 0** ✓ · **V2** drift real (checkpoint 247 vs real 248, y boundaries 17→18) → **exit 1** con unified-diff útil ✓ | ✅ |
| 2 | Step `estado-drift` en el job **`go`** de `ci.yml`, tras `build` (`bash scripts/estado.sh --check`) | deps confirmadas por investigación+auditoría: `setup-go`✓ · python3 stdlib (0 pip)✓ · `git log` ok con checkout shallow✓ · `dogfood/dev-full-cycle.graph.json` trackeado✓ | ✅ |
| 3 | Neutralización de la fecha (D1) — sin falso positivo por commit de otro día | **V3** mutar SOLO `medido <fecha>` → **exit 0** ✓ · regex anclado a ISO `medido (\d{4}-\d{2}-\d{2}\|s/f)` (blindaje de auditoría) | ✅ |
| 4 | Doc-as-code: fila `ci-estado-drift` en `conventions/ci.md` + L2 + changelog (v1.0→**1.1**, 6 checks) | fila severidad `error` / enforcer `ci.yml (estado.sh --check)`; checklist = 6 filas; changelog coherente | ✅ |
| 5 | **Sin capability de producto** — `estado.sh`+`ci.yml` fuera del universo R2 | `capSourceExts={.go,.ts,.tsx,.rs}` × `{cmd,internal,web/src,web/src-tauri/src}`; `scripts/`+`.github/` NO cubiertos ⇒ 0 huérfano, 0 cap requerida (auditado) | ✅ |
| 6 | Retro-compat del modo default (sin flag) | rama `generate` byte-equivalente al original (`norm()` solo se llama en `check`; `SystemExit` antes de generate); `bash scripts/estado.sh` regenera idéntico | ✅ |
| 7 | Sin regresión | `go build`✓ · `go vet`✓ · `go test ./...` 0 FAIL · fitness ok · checkpoint self-consistent (247→**248**, deferred 205→**206**: el propio check nuevo entra `deferred`, mecanismo CI) | ✅ |
| 8 | Auditoría adversarial independiente (subagente) | veredicto **MERGEABLE**; falso-negativo/positivo/CI-wiring/determinismo/retro-compat/doc-as-code = 6/6 ✓; 2 hallazgos BAJA resueltos (regex blindado · ledger+BACKLOG en este cierre) | ✅ |
| 9 | **Δ1 resuelto — auto-regen `pre-commit estado-cifras`** (glob `*.{go,ts,tsx,rs,yaml,md}`; `--check`→regen→`git add`) | **commit REAL de prueba**: drift 18→17 boundaries → el commit salió con **2 files** (trigger + checkpoint.md fresco) ✓. `pre-push` (lefthook lo saltea) y `stage_fixed` (no stagea cross-file) DESCARTADOS con evidencia E2E — ver `decisiones.md` D5 | ✅ |
| 10 | Doc-as-code del hook: `conventions/git-hooks.md` v1.1 (`lefthook-precommit-estado`) | fila sev `warn` / enforcer `/lefthook.yml`; ruleset 248→**249**, checkpoint regenerado self-consistent | ✅ |

## Desviaciones registradas (se consultan, jamás se maquillan)

- **Δ1 — RESUELTO (D5, orden «resolvé antes de firmar»).** Ya NO es «detecta pero no regenera»: el
  hook `pre-commit estado-cifras` regenera+`git add` el checkpoint en el commit (cifras «solas»).
  Camino recorrido con evidencia E2E: `pre-push` descartado (lefthook lo saltea), `stage_fixed`
  descartado (no stagea cross-file), `pre-commit`+`git add` elegido y verificado. Costo aceptado por
  el operador: ~20-50s por commit relevante. CI `--check` = backstop para `--no-verify`.
- **Δ2 — el ruleset pasó 247→248.** Agregar la fila `ci-estado-drift` a `ci.md` sumó 1 check
  (deferred, mecanismo CI). Efecto esperado y self-consistent: el checkpoint se regeneró en el mismo
  paquete, así el gate queda verde.
- **Δ3 — `norm()` normaliza SOLO la fecha.** Blindado a fecha ISO tras la auditoría para que un
  template futuro con cifras tras «medido» no pueda enmascararse.
- **Δ4 — el hook `pre-commit` corre en casi todo commit (~20-50s).** El glob `*.{go,ts,tsx,rs,yaml,md}`
  matchea la mayoría de commits → 3× `go run` cada vez. Costo aceptado por el operador (D5) a cambio
  de «cifras solas»; saltable con `--no-verify`; el gate duro es CI.
- **Δ5 — blind-spots LOCALES del hook (backstopeados por CI).** El glob no cubre `.json` (el feed
  `dogfood/dev-full-cycle.graph.json` de `--arnes`), y `--check` lee el árbol-de-trabajo, no lo staged.
  ⇒ hay commits que localmente podrían no auto-curarse. **Sin hueco de sistema:** el CI `estado.sh
  --check` regenera TODAS las cifras desde el estado real y rompe el merge. El hook es best-effort;
  CI es la garantía.

## Fuera de alcance (siguientes de N2, → BACKLOG)

- Derivación **LIVE** (`vivo ⟺ check verde corriendo`) → CI (Paquete B del eje N2).
- Validar los ~40 `vivo·nc` (test por-cap) (Paquete C).

## Firma

- [x] 🧑‍⚖️ **Gate humano** — **FIRMADO 2026-07-10** por orden del operador («firma y commitea directo
  a main»). Respaldo: 10 filas ✅ verificadas en vivo · Δ1 resuelto con auto-cura `pre-commit` probada
  con commit REAL · V1-V4 (drift real→exit1, solo-fecha→exit0, build/test 0 FAIL) · **2 auditorías
  adversariales independientes → MERGEABLE** (detect-gate + hook completo, 7/7 y 6/6 ✓). 5 desviaciones
  aceptadas (todas backstopeadas por el CI `--check`). Cero pass fabricado.
