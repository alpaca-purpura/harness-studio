# PARIDAD — Homologación de metodología

> Gate humano final 🧑‍⚖️. Cada fila = entregable ↔ verificación REAL ejecutada. Estado `✅` = hecho
> y verificado en vivo esta sesión. Las desviaciones se registran, jamás se maquillan.

## Matriz entregable ↔ verificación

| # | Entregable (adopción del método del plugin) | Verificación REAL | Estado |
|---|---|---|---|
| 1 | Seam `project.config.yaml` (12 slots Go+React+Tauri) + `scripts/harness_config.py` (copia project-layer del kit) | `python3 scripts/harness_config.py --doctor` → **exit 0**, slots resuelven | ✅ |
| 2 | Árbol `docs/{product,architecture,process}/` + READMEs de propiedad (R1/R2/R3) | árbol creado; READMEs declaran owner+convención | ✅ |
| 3 | Templates del plugin verbatim (`00-story`,`00-research`,`01-spec`,`story.yaml`) + `capability.template.yaml` homologado | copiados a `docs/product/_templates/` | ✅ |
| 4 | Skill **`/pm`** front-door seam-driven (bootstrap closure-gate + 10 estados + auto-chain) | bootstrap ejecutado contra estructura nueva (seam+scan+estado) | ✅ |
| 5 | `scripts/cap_doctor.py` (doctor local + `--index`) | corre; **82 caps válidas**; genera INDEX | ✅ |
| 6 | 3 ejes temporales → `docs/product/` (`checkpoint`/`BACKLOG`/`LEDGER`/`ledger/`) | `git mv`; `estado.sh`+router actualizados; 0 refs de código rotas | ✅ |
| 7 | **CAPABILITIES.md → 82 hojas YAML** por-cap (15 módulos) + `_coverage.yaml` | `capabilities_to_yaml.py` determinista; `cap_doctor` 82 ✓ | ✅ |
| 8 | Enforcer R1/R2 **flipeado a leer YAML** (`capability_trace_test.go`) + CAPABILITIES.md retirado → `INDEX.md` | `go test ./docs/architecture/fitness/ -run TestCapability` → **ok** | ✅ |
| 9 | **`arch/`+`knowledge/`+`STACK.md` → `docs/architecture/`** (embed-coupled: go:embed+parser+go-arch-lint+schema+~115 refs) | `go build`✓ · `go test ./...` **11 ok/0 FAIL** · `go-arch-lint` **OK** | ✅ |
| 10 | Motor de conformance lee el ruleset reubicado | `conformance --todo` **247** · `--arnes` **21·20·1** (idénticos al baseline) | ✅ |
| 11 | `historias/` → `docs/product/stories/` (10 paquetes) + records sueltos → `docs/product/research/` (10) | `git mv`; refs Go/config/router actualizadas; 0 colgantes | ✅ |
| 12 | Router `CLAUDE.md` reapunta a `docs/` + `/pm` | tabla "Necesito X → leo Y" actualizada | ✅ |
| 13 | **VISION/METODOLOGIA/UX raíz → `docs/`** (D8: `product/vision.md` · `product/ux.md` · `process/metodologia.md`) + 33 links markdown recomputados + schema/seam/`/pm`/README/CLAUDE cableados | 0 links vivos rotos · `--todo` **247·40·0** · `--arnes` **21·20·1** · build/test/`go-arch-lint`/`cap_doctor`/`--doctor` verdes (== baseline) | ✅ |

## Desviaciones registradas (se consultan, jamás se maquillan)

1. **`harness_config.py` es una COPIA project-layer del kit** (ADOPTING §1 `cp -r`), no un consumo
   directo del plugin — porque en modo-plugin el loader parent-walkea desde su propia ubicación y no
   ve el config del repo. Fix va upstream al kit (soporte plugin-mode del seam).
2. **CAPABILITIES.md se retiró y el árbol YAML es la fuente**, pero el `estado` de cada cap se migró
   con el valor que tenía en el MD (aún no re-derivado por `arnesia conformance`; R4 sigue `pendiente`
   como estaba). Deuda pre-existente, no introducida.
3. **`docs/product/research/` tiene records sueltos** (planes/auditorías 2026-07-04…07-06) que en el
   modelo estricto del plugin serían per-story `00-research` — se dejaron como research cross-cutting
   (honesto; reorganizables luego).
4. **Arneses secundarios `/po` `/architect` `/dev-team` `/auditor` NO forjados** — `/pm` los referencia
   (auto-chain) pero aún no existen; declarado en la skill. Se forjan + upstream en paquete siguiente.
5. **Cifras de arch/knowledge en checkpoint** aún tecleadas (deuda pre-existente en BACKLOG).
6. **`CLAUDE-nuevo.md`** (draft superseded en `stories/2026-07-09-reorg-docs/`) tiene 4 links
   `./VISION.md`… rotos — **pre-existentes a D8** (era draft root-relative parqueado en subdir → ya
   rotos; además apunta a `CAPABILITIES.md`/`ESTADO.md`/`historias/` muertos). Snapshot histórico; no
   se poli-parcha un draft íntegramente muerto. Su limpieza/borrado = paquete `reorg-docs`.

## Fuera de alcance de esta sesión (paquetes siguientes)

- Upstream del método (`/pm` + `cap_doctor` + scaffolder + capability schema) al kit `harness@prenter-marketplace` (nueva versión).
- Replicar `docs/` + seam + `/pm` a **cockpit** y **dev-studio**.
- Forjar los 4 arneses de rol secundarios.
- ~~Mover `VISION.md` / `METODOLOGIA.md` / `UX.md` raíz → `docs/`~~ → **HECHO** (D8/Task 9, fila 13).

## Firma

- [ ] 🧑‍⚖️ **Gate humano** — revisar `docs/` + correr `/pm` + confirmar la matriz. `chris_verify.signoff → true`.
