# Spec — Versionado y changelog metodológicos (RF-232)

> Paquete `2026-07-26-versionado-y-changelog-metodologicos`. **El QUÉ**, no el cómo.
> Decisiones: [`decisiones.md`](./decisiones.md) (VC-D1..VC-D7). Numeración: continúa desde RF-231.
> Es un RF **de proceso**: la superficie es la línea de comandos y el CI, no la app.

### RF-232 — La versión no sube sin decir qué trae, y nadie tiene que acordarse

## § Mapa funcional

### 1. Happy path

1. Se construye algo (código, fix, superficie). En **el mismo turno**, quien lo construyó anota:
   `python3 scripts/changelog.py add Agregado "…"` → la entrada aterriza en `[Sin publicar]`.
2. Llega el momento de publicar: `make bump-patch` (o `minor`/`major` según el criterio escrito).
3. `scripts/bump.sh` valida el changelog **antes de tocar nada**, sincroniza los 3 manifiestos y
   promueve `[Sin publicar]` → `## [X.Y.Z] — AAAA-MM-DD`.
4. `make installer` empaqueta esa versión en `instaladores/vX.Y.Z/`.
5. CI verifica que la versión de los manifiestos tenga su sección con contenido.

### 2. Bifurcaciones

```
`make bump-*`
├─ ¿[Sin publicar] tiene al menos una entrada?
│   ├─ NO → Bif-1 · ABORTA con el comando exacto para arreglarlo; CERO archivos tocados → [SC-2]
│   └─ SÍ ↓
├─ ¿el changelog tiene forma válida? (6 categorías · semver+fecha · orden · sin relleno)
│   ├─ NO → Bif-2 · aborta y lista TODAS las fallas juntas → [SC-3]
│   └─ SÍ ↓
├─ ¿la versión destino ya existe en el changelog?
│   ├─ SÍ → Bif-3 · aborta («¿bump repetido?») → [SC-4]
│   └─ NO ↓
└─ Bif-4 · bumpea los 3 manifiestos + promueve la sección → [SC-1]

alguien bumpea A MANO (editor, sed, merge), sin pasar por make
└─ Bif-5 · CI rojo: la versión de los manifiestos no tiene sección → [SC-5]

`changelog.py add` con categoría inventada ("Varios")
└─ Bif-6 · rechaza y lista las 6 válidas → [SC-6]
```

### 3. Reglas de negocio

| ID | Regla | Ancla |
|---|---|---|
| **RN-1** | Toda versión publicada ≥ `convencion-desde` tiene sección con ≥1 entrada | VC-D2 |
| **RN-2** | El bump valida ANTES de escribir: si falla, **cero** archivos modificados | VC-D2 |
| **RN-3** | Un solo punto de bump (`scripts/bump.sh`); el Makefile no duplica la lógica | VC-D2 · B-D2 |
| **RN-4** | Las categorías son 6 y el set es cerrado | VC-D1 |
| **RN-5** | La entrada se escribe en el turno de construcción, no en el de release | VC-D6 |
| **RN-6** | La historia previa a `0.2.22` no se inventa: se declara no reconstruida | VC-D4 · honestidad §4 |
| **RN-7** | El bump no toca git (ni commit ni tag): queda revisable en el working tree | Makefile vigente |
| **RN-8** | Sello de build ≠ versión de release: números distintos, mecanismos distintos | VC-D7 · RF-231 |

### 4. Criterios de aceptación

- [x] **AC-1** — Existe `CHANGELOG.md` con formato Keep a Changelog y las 6 categorías.
- [x] **AC-2** — `make bump-patch` con entradas: bumpea los 3 manifiestos y promueve la sección.
- [x] **AC-3** — `make bump-patch` con `[Sin publicar]` vacía: **falla y no toca ningún archivo**.
- [x] **AC-4** — Existen `bump-minor` y `bump-major` con criterio escrito de cuándo usar cada uno.
- [x] **AC-5** — `changelog.py add` acepta alias, es idempotente y rechaza categorías inventadas.
- [x] **AC-6** — CI falla si los manifiestos declaran una versión sin sección en el changelog.
- [x] **AC-7** — La convención `versionado.md` documenta changelog **y** sello de build.
- [x] **AC-8** — La regla vive en `CLAUDE.md` + metodología §10: una sesión nueva la lee sin que el operador la dicte.
- [ ] **AC-9** — 🧑‍⚖️ Gate humano: el operador corre un ciclo real y firma.

## § Gherkin

```gherkin
Escenario: SC-1 · publicar una versión con cambios registrados
  Dado que [Sin publicar] tiene al menos una entrada
  Cuando se corre make bump-patch
  Entonces los 3 manifiestos quedan en la versión nueva
  Y el changelog gana una sección fechada con esas entradas
  Y [Sin publicar] queda vacía con las 6 categorías listas

Escenario: SC-2 · intentar publicar una versión muda
  Dado que [Sin publicar] no tiene ninguna entrada
  Cuando se corre make bump-patch
  Entonces el bump falla
  Y dice el comando exacto para agregar la entrada
  Y ningún manifiesto queda modificado

Escenario: SC-3 · changelog con forma rota
  Dado un changelog con una categoría inventada o una versión sin fecha
  Cuando se valida
  Entonces se listan TODAS las fallas juntas, no la primera

Escenario: SC-4 · bump repetido a una versión que ya existe
  Cuando se intenta promover una versión ya presente en el changelog
  Entonces se rechaza

Escenario: SC-5 · alguien bumpeó a mano
  Dado un manifiesto editado a mano a una versión sin sección
  Cuando corre el fitness en CI
  Entonces falla diciendo que se bumpee con make

Escenario: SC-6 · categoría fuera de las 6
  Cuando se intenta agregar una entrada en "Varios"
  Entonces se rechaza y se listan las 6 válidas

Escenario: SC-7 · historia previa
  Dado que los manifiestos declaran una versión anterior a convencion-desde
  Cuando corre el fitness
  Entonces skipea con el motivo visible, sin exigir historia inventada
```

## § Matriz de cobertura

| ID | Escenario | Verificación real | Estado |
|---|---|---|---|
| Bif-1 · RN-2 | SC-2 | `make bump-patch` con `[Sin publicar]` vacía → `Error 1`, `Cargo.toml` intacto en 0.2.22 | ✅ corrido |
| Bif-2 · RN-4 | SC-3 · SC-6 | `changelog.py add Varios` → exit 1 · `TestChangelogExisteYTieneForma` | ✅ |
| Bif-3 | SC-4 | `changelog.py release` sobre versión existente (rama `ya existe en el changelog`) | ✅ código + lectura |
| Bif-4 · RN-1 | SC-1 | `make bump-patch` 0.2.21→0.2.22 y `make bump-minor` 0.2.22→0.3.0, changelog promovido las dos veces | ✅ corrido |
| Bif-5 | SC-5 | `TestChangelogCubreLaVersionDeLosManifiestos` (pasó con 0.2.22 presente) + job `changelog` de lefthook | ✅ |
| Bif-6 · RN-4 | SC-6 | salida real: `categoría 'Varios' inválida. Usá una de: …` | ✅ corrido |
| RN-3 | — | `Makefile` delega los 3 targets en `scripts/bump.sh`; no hay otro `sed` de versión | ✅ lectura |
| RN-5 | — | `CLAUDE.md` regla dura + `metodologia.md` §10 regla 1 + cabecera de `CHANGELOG.md` | ✅ |
| RN-6 | SC-7 | `TestChangelogCubreLaVersionDeLosManifiestos` → `SKIP: versión 0.2.21 es anterior a la convención (0.2.22)` | ✅ corrido |
| RN-7 | — | `bump.sh` no invoca git; el Makefile lo declara | ✅ |
| RN-8 | — | `versionado.md` §identidad-de-build-≠-versión-de-release (tabla de dos números) | ✅ |
| AC-9 | — | ciclo real corrido por el operador | ⛔ **abierto** |

**Huecos:** ninguno sin verificación salvo AC-9.

## § Límites aceptados

| Límite | Por qué |
|---|---|
| El changelog se escribe **a mano** (no se deriva de commits) | Un changelog generado de mensajes de commit dice «fix(portafolio): ajuste» — ruido de implementación, no lo que el usuario nota. La categoría y la redacción son un acto editorial |
| No hay tag git automático | El bump deja el cambio en el working tree para revisar (RN-7). Automatizarlo es otra decisión, no esta |
| `sot-unica` sigue sin enforcer perfecto | Un `git diff` no distingue edición manual de `make` — pero ahora el bump a mano queda **rojo en CI** por falta de sección: el gap se achicó, no se cerró |
| Las entradas iniciales de `[Sin publicar]` se derivaron de las capabilities nuevas y del paquete de RF-231 | Es lo verificable hoy; lo construido por otras sesiones que no dejó capability aún puede faltar — se agrega al cerrar cada paquete |

## Trazabilidad

- **Convención:** [`versionado.md`](../../../architecture/conventions/versionado.md) v1.2 (4 checks nuevos).
- **Capabilities:** ninguna nueva — `scripts/` y `Makefile` están fuera del alcance de R2
  (`cmd/` · `internal/` · `web/src` · `web/src-tauri/src`), verificado en `capability_trace_test.go`.
- **PARIDAD:** [`PARIDAD.md`](./PARIDAD.md).
