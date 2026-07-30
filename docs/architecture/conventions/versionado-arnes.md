---
regla: versionado-arnes
version: 1.0
updated: 2026-07-30
status: proposed
sources:
  - url: https://semver.org/spec/v2.0.0.html
    autoridad: estándar
    revisado: 2026-07-30
enforced_by:
  - internal/usecase/portafolio_test.go:TestIdentificarSellaYRekey
  - internal/usecase/portafolio_test.go:TestIdentificarConMarketplacePueblaHome
severity: medium
---

# Versionado de ARNÉS: X.Y.Z del autor en plugin.json, sello de extracción aparte, tag por arnés

> **No confundir con [`versionado.md`](./versionado.md)**, que versiona la APP (Cargo.toml
> SoT, `make bump-*`, `enforced` con su propio checklist — no se mezcla, B-D3). Esta hoja
> versiona el PRODUCTO que la fábrica vende: cada arnés del Portafolio. Nace con B1 del
> paquete [`2026-07-30-volverlo-de-arnesia-y-publicar`](../../product/stories/2026-07-30-volverlo-de-arnesia-y-publicar/INDEX.md)
> (B-D2/B-D3 firmadas 🧑‍⚖️ 2026-07-30).

## L1 · Principio

**SemVer 2.0.0 plano para la versión de autor** (`X.Y.Z`, sin prefijo `v` en el manifiesto),
con una restricción dura CC-native: **Claude Code compara versiones de plugin como
STRINGS** — cualquier metadato dentro del string (build, fecha, sello) rompe el
update-check. Coherente con la identidad de build de la app
([`versionado.md`](./versionado.md) §identidad-de-build): **versión deliberada ≠ sello
automático** — dos números, dos preguntas, jamás fundidos en uno.

## L2 · Realización

- **SoT = `.claude-plugin/plugin.json` → `version`** — la X.Y.Z que el AUTOR declara
  (CC-native: es el campo que CC ya lee). El sello `arnes.l0.json` **NO lleva `version`**
  (B-D2): el schema `graph.l0` la documenta como derivada del loader (D-DOM-1), y el
  `"0.1.0"` que `selloDe` tecleaba a mano era la ambigüedad — dos números para la misma
  pregunta. `Identificar` genera el plugin.json mínimo (`{name, version: "0.1.0",
  description}`) SOLO si falta y el target NO es `proyecto-instalado` (ahí escribirlo
  cambiaría la detección del loader — plugin manda sobre `.claude/` — y rompería el arnés:
  queda el aviso honesto «sin plugin.json: no publicable en esta forma»).
- **El «a.b.c.d» pedido = `X.Y.Z` + sello de extracción `AAMMDDHHMM`.** El sello lo
  estampará **publish** (B2) como campo ADITIVO `sello` en la fila de
  `catalogo.json.versiones[]` — **jamás dentro del string semver** (CC compara strings).
  Misma mecánica que el sello de build de la app: automático, del acto de publicar, nunca
  tecleado.
- **Tag por publicación: `<id>/vX.Y.Z`** en el repo del marketplace (namespacing por arnés;
  se taggea TRAS el push — un fallo de tag es aviso visible, no rollback fantasma, B-D5).
- **Bump del arnés = editar el plugin.json del canónico** (vía mejorar-conversando). Bump
  interactivo desde la UI queda FUERA del MVP (B-D6, recorte explícito).

## Gaps declarados (por qué `proposed`)

- **El enforcer del lado publish llega con el publisher (B2):** sello aditivo en
  `catalogo.json`, tag `<id>/vX.Y.Z`, no-reuso de versión publicada. Hasta que B2 landee,
  esos checks difieren honestos — no hay pass fabricado.
- **Changelog-de-arnés = Fase 2 (B-D6):** un arnés publicado aún no declara qué trae cada
  versión. Gap visible, no silencio.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| sello-sin-version | el `arnes.l0.json` que escribe Identificar NO lleva `version`; la SoT es `plugin.json.version` | error | «dos números para la misma pregunta» | internal/usecase/portafolio_test.go:TestIdentificarSellaYRekey |
| plugin-json-si-falta | Identificar genera el plugin.json mínimo cuando falta y el target no es `proyecto-instalado`; en proyecto-instalado NO lo genera (rompería el detector del loader) y deja aviso honesto | error | «arnés sellado sin SoT de versión, o loader re-detectado» | internal/usecase/portafolio_test.go:TestIdentificarConMarketplacePueblaHome (genera) + TestIdentificarNoGeneraPluginJSONEnProyectoInstalado (no genera + aviso) |
| sello-extraccion-aditivo | publish estampa `sello: AAMMDDHHMM` como campo aditivo en `catalogo.json`, nunca dentro del string semver | error | «metadato de build dentro del semver — CC compara strings» | gap: llega con el publisher (B2) |
| tag-por-arnes | cada publicación taggea `<id>/vX.Y.Z` tras el push; fallo de tag = aviso visible | warn | «versión publicada sin tag rastreable» | gap: llega con el publisher (B2) |
| changelog-de-arnes | toda versión publicada de un arnés declara qué trae | warn | «versión de arnés muda» | gap: Fase 2 (B-D6) |

## Changelog

- 2026-07-30 · v1.0 · Nace con B1 del paquete `volverlo-de-arnesia-y-publicar` (B-D2/B-D3):
  el sello deja de llevar `version` (la SoT pasa a `plugin.json.version`, que Identificar
  genera si falta fuera de proyecto-instalado), el «a.b.c.d» pedido se resuelve como
  X.Y.Z + campo aditivo `sello` (lo estampará publish, B2), y el tag por arnés queda
  `<id>/vX.Y.Z`. 2 checks con test colocado desde el día 1; 3 gaps declarados hasta el
  publisher. `status: proposed` — se promueve cuando B2 traiga el enforcer del lado
  publish.
