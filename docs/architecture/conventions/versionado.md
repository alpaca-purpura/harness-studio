---
regla: versionado
version: 1.2
updated: 2026-07-26
status: enforced
ledger: HS-24
sources:
  - url: https://semver.org/spec/v2.0.0.html
    autoridad: estándar
    revisado: 2026-07-15
  - url: https://keygen.sh/docs/api/releases/
    autoridad: oficial
    revisado: 2026-07-15
  - url: https://goreleaser.com/limitations/semver/
    autoridad: oficial
    revisado: 2026-07-15
  - url: https://v2.tauri.app/reference/config/
    autoridad: oficial
    revisado: 2026-07-15
enforced_by:
  - docs/architecture/fitness/arch_test.go:TestVersionManifestsInSync
  - docs/architecture/fitness/changelog_test.go:TestChangelogExisteYTieneForma
  - docs/architecture/fitness/changelog_test.go:TestChangelogCubreLaVersionDeLosManifiestos
  - scripts/bump.sh
  - scripts/changelog.py
  - lefthook.yml:pre-commit.changelog
  - /Makefile
severity: medium
---

# Versionado: SemVer 2.0 plano, una sola fuente de verdad, sin canal hasta que exista uno real

## L1 · Principio (estándar de industria)

**SemVer 2.0.0** (`MAJOR.MINOR.PATCH` + prerelease/build metadata opcionales) — el string de
versión es un contrato: PATCH = fix compatible, MINOR = feature compatible, MAJOR = breaking.
*(estándar: semver.org)*. Práctica complementaria confirmada en vivo para el fork
build-local-vs-release-oficial: no mutar archivos fuente en un build de prueba/snapshot —
bumpear versión es una acción deliberada y separada (`goreleaser --snapshot`, versionado de
Kubernetes vía `git describe`).

## L2 · Realización (este repo Go+Rust+TS)

- **Fuente de verdad:** `web/src-tauri/Cargo.toml` (`[package].version`). Tauri 2 cae a este
  valor si `tauri.conf.json` omite `version` — igual lo mantenemos explícito ahí (y en
  `web/package.json`) porque la doc oficial de Tauri lo recomienda así; el Makefile sincroniza
  los 3 para que ninguno quede a mano.
- **Formato en los 3 manifiestos:** semver plano `X.Y.Z`, **sin prefijo `v`** — es el formato
  exacto que exige `Release.version` de Keygen (server de licencias propuesto en
  `docs/product/stories/2026-07-15-instalador-publico-licencias-org/`, IL-D2/IL-D9), así no hay
  transformación que hacer cuando se cablee esa story.
- **Mecanismo:** `/Makefile` → `scripts/bump.sh` (punto único) — `make bump-patch|bump-minor|bump-major`
  sube el número, sincroniza los 3 archivos **y promueve el changelog** (§changelog-y-bump: sin
  entradas, el bump falla y no toca nada); `make installer` = bump-patch + `scripts/bundle.sh` + copia a `instaladores/vX.Y.Z/`
  (prefijo `v` SOLO en el nombre de carpeta de salida, mismo estilo que los tags de
  `.goreleaser.yaml`). El Makefile no commitea ni taggea — el bump queda en el working tree,
  revisable antes de commitear.
- **Canal (dev/beta/rc/alpha/stable):** no existe hoy — un solo track, sin sufijo de
  prerelease. Cuando exista un canal real, se agrega como `X.Y.Z-dev.N` y el campo `channel` de
  Keygen matchea el tag de prerelease literal (regla dura de Keygen: el channel DEBE matchear
  el prerelease tag) — el vocabulario de los 5 canales de Keygen (`stable/rc/beta/alpha/dev`) ya
  queda reservado, no hay que inventarlo después.
- **Dos tracks de versión independientes, no confundir:** (a) el bundle de escritorio Tauri
  (este nodo — `.deb`/`.rpm`/`.AppImage` vía `scripts/bundle.sh`) y (b) el binario standalone
  Go/CLI vía `.goreleaser.yaml` (tags git `vMAJOR.MINOR.PATCH`, pipeline nunca ejercitado — ver
  `research.md` §8 de la story de licencias). No comparten número de versión; son artefactos
  distintos con ciclos de release distintos.
- **`make installer` genera el artefacto — NO garantiza que lo estés VIENDO corriendo** (incidente
  real 2026-07-25, ledger pendiente): en una máquina que alguna vez migró a self-update
  (`self-update-sin-sudo`, CAP-60), existe `~/.local/bin/arnesia`. El shell Tauri instalado
  (`web/src-tauri/src/lib.rs`, bloque `override_local`) **prefiere SIEMPRE** ese binario de
  usuario sobre el sidecar recién empaquetado en el `.deb`/`.rpm`/`.AppImage` — es la mecánica
  que permite que el self-update se sobreviva a sí mismo tras la instalación inicial (con sudo,
  una sola vez). Efecto colateral: reinstalar un `.deb` nuevo en una máquina así **no cambia nada
  visible** hasta que `~/.local/bin/arnesia` también se actualice. Dos binarios, dos dueños:

  | Binario | Quién lo pone ahí | Quién lo actualiza | Requiere sudo |
  |---|---|---|---|
  | `/usr/bin/arnesia-app` (shell) + sidecar embebido | `make installer` → `.deb`/`.rpm`/`.AppImage` | reinstalar el paquete | sí |
  | `~/.local/bin/arnesia` (daemon real, si existe) | migración inicial única o el propio self-update | botón «Actualizar» (Ajustes) o `make dev-sync` | no |

  **Procedimiento correcto para verificar un build local recién generado:**
  1. `make installer` — arma el paquete versionado en `instaladores/vX.Y.Z/`.
  2. Si `~/.local/bin/arnesia` existe en la máquina de prueba, correr **también** `make dev-sync`
     (target nuevo: `scripts/bundle.sh --daemon-only` + write-tmp→rename atómico sobre ese path
     + mata el proceso viejo) — o el botón «Actualizar» de Ajustes, que hace lo mismo vía HTTP.
  3. Verificar que cambió de verdad, no asumir: `strings <binario> | grep -c '<algo del cambio>'`
     o comparar `mtime`/`git describe` embebido, antes de reportar "ya lo probé".

  **Gotcha adicional cazado armando `make dev-sync` (mismo incidente):** sobreescribir el
  destino con `install`/`cp` DIRECTO (sin `mv` atómico) sobre un binario que está corriendo en
  ese momento trunca las páginas mapeadas del proceso vivo y lo **crashea** (no es solo un
  ETXTBSY prolijo — el daemon murió a mitad de request). `updater.go` ya resuelve esto con
  write-tmp→rename (`Instalar()`); `make dev-sync` copia exactamente ese patrón, nunca
  `install`/`cp` directo sobre `$(DEV_DAEMON)`.

## Changelog y bump — el registro NO es opcional (2026-07-26, VC-D1..VC-D6)

> Orden del operador: *«que ya sea metodológico, no debo andar diciéndote nada»*. Por eso esto no
> es un consejo: es un gate. Paquete: `docs/product/stories/2026-07-26-versionado-y-changelog-metodologicos/`.

**Toda versión publicada declara qué trae**, en [`CHANGELOG.md`](../../../CHANGELOG.md) (raíz),
formato Keep a Changelog 1.1.0 con las 6 categorías canónicas en español: **Agregado · Cambiado ·
Deprecado · Eliminado · Corregido · Seguridad**. Ninguna otra categoría es válida.

- **Se escribe en el MISMO turno en que se construye**, igual que `decisiones.md` (metodología §10):
  `python3 scripts/changelog.py add Agregado "qué cambió"`.
- **El bump lo promueve, no lo pide:** `make bump-patch|bump-minor|bump-major` →
  `scripts/bump.sh` → (1) valida el changelog, (2) sincroniza los 3 manifiestos, (3) mueve
  `[Sin publicar]` a `## [X.Y.Z] — AAAA-MM-DD`. **Con `[Sin publicar]` vacía el bump falla ANTES
  de tocar ningún manifiesto** — no queda un working tree a medias.
- **Un solo punto de bump**, por el mismo motivo por el que hay un solo punto de sello
  (`bundle.sh`, RF-231/B-D2): si hubiera dos caminos, uno se olvidaría del changelog.
- **Qué bump elegir** (SemVer, sin adivinar):

  | Bump | Cuándo | Ejemplo de este repo |
  |---|---|---|
  | `bump-patch` | fix compatible, sin superficie nueva | el fix de `PATH` del self-update |
  | `bump-minor` | superficie nueva compatible | la identidad de build en Ajustes (RF-231) |
  | `bump-major` | rompe algo que el usuario ya usaba | mover `~/.local/bin/arnesia` de lugar |

- **La historia anterior a `0.2.22` NO se reconstruye.** Hay 20 releases previos sin changelog;
  inventarlos sería mentir. El archivo lleva `<!-- convencion-desde: 0.2.22 -->` y el enforcement
  no valida nada anterior (honestidad §4). Ese tramo vive en `LEDGER.md` + `git log`.

## Identidad de build ≠ versión de release (RF-231)

Un binario tiene **dos números y contestan preguntas distintas**. Confundirlos fue el bug original:
Ajustes mostraba solo el commit y no podía responder «¿corro lo último que compilé?».

| | Qué dice | Quién lo mueve | Dónde se ve |
|---|---|---|---|
| **Versión de release** `X.Y.Z` | qué contrato de funcionalidad es | `make bump-*` (deliberado, con changelog) | los 3 manifiestos · `instaladores/vX.Y.Z/` |
| **Sello de build** `AAMMDDHHMM` | qué COMPILACIÓN exacta es | `scripts/bundle.sh`, **solo** (automático, siempre) | `GET /api/version` · tarjeta de Ajustes |

El identificador completo es `X.Y.Z.AAMMDDHHMM` (ej. `0.2.21.2607260225`): dos compilaciones del
mismo commit comparten versión y commit, y **no** comparten sello. Un `go build` pelado (CI,
`go run`) no sella nada y reporta `dev` — decir «no sé qué build soy» es honesto; inventar un
número no. El sello se inyecta en `bundle.sh` y **nunca** en el Makefile: el self-update corre ese
mismo script, y desde el Makefile la app perdería su identidad al actualizarse a sí misma.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| changelog-por-version | toda versión ≥ `convencion-desde` en los manifiestos tiene su sección en `CHANGELOG.md`, con al menos una entrada | error | «se publicó una versión que no dice qué trae» | `docs/architecture/fitness/changelog_test.go:TestChangelogCubreLaVersionDeLosManifiestos` + `scripts/bump.sh` (gate) + `lefthook.yml` job `changelog` |
| changelog-forma | 6 categorías canónicas · semver plano con fecha · orden descendente · una sola `[Sin publicar]` · sin entradas de relleno (TBD/pendiente/…) | error | «changelog que existe pero no dice nada» | `docs/architecture/fitness/changelog_test.go:TestChangelogExisteYTieneForma` |
| bump-solo-por-make | la versión sube por `make bump-patch\|bump-minor\|bump-major` (→ `scripts/bump.sh`), nunca editando manifiestos a mano | error | «versión bumpeada a mano, changelog sin promover» | `changelog_test.go` (la versión a mano queda sin sección → rojo) + `arch_test.go:TestVersionManifestsInSync` (drift entre los 3) |
| build-sellado | todo binario distribuible sale de `scripts/bundle.sh` y lleva sello `AAMMDDHHMM`; un `go build` pelado reporta `dev` y no se distribuye | error | «un instalador que no sabe qué build es» | `scripts/bundle.sh` (único punto de `-ldflags`) + `internal/adapters/selfupdate/identidad_test.go` (12 tests) |
| sync-3-manifiestos | `Cargo.toml` == `tauri.conf.json` == `package.json` (mismo string de versión) | error | «versión en drift entre manifiestos» | `docs/architecture/fitness/arch_test.go:TestVersionManifestsInSync` |
| sin-prefijo-v-en-manifiesto | el campo `version` en los 3 manifiestos nunca lleva `v` adelante | error | «prefijo v en manifiesto rompe Keygen Release.version» | `docs/architecture/fitness/arch_test.go:TestVersionManifestsInSync` |
| sot-unica | `Cargo.toml` es el único archivo que se edita a mano; `tauri.conf.json`/`package.json` solo se tocan vía `make bump-patch` | warn | «version editada a mano fuera del Makefile» | revisión (gap: no hay enforcer que distinga edición manual de edición-por-Makefile) |
| carpeta-versionada-no-pisa | `make installer` nunca sobreescribe una carpeta `instaladores/vX.Y.Z/` ya existente | error | «instalador viejo pisado» | `/Makefile` (el bump lee el archivo en disco en cada corrida — siempre avanza) |
| dev-daemon-sincronizado | si `~/.local/bin/arnesia` existe en la máquina, `make installer` avisa explícito que hay que correr `make dev-sync` para verlo reflejado | warn | «instalé un .deb nuevo y la UI sigue vieja» | `/Makefile` (aviso post-build) + `make dev-sync` (mecanismo real) |

## Changelog

- 2026-07-15 · v1.0 · Nodo fundacional. Nace del Makefile de instaladores locales (`make
  installer`). L1 = SemVer 2.0.0. L2 = `Cargo.toml` SoT + sync 3 manifiestos + `/Makefile`,
  formato coherente con Keygen (story `instalador-publico-licencias-org`, IL-D9). Enforcer real
  desde el día 1: `TestVersionManifestsInSync` (cazó y corrigió un drift real preexistente —
  `package.json` en 0.1.0 vs `Cargo.toml`/`tauri.conf.json` en 0.2.0). Gap honesto: `sot-unica`
  no tiene enforcer automático (no se puede distinguir a nivel de git diff una edición manual de
  una hecha por `make bump-patch`) — `status: proposed`. 4 checks.
- 2026-07-15 · **FIRMADA 🧑‍⚖️ (HS-24)**. IL-D9 pasa de propuesta a decidida — `status: enforced`
  (`TestVersionManifestsInSync` verificado en vivo, PASS). Gap `sot-unica` sigue abierto
  (sin enforcer), no bloquea la firma: es warn, no error.
- 2026-07-25 · v1.1. Incidente real: se generó un instalador nuevo (v0.2.18), se instaló el
  `.deb`, la UI siguió mostrando contenido viejo. Causa raíz investigada a fondo (no se
  adivinó): el shell Tauri instalado prefiere `~/.local/bin/arnesia` (override local del
  self-update, `web/src-tauri/src/lib.rs`) por sobre el sidecar del paquete nuevo — mecánica
  deliberada del self-update-sin-sudo (CAP-60), pero con el efecto colateral de que `make
  installer` solo, sin sincronizar ese binario aparte, no se refleja en una máquina que ya
  migró. Fix: +1 check (`dev-daemon-sincronizado`) + target nuevo `make dev-sync` (mecanismo
  real, no solo prosa) + aviso automático al final de `make installer` cuando detecta el
  override. 4→5 checks.
- 2026-07-26 · **v1.2 — el changelog deja de ser opcional y la identidad de build entra a la
  convención.** Orden del operador: que actualizar la versión y registrar qué se agrega/corrige/
  elimina sea metodológico, «no debo andar diciéndote nada». Dos huecos medidos ese día: (a) el
  repo tenía **20 releases** en `instaladores/` (v0.2.2…v0.2.21) y **cero** changelog —
  `ls CHANGELOG*` vacío; (b) esta misma hoja, `enforced` desde v1.1, **no mencionaba el sello de
  build de RF-231** (`grep sello|ldflags|RF-231` sin resultados) — describía un mundo de
  solo-semver un día después de que la identidad del binario pasara a tener dos números. Fix as-code,
  no prosa: `CHANGELOG.md` (Keep a Changelog 1.1.0, 6 categorías cerradas) + `scripts/changelog.py`
  (check/add/release) + `scripts/bump.sh` (**punto único** de bump: valida ANTES de tocar
  manifiestos y promueve `[Sin publicar]`) + `make bump-minor`/`bump-major` con criterio escrito +
  `changelog_test.go` (3 tests, CI) + job `changelog` de lefthook. Historia previa **no
  reconstruida** a propósito (`convencion-desde: 0.2.22`). 5→9 checks. Paquete:
  `docs/product/stories/2026-07-26-versionado-y-changelog-metodologicos/`.
