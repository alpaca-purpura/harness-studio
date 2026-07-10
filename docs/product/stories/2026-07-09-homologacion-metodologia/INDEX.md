# Homologación de metodología (3 repos) — INDEX

**Slug:** `2026-07-09-homologacion-metodologia` · **Estado global:** 🟢 `done` — firmada + cerrada (HS-19, 2026-07-09)

## Marco

Homologar la forma de crear software (CLAUDE.md, skills, docs) entre **harness-studio**,
**cockpit**, **dev-studio**; referencia = **luana-vitalia/vitalia**. Target: una sola carpeta
`docs/{product,architecture,projects/{epics,stories,research}}` + skill `/pm` + arneses de
proceso. Esta sesión: construirlo **bien en harness-studio primero**; replicar después.

## Flujo de gates 🧑‍⚖️

1. Exploración 4 repos — ✅ hecha (2026-07-09)
2. Cruxes D1–D8 firmados — ✅ (ver `decisiones.md`)
3. Spec del target + plantillas — ✅
4. Implementar (adopción plena D7) — ✅ Tasks 1–9
5. PARIDAD / gate humano — ✅ **FIRMADO 2026-07-09 (HS-19)** tras corregir el drift de 65 links

## Estado

- [x] Barrido de los 4 repos
- [x] D1–D7 firmadas (ver `decisiones.md`)
- [x] **Task 1** — seam `project.config.yaml` + `--doctor` verde ✓
- [x] **Task 2** — árbol `docs/{product,architecture,process}` + READMEs + templates del plugin ✓
- [x] **Task 3** — `cap_doctor.py` + schema capability homologado ✓
- [x] **Task 4** — skill `/pm` (bootstrap verificado contra estructura nueva) ✓
- [x] **Task 5** — ejes temporales → `docs/product/` (checkpoint/BACKLOG/LEDGER/ledger) + refs ✓
- [x] **Task 6** — `CAPABILITIES.md` → **82 YAML por-cap** + enforcer R1/R2 flipeado a YAML + INDEX generado — **go test ./... 11 ok / 0 FAIL** ✓
- [x] **Task 7** — `arch/`+`knowledge/`+`STACK.md` → `docs/architecture/` (go:embed + parser + go-arch-lint + schema + ~115 refs) — **build/test/lint/conformance idénticos al baseline** ✓
- [x] **Task 8** — `historias/` → `docs/product/stories/` (10) + records → `docs/product/research/` (10) + `checkpoint.md` + `PARIDAD.md` ✓
- [x] **Task 9 (D8)** — `VISION/METODOLOGIA/UX` raíz → `docs/` (`product/vision.md` · `product/ux.md` · `process/metodologia.md`) + 33 links recomputados + schema/seam/`/pm`/README/CLAUDE cableados — **build/test/`go-arch-lint`/conformance `--todo` 247·40·0/`--arnes` 21·20·1 == baseline** ✓
- [x] 🧑‍⚖️ **Gate humano** — `chris_verify.signoff: true` (2026-07-09, HS-19; ver `PARIDAD.md` fila 14 + firma)
- [x] **Cierre HS-19** — gate en vivo destapó 65 links vivos rotos (claim Task 9 falso) → corregidos + re-verificado 0 rotos, cifras sin regresión

## Retomar aquí

**CERRADO (HS-19, 2026-07-09).** Homologación en harness-studio COMPLETA, verde y COHERENTE: método
del plugin adoptado, TODA la doc en `docs/`, y el drift de links que la migración había dejado
(65 vivos rotos, pese al claim «0 rotos») corregido y re-verificado. Gate PARIDAD 🧑‍⚖️ firmado; ficha
`ledger/HS-19.md`. **Paquetes siguientes (en `BACKLOG.md`):** (a) upstream del método al kit
`harness@prenter-marketplace` (nueva versión); (b) replicar `docs/`+seam+`/pm` a **cockpit** y
**dev-studio**; (c) forjar arneses secundarios `/po`·`/architect`·`/dev-team`·`/auditor`.
