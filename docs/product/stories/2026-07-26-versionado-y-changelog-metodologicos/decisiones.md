# Decisiones — Versionado y changelog metodológicos (2026-07-26)

> Orden del operador, literal: *«arregla todo para que siga las mejores prácticas y nosotros
> construyamos tranquilos sin que tenga que recordarte que debes cambiar y actualizar el
> versionamiento y que como parte del mismo debes también agregar qué hay de nuevo / se corrige /
> se elimina en la nueva versión. Ya debe ser metodológico, no debo andar diciéndote nada.»*

## El agujero (medido)

| Qué | Estado antes de este paquete |
|---|---|
| Sello de build `AAMMDDHHMM` | ✅ automático en todo `bundle.sh` (RF-231, ayer) |
| Bump de semver | ⚠️ automático **solo** en `make installer` (`installer: bump-patch`), y solo PATCH |
| Minor / major | ❌ a mano, sin target ni criterio escrito |
| **Qué trae cada versión** (agregado/corregido/eliminado) | ❌ **no existe**: `ls CHANGELOG*` → nada |
| La convención `versionado.md` (v1.1, `enforced`) | ⚠️ no menciona el sello de build ni ningún changelog — `grep sello\|ldflags\|RF-231` sale vacío |

El pedido no es «escribí un changelog»: es que **el proceso no dependa de que alguien se acuerde**.
Un archivo nuevo sin dientes se cae en tres sesiones.

## VC-D1 · El changelog es `CHANGELOG.md` en la raíz, formato Keep a Changelog 1.1.0 — **DECIDIDA**

Estándar de industria, legible por humanos y parseable, con las 6 categorías canónicas
(traducidas: **Agregado · Cambiado · Deprecado · Eliminado · Corregido · Seguridad**).

**Por qué no reusar el LEDGER:** son dos ejes distintos y confundirlos rompe los dos. El
`LEDGER.md` es historia de **decisiones internas** por ficha `HS-NN` («por qué elegimos esto»);
el changelog es el contrato **por versión publicada** («qué cambió en el binario que instalaste»).
Un usuario del `.deb` no lee fichas HS.

## VC-D2 · El bump PROMUEVE `[Sin publicar]` — versionar sin changelog es imposible, no «desaconsejado» — **DECIDIDA**

El bump deja de ser «tocar 3 manifiestos» y pasa a ser un acto atómico:

```
make bump-patch → 1. valida el changelog (falla si [Sin publicar] está vacío)
                  2. sincroniza Cargo.toml + tauri.conf.json + package.json
                  3. promueve [Sin publicar] → ## [X.Y.Z] — AAAA-MM-DD
```

Si no hay nada escrito en `[Sin publicar]`, **el bump falla y no toca ningún archivo**. Es la
diferencia entre una norma y un gate: una norma se olvida, un gate se choca.
`make installer` hereda el gate porque depende de `bump-patch`.

## VC-D3 · Tres targets con criterio escrito, no solo patch — **DECIDIDA**

`make bump-patch` · `make bump-minor` · `make bump-major`. El default de `make installer` sigue
siendo **patch** (no cambia el hábito vigente). El criterio deja de vivir en la cabeza:

| Bump | Cuándo | Ejemplo real de este repo |
|---|---|---|
| PATCH | fix compatible, sin superficie nueva | el fix de `PATH` en el self-update |
| MINOR | superficie nueva compatible | la identidad de build en Ajustes (RF-231) |
| MAJOR | rompe algo que el usuario ya usaba | mover `~/.local/bin/arnesia` de lugar |

## VC-D4 · La historia previa NO se reconstruye — marcador `convencion-desde` — **DECIDIDA**

Hay **20 releases** en `instaladores/` (v0.2.2 … v0.2.21) sin changelog. Redactarlos hoy sería
inventar: nadie puede jurar qué entró en v0.2.7. El archivo lleva
`<!-- convencion-desde: 0.2.22 -->` y **dice** que lo anterior vive en git y en el LEDGER. El
enforcement solo exige entradas de esa versión en adelante.

Es la regla de honestidad §4 aplicada al changelog: un hueco declarado, jamás un relleno plausible.

## VC-D5 · Tres capas de enforcement, porque una sola no aguanta — **DECIDIDA**

| Capa | Qué agarra | Cuándo muerde |
|---|---|---|
| `scripts/changelog.py` | bump sin entradas · formato roto · versión duplicada | al correr `make bump-*` (bloquea) |
| `docs/architecture/fitness/changelog_test.go` | la versión de los manifiestos sin su sección · categoría inventada · sección vacía | `go test ./docs/architecture/fitness/` → **CI** |
| `lefthook.yml` job `changelog` | commitear un bump de manifiesto sin changelog coherente | `pre-commit`, solo si tocás los 3 manifiestos |

La doc sola (v1.1 lo probó) no alcanza: describía un mundo de solo-semver y nadie lo notó por un día.

## VC-D6 · La entrada se escribe en el MISMO turno que el código, como las decisiones — **DECIDIDA**

Mismo mecanismo que ya funciona en §10 para `decisiones.md`. Un comando de una línea para que no
haya excusa de fricción:

```bash
python3 scripts/changelog.py add Agregado "identidad de build en Ajustes: semver + sello"
```

Y la regla queda escrita donde una sesión nueva la lee **sin que el operador la dicte**:
`CLAUDE.md` (router, reglas duras) + `metodologia.md` §10 (flujo del paquete) +
`versionado.md` v1.2 (la convención, con checklist evaluable).

## VC-D7 · La convención sube a v1.2 y absorbe el sello de build — **DECIDIDA**

`versionado.md` (v1.1, `status: enforced`) no menciona la identidad de build de RF-231. Se agrega
la sección «identidad de build ≠ versión de release» + 3 filas nuevas de checklist
(`changelog-por-version` · `bump-solo-por-make` · `build-sellado`), y el nodo pasa a **v1.2**.

## ⏳ Pendiente de firma 🧑‍⚖️

VC-D1..VC-D7 quedan **DECIDIDAS y construidas**. Falta la firma del gate de PARIDAD.
