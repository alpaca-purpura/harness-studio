# Decisiones — estado.sh → CI drift-gate

> Paquete de trabajo (eje N2, Paquete A). Regla §10: toda decisión conversada se escribe acá
> EN EL MISMO TURNO. Firmas 🧑‍⚖️ entre etapas.

## Norte pedido por el operador (cita cruda)

- «arranca y arregla todo con mucho cuidado, investigando antes, poniendo tus propios métodos de
  validación y verificación de que lo hiciste bien y al final auditas. Trata de usar subagentes
  para no llenar el contexto de esta conversación donde sea necesario.»
- Eje elegido: **N2 — endurecer honestidad** (cablear estado.sh + derivación LIVE + validar
  vivo·nc a CI). Este paquete A = solo el primer item (estado.sh → CI drift-gate).

## Preguntas abiertas (a resolver con la investigación)

- **D1 — gotcha del date-stamp:** `estado.sh` embebe `STAMP=$(git log -1 --date=short)` en el
  bloque. ¿Un commit de OTRO día (sin cambio de código) haría que CI regenere una fecha distinta →
  diff → rojo? Si sí, el gate sería demasiado agresivo. Opciones: (a) ignorar la fecha en el
  check; (b) checar solo las cifras (números), no el stamp; (c) modo `--check` en estado.sh que
  normaliza. → **decidir tras investigación.**
- **D2 — dónde vive el job:** ¿fold como step del job `go` (ya tiene Go; ubuntu trae python3) vs
  job nuevo `docs-stats` (duplica setup-go)? → tentativa: **fold en `go`**.
- **D3 — lefthook sí/no:** `estado.sh` corre `go run` ×2 (segundos). ¿Muy lento para pre-commit
  (filosofía = feedback rápido, el job `capabilities` es ~4ms)? → tentativa: **CI-only, el gate
  DURO es CI (doctrina ci.md); lefthook queda fuera**.
- **D4 — modo `--check`:** ¿estado.sh necesita un flag que NO escriba y solo compare+exit1, o el
  patrón `regenera && git diff --exit-code` (como tokens-sync) basta? → decidir según idempotencia.
- **D-módulo:** `cap_target` real + módulo del capability (conformance vs indice). → confirmar.

## Investigación (subagente, 2026-07-09) — hallazgos con evidencia

- `estado.sh` escribe **SOLO** `docs/product/checkpoint.md`, solo el bloque `<!--stats … /stats-->`
  (un `open(p,'w')`, `re.subn count=1`); resto del archivo byte-idéntico. **Determinista/idempotente**
  (md5 del bloque idéntico ×3 corridas; 0 diff en HEAD limpio).
- **Date-stamp gotcha CONFIRMADO:** `STAMP=$(git log -1 --date=short)` = committer date de HEAD (líneas
  18, 54-55), inyectado como `medido {stamp}` en 2 líneas. Un commit de **otro día sin cambio de
  cifras** → diff de 2 líneas → exit 1. Falso positivo. **El gate DEBE ignorar la fecha.**
- Sin red, sin sidecar/daemon, solo `go run` (compila de fuente); python3 stdlib (`re/sys/os/json/
  glob/subprocess/Counter`), **cero pip**; deps (`jsonschema-go`, `yaml.v3`) ya cacheadas por el job
  `go`. `dogfood/dev-full-cycle.graph.json` trackeado. Nada falta si el step vive en el job `go`.
- **Es GENERATE-ONLY:** no sale ≠0 por drift; regenera y sale 0. La detección de drift debe venir de
  un `--check`/`git diff` aparte.
- **~22-24 s** por corrida (3× `go run`) → demasiado lento para lefthook pre-commit (el hook
  `capabilities` es ~4 ms).
- Alcance de R2 (`capability_trace_test.go`): universo = `cmd/`·`internal/`·`web/src/`·
  `web/src-tauri/src/` × `.go/.ts/.tsx/.rs`. **`scripts/*.sh` y `.github/` NO están** → tocar
  `estado.sh`+CI no crea huérfano ni exige capability de producto.

## Decisiones firmadas 🧑‍⚖️ (2026-07-09)

- **D1 — la fecha se NEUTRALIZA en el check.** El gate compara **cifras**, ignora `medido <fecha>`.
  Semántica honesta: la fecha = procedencia de la última regeneración (cuándo se midieron), **no está
  obligada a igualar HEAD**. Si las cifras no cambiaron, la fecha vieja sigue siendo verdad. Se
  implementa normalizando `medido [^,\n]+ → medido <DATE>` en ambos lados antes de comparar.
- **D2 — el job vive en `go`.** Ya tiene `setup-go` + cache de módulos; ubuntu-latest trae python3.
  Ponerlo en `ts` exigiría sumar `setup-go` (rechazado). Step después de `build`.
- **D3 — CI-only, sin lefthook.** 22 s es prohibitivo para pre-commit; el gate DURO es CI (doctrina
  `conventions/ci.md`: «el gate duro es CI, no el hook»). lefthook queda fuera.
- **D4 — modo `--check` en `estado.sh` (SÍ).** Como es generate-only, agrego `estado.sh --check`:
  regenera el bloque en memoria, normaliza la fecha (D1), compara contra el commiteado, imprime un
  unified-diff + hint y **exit 1** si hay drift de cifras; **exit 0** si están en sync; **NO escribe
  el árbol**. Más limpio que `regenera && git diff --exit-code` (que ensucia el working-tree del
  runner y encima rompe por la fecha). El modo default (sin flag) queda **idéntico** (regenera
  in-place) — retro-compatible.
- **D-módulo — NO se crea capability de producto.** `estado.sh`+`ci.yml` están fuera del universo R2
  (infra/maquinaria de honestidad, no código de producto). Registro honesto en: este paquete +
  `conventions/ci.md` (check nuevo `ci-estado-drift`) + `ledger/HS-21.md` + BACKLOG (item→done).
  Inflar capabilities con un puntero a un `.sh` sería deshonesto (las caps mapean código de producto).

## Superficie de cambio (spec mínima)

1. `scripts/estado.sh` — parseo de `--check` + rama de comparación en el python embebido (D4).
2. `.github/workflows/ci.yml` — step `estado-drift` en el job `go`, tras `build` (D2): `bash scripts/estado.sh --check`.
3. `docs/architecture/conventions/ci.md` — fila `ci-estado-drift` en la Checklist + línea L2 + changelog (doc-as-code del required check).

## D5 — Resolución de Δ1 «regenerarse solas» (2026-07-10, orden del operador «resolve la desviación antes de firmar»)

Investigación empírica (no supuestos — todo probado E2E con push/commit reales):

- **`pre-push` self-heal → DESCARTADO (no dispara).** Implementado y testeado con `git push` real a
  bare remotes: lefthook 1.13.6 lo **saltea** («no matching push files») en rama-nueva, up-to-date, y
  hasta en push trunk-based con archivos en rango. Solo corre con `--force` (que git no pasa). No
  verificable → shippearlo sería un self-heal falso. Revertido (revisa D3: la razón «CI-only» sigue
  válida para pre-push, pero por IMPOSIBILIDAD técnica, no solo por velocidad).
- **`stage_fixed: true` → DESCARTADO (no stagea).** Commit real de prueba: el hook regeneró el
  checkpoint pero `stage_fixed` **NO lo stageó** (solo re-stagea archivos YA staged que matchean el
  glob; checkpoint.md casi nunca está staged cuando el trigger es otro archivo).
- **`pre-commit` + `git add` explícito → ELEGIDO (funciona, verificado E2E).** Commit real de prueba:
  el hook detectó drift (18→17 boundaries), regeneró y `git add docs/product/checkpoint.md` → el
  commit salió con **2 files: trigger + checkpoint.md fresco**. Único mecanismo que dispara Y stagea
  confiable. Sin race (es el único comando que toca el index).

**Decisión firmada:** hook `pre-commit` **`estado-cifras`** (glob `*.{go,ts,tsx,rs,yaml,md}`): decide
con `estado.sh --check` (date-insensitive) → si driftaron, `estado.sh` + `git add` del checkpoint. El
operador aceptó el costo (~20-50s por commit que toque esos archivos) a cambio de «cifras nunca stale
en un commit». Saltable con `--no-verify`; **el gate DURO sigue siendo CI** (`estado.sh --check`).
Doc-as-code en `conventions/git-hooks.md` v1.1 (`lefthook-precommit-estado`) → ruleset 248→**249**.

**Δ1 RESUELTO:** el gate ya no solo detecta — las cifras se regeneran **solas** en el commit local, y
CI queda de backstop. (El único matiz que persiste: `--no-verify` saltea el hook, pero CI lo caza.)

## Validación propia planeada (métodos)

- **V1 no-drift:** repo limpio → `estado.sh --check` exit 0.
- **V2 drift real:** mutar una cifra en checkpoint.md → exit 1 (+ diff útil). Restaurar.
- **V3 no-falso-positivo por fecha:** mutar SOLO `medido <fecha>` en checkpoint.md → exit 0.
- **V4 sin regresión:** `go build ./...` · `go vet ./...` · `go test ./...` verdes; default `estado.sh`
  sigue regenerando idéntico (0 diff en HEAD limpio).
