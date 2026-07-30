# Spec-lite — volverlo de arnesia y publicar

> Sancionada por la firma del corte (decisiones.md). Detalle de diseño verificado
> contra código por 2 agentes arquitectos (2026-07-30).

## B1 — Sello válido + versionado de arnés

**RF-B1.1** `Identificar` acepta `SolicitudIdentificar{InstallPath, ID, Nombre, Rol,
Proceso, Empresas, Marketplace}`; `selloDe` emite `domain.Arnes{ID, Nombre, Rol,
Proceso, Empresas, Marketplace, ReportaA:nil}` SIN `Version`. Guardas → 400
`ErrIdentificarSelloIncompleto` (rol/proceso/≥1 empresa).
**RF-B1.2** Validación pre-escritura: `PortafolioService` gana `ports.SchemaValidator`
(SchemaSet embebido); valida `{arnes: sello, nodos: []domain.Box{}}` (slice no-nil)
contra `graph.l0.schema.json`; inválido ⇒ `ErrIdentificarSelloInvalido`, NADA se
escribe.
**RF-B1.3** plugin.json mínimo si falta Y tipo ≠ `proyecto-instalado`; en
proyecto-instalado, aviso «sin plugin.json: no publicable en esta forma».
**RF-B1.4** HTTP body extendido + `openapi.yaml` sincronizado.
**RF-B1.5** Hoja `versionado-arnes.md` (B-D3).
**RF-B1.6** FE: form 4 campos (prellenos id=basename, empresa de la entrada; datalist
roles por prop `rolesConocidos`), botón disabled con motivo si incompleto; stories
form-válido / form-incompleto.

AC: (a) sello escrito PASA el schema real — test con SchemaSet real; (b) incompleto ⇒
nada escrito; (c) proyecto-instalado no gana plugin.json; (d) `Marketplace` poblado ⇒
re-key con `Identidad.Home`; (e) E2E laptop: dir crudo → Identificar → `cat
arnes.l0.json` válido → conformance del arnés verde.

## B2 — Publicar mínimo (write-side prenter-marketplace)

**RF-B2.1** `domain/publicar.go`: `SolicitudPublicacion`, `ResultadoPublicacion{Commit,
Tag, Avisos}`, centinelas (`SinAuth`, `PushRechazado`, `VersionYaPublicada`,
`VersionInvalida`, `SinCanonico`, `SinHome`, `NoPropio`, `ConformanceRojo`),
`RaizPublicaciones`, `SlugRepo`.
**RF-B2.2** Adapter `publish/publisher.go` real: clone/`fetch`+`reset --hard
origin/HEAD` → guarda idempotencia (`plugins/<id>/<v>/` poblado ⇒ YaPublicada, nada se
toca) → copiar árbol canónico (excl. `.git`, mismo set que `HashFormaPlugin`) →
RMW `marketplace.json` (source del `name==id` re-apuntado; filas ajenas intactas) +
`catalogo.json` (`versiones[] += {version, estado:"habilitada", fuente, fecha,
sello:AAMMDDHHMM}`) → commit → push (clasificación stderr → centinelas) → tag
`<id>/vX.Y.Z` tras push (fallo ⇒ aviso, no rollback).
**RF-B2.3** Usecase `MarketplaceService.Publicar(ctx, clave)` con `SetPublicar(pub,
cargar, conf)`: guardas canónico→home→propio→semver→gate `conf.RunGraph` (reporte
adjunto si rojo) → publicar → invalidar caché catálogo + `canonico.Version` upsert.
**RF-B2.4** `AccionPublicar` habilitada SOLO `SituacionMiCopiaAdelantada`.
**RF-B2.5** HTTP `POST /api/portafolio/arneses/{clave}/publicaciones` + openapi;
status: 400 (sin-canónico/home/no-propio/versión), 409 (ya-publicada · push-rechazado
· gate-rojo con `{error, conformance}`), 503 (sin-auth), 500 local.
**RF-B2.6** FE: botón `▲ Publicar` vivo en drawer (patrón `onTraerCanonico`:
spinner/error inline `role="alert"`; 409 con checks FAIL listados) + fix asserts de
motivo-disabled viejos (`marketplaces.ts:810,864`, `selectors.test.ts:255-257`) +
stories habilitado/curso/error.

AC: (a) test bare-repo git desde fixture prenter: árbol publicado + marketplace.json
round-trip con `parse.go` + fila catalogo + tag; (b) republicar misma versión ⇒ 409 y
working tree intacto; (c) push rechazado ⇒ 409 «reintentá»; (d) conformance rojo
bloquea con reporte; (e) E2E laptop contra repo GitHub real (+ `gh auth logout` ⇒ 503).

## Recortes si aprieta (en orden)
celda Publicar del catálogo → CLI publish → test push-rechazado integrado → datalist
roles → campo `sello` en catalogo.json.
**Innegociables:** validación sello pre-escritura · gate conformance · idempotencia ·
push sin force.

## Verificación
Por commit: `go test ./... -race` · fitness R1-R4 + openapi + changelog · lint ·
`npm run verify` · `conformance --todo` no cae · `estado.sh --check`. Por bloque:
skill `verificando-binario-instalado` → `make installer` → laptop.
