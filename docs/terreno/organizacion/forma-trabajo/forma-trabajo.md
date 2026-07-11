---
id: organizacion/forma-trabajo
tipo: reference            # Diátaxis
dimension: forma-trabajo
territorio: organizacion
zachman: How
resumen: "Convenciones de código + disciplina PARIDAD/trunk-based/paquete que todo trabajo respeta."
activacion: always
punteros_auto:             # hoy sembrados a mano (golden); auto-derivables por scraping (P4), verificados en CI
  - docs/architecture/conventions/INDEX.md
  - docs/architecture/conventions/commits.md
  - docs/architecture/conventions/go-style.md
  - docs/architecture/conventions/ts-style.md
  - docs/architecture/conventions/naming.md
  - docs/architecture/conventions/git-hooks.md
  - docs/architecture/conventions/ci.md
  - docs/process/metodologia.md#10-disciplina-de-desarrollo-por-paquete-de-trabajo
  - lefthook.yml
budget_tok: 360
---

# forma-trabajo — cómo se trabaja en ArnesIA (NORMA)

La **forma de trabajo** es durable y la comparte todo el proyecto. Es cara-NORMA (D18): el *cuándo* se
aplica vive en `proceso` (cajas build/paridad), el *artefacto* de cada paquete vive en WIP. Enforcement
automático → [`knowledge/`](./knowledge/INDEX.md).

## Convenciones de código (SSoT durable)
Viven as-code en **[`docs/architecture/conventions/`](../../../architecture/conventions/INDEX.md)** — no se duplican aquí:
- **Go** → `go-style.md` · **TS/React** → `ts-style.md` + `ts-types.md` · **naming** → `naming.md`
- **commits** (Conventional + `Co-Authored-By`) → `commits.md` · **git-hooks** (lefthook) → `git-hooks.md`
- **CI** → `ci.md` · **editor** (Biome, format-on-save) → `editor.md` · **cadencia** → `CADENCE.md`

## Disciplina de proceso (durable)
- **PARIDAD** — nada se da por «listo» sin verificar en vivo contra el binario/instalado; desviaciones
  se firman 🧑‍⚖️, no se esconden (metodología §3 contrato de caja + §4 honestidad).
- **Paquete de trabajo** — toda feature: mockup→decisiones→spec→implementar→PARIDAD, con firmas entre
  etapas; decisión conversada → `decisiones.md` en el mismo turno (metodología §10).
- **Trunk-based** — `main` única; commit/push directo; tags semver en releases (CLAUDE.md §Reglas duras).
- **Honestidad automática (N2)** — cifras se GENERAN (`scripts/estado.sh`), nunca se teclean; drift-gate
  en CI + auto-cura `pre-commit` (HS-21).
- **Capabilities = SSoT funcional** — ningún cambio de código sin construir/modificar un capability (R1-R4).

## Traza
Esta dimensión **no aporta capabilities** (Producto=Σ caps; Organización define el cómo). Su cumplimiento
se verifica por los enforcers de `knowledge/` (hooks + CI), no por un check de capability.
