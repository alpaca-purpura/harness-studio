# Decisiones — volverlo de arnesia y publicar

> **FIRMADAS 🧑‍⚖️ 2026-07-30 (Chris)** — vía aprobación del plan de sesión del MVP de
> 1 día (el operador eligió además: prioridad n.º 1 = este paquete · validación directo
> en laptop · §10 estricto por carril).

- **B-D1 — rol/proceso se PIDEN, jamás se inventan.** El form de Identificar gana
  campos `rol` y `proceso` (texto libre + datalist sugerido desde `GET /api/harnesses`)
  y exige ≥1 `empresa` (el `anyOf` del schema). Sin catálogo de roles en el sistema
  (verificado), el operador es la autoridad. Vacíos ⇒ 400 honesto, botón disabled con
  motivo. NO se relaja el schema (doctrina firmada + dogfood 15/15 dependen de él).
- **B-D2 — el sello NO lleva `version`; SoT = `plugin.json.version`.** El schema ya
  declara `version` como derivado del loader; el `"0.1.0"` a mano de `selloDe` era la
  ambigüedad y se elimina. Identificar genera `.claude-plugin/plugin.json` mínimo
  (`{name, version:"0.1.0", description}`) SOLO si falta Y el target NO es
  `proyecto-instalado` — escribirlo ahí cambia la detección del loader
  (`loader.go:163-171`) y rompería el arnés. En `proyecto-instalado`: sello sin
  version + aviso honesto «sin plugin.json: no publicable en esta forma».
- **B-D3 — versionado de arnés = hoja NUEVA** `docs/architecture/conventions/
  versionado-arnes.md` (no se mezcla con `versionado.md` de la app, que está
  `enforced` con su propio checklist). Contenido: X.Y.Z autor en plugin.json
  (CC-native) · el «a.b.c.d» pedido = X.Y.Z + sello de extracción `AAMMDDHHMM` que
  estampa publish como campo aditivo `sello` en `catalogo.json` (coherente con la
  identidad de build de la app: versión deliberada ≠ sello automático; el sello JAMÁS
  entra al string semver — CC compara strings) · tag `<id>/vX.Y.Z` · status
  `proposed` con gaps declarados (enforcer llega con el publisher).
- **B-D4 — mockup = superset del Storybook vigente, sin html nuevo.** Disciplina
  `mockups/INDEX.md`: jamás proponer UI desde un html suelto; la superficie nueva son
  4 inputs en el form sin-sello del drawer (stories existentes) + el botón `▲
  Publicar` que ya existe disabled. Las stories nuevas SON el mockup.
- **B-D5 — publish rediseña el puerto** (`ports/publish.go` era stub sin
  consumidores): clones del marketplace-home en `~/.arnesia/publicaciones/<slug>/`
  (checkouts/ son canónicos de arneses, no se mezclan) · escritura read-modify-write
  con `map[string]any` sobre `marketplace.json`/`catalogo.json` ajenos (los structs de
  `parse.go` son lossy a propósito — solo releer/round-trip) · gate `conf.RunGraph`
  verde ANTES de push · `AccionPublicar` habilitada SOLO en
  `SituacionMiCopiaAdelantada` · pull-antes-de-push (`fetch` + `reset --hard
  origin/HEAD`) · push SIN force, rechazo ⇒ 409 «reintentá» · tag-TRAS-push, fallo de
  tag = aviso visible (no rollback fantasma) · auth `gh`/PAT calcada de
  `traer/externo.go` (duplicación documentada: go-arch-lint prohíbe `publish`→`traer`).
- **B-D6 — recortes explícitos del MVP** (gaps declarados, no silencio): bump
  interactivo FUERA (la versión ES la de plugin.json del canónico; editarla =
  mejorar-conversando) · changelog-de-arnés = Fase 2 · canales de catalogo.json no se
  mutan · CLI publish recableado = primero en cortar si aprieta.
