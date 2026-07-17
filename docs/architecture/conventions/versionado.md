---
regla: versionado
version: 1.0
updated: 2026-07-15
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
- **Mecanismo:** `/Makefile` — `make bump-patch` bumpea el patch (Z) y sincroniza los 3
  archivos; `make installer` = bump + `scripts/bundle.sh` + copia a `instaladores/vX.Y.Z/`
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

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| sync-3-manifiestos | `Cargo.toml` == `tauri.conf.json` == `package.json` (mismo string de versión) | error | «versión en drift entre manifiestos» | `docs/architecture/fitness/arch_test.go:TestVersionManifestsInSync` |
| sin-prefijo-v-en-manifiesto | el campo `version` en los 3 manifiestos nunca lleva `v` adelante | error | «prefijo v en manifiesto rompe Keygen Release.version» | `docs/architecture/fitness/arch_test.go:TestVersionManifestsInSync` |
| sot-unica | `Cargo.toml` es el único archivo que se edita a mano; `tauri.conf.json`/`package.json` solo se tocan vía `make bump-patch` | warn | «version editada a mano fuera del Makefile» | revisión (gap: no hay enforcer que distinga edición manual de edición-por-Makefile) |
| carpeta-versionada-no-pisa | `make installer` nunca sobreescribe una carpeta `instaladores/vX.Y.Z/` ya existente | error | «instalador viejo pisado» | `/Makefile` (el bump lee el archivo en disco en cada corrida — siempre avanza) |

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
