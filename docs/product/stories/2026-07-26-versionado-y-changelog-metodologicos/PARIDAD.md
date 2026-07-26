# PARIDAD — Versionado y changelog metodológicos (RF-232)

> Paquete `2026-07-26-versionado-y-changelog-metodologicos`. Cada RF ⇒ código real + su
> verificación; cada verificación ⇒ corrida de verdad, con su salida. **Fecha: 2026-07-26.**
> Firma 🧑‍⚖️ del gate: **PENDIENTE**.

## Fila por fila

| Qué exige el RF | Código | Verificación | Estado |
|---|---|---|---|
| existe el changelog, con forma | `CHANGELOG.md` | `python3 scripts/changelog.py check --exige-entradas` | ✅ `changelog ✓ (convención desde 0.2.22)` |
| el bump publica y promueve | `scripts/bump.sh` · `Makefile:bump-patch` | `make bump-patch` real (y revertido después) | ✅ `version: 0.2.21 -> 0.2.22` · `changelog: [Sin publicar] → [0.2.22] — 2026-07-26` |
| los 3 manifiestos quedan en lockstep | idem | `grep` de los 3 tras el bump | ✅ `0.2.22` en `Cargo.toml`, `tauri.conf.json`, `package.json` |
| **sin entradas NO se puede publicar** | `bump.sh` (valida antes de escribir) | segundo `make bump-patch` con `[Sin publicar]` vacía | ✅ `Error 1` + `[Sin publicar] está vacía…`; `Cargo.toml` **quedó en 0.2.22**, no subió a 0.2.23 |
| minor/major existen y funcionan | `Makefile:bump-minor|bump-major` | `make bump-minor` real | ✅ `version: 0.2.22 -> 0.3.0` + sección `## [0.3.0] — 2026-07-26` |
| `add` con alias e idempotente | `changelog.py:cmd_add` | `add Corregido …` y luego `add corregido …` (mismo texto) | ✅ 1ª: `Corregido: …` · 2ª: `ya estaba en Corregido — no se duplica` |
| categoría inventada se rechaza | `changelog.py:ALIAS/CATEGORIAS` | `add Varios "algo"` | ✅ exit 1 + `Usá una de: Agregado · Cambiado · Deprecado · Eliminado · Corregido · Seguridad` |
| CI agarra el bump a mano | `changelog_test.go` | `go test -run TestChangelog` con los manifiestos en 0.2.22 | ✅ `PASS` (con 0.2.21 → `SKIP` honesto por `convencion-desde`) |
| el gate real se auto-verifica | `changelog_test.go:TestChangelogScriptSeValidaASiMismo` | ejecuta `changelog.py check` desde el test | ✅ pass |
| tercera capa local | `lefthook.yml` job `changelog` | glob limitado a los 3 manifiestos → corre los 2 tests de forma/cobertura | ✅ declarado (glob verificado por lectura; no se disparó un commit real) |
| la convención lo dice | `versionado.md` v1.2 | 4 checks nuevos + sección de identidad de build | ✅ `enforced_by` extendido a 7 entradas |
| una sesión nueva lo lee sin que se lo digan | `CLAUDE.md` (regla dura + fila) · `metodologia.md` §10 | lectura | ✅ |

## Suites

| Suite | Resultado |
|---|---|
| `go test ./docs/architecture/fitness/ -run TestChangelog` | **2 pass · 1 skip honesto** (con 0.2.21) → **3 pass** (con 0.2.22) |
| `go test ./docs/architecture/fitness/ -run TestVersionManifestsInSync` | pass durante todo el ciclo de prueba |
| `python3 scripts/changelog.py check` | ✓ |
| `gofmt` / `go vet` sobre `changelog_test.go` | limpio |

## Ciclo E2E corrido de verdad (y revertido)

```
make bump-patch   → 0.2.21 → 0.2.22   + [Sin publicar] promovida         ✅
go test           → TestChangelogCubreLaVersionDeLosManifiestos: PASS    ✅
make bump-patch   → ABORTA (changelog vacío); Cargo.toml sigue en 0.2.22 ✅
changelog.py add  → entrada nueva (+ alias + idempotencia)               ✅
make bump-minor   → 0.2.22 → 0.3.0    + sección [0.3.0]                  ✅
restore backup    → 0.2.21 y CHANGELOG.md como estaban                   ✅
```

**Estado tras la prueba:** los 3 manifiestos en `0.2.21`, `CHANGELOG.md` con `[Sin publicar]`
llena y el bloque `[0.2.21] y anteriores`; fitness verde. Backup en el scratchpad de la sesión.

## ⛔ Lo que NO se probó

- **El hook de lefthook disparándose en un commit real.** El job está declarado con glob a los 3
  manifiestos y corre los mismos tests ya verificados a mano, pero no se hizo un commit de prueba
  (nada de este trabajo está commiteado todavía).
- **Un release completo de punta a punta** (`make installer` → `.deb` → instalar): pide rust/Tauri
  y sudo. Es el gate AC-9, del operador.
- **`changelog.py release` sobre una versión ya existente** (Bif-3): la rama está escrita y leída,
  pero no se forzó el caso a mano.
