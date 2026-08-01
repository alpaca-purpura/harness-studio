---
regla: marketplace-referencia-es-solo-procedencia
version: 1.3
updated: 2026-08-01
status: enforced
ledger: HS-27
sources:
  - url: https://martinfowler.com/bliki/BoundedContext.html
    autoridad: experto
    revisado: 2026-07-25
  - url: https://learn.microsoft.com/en-us/azure/architecture/patterns/anti-corruption-layer
    autoridad: estándar
    revisado: 2026-07-25
enforced_by:
  # Los 5 enforcers son tests COLOCADOS junto al código (el patrón sancionado en
  # portafolio-identidad-y-deriva-honesta v1.1) más un source-scan en `fitness/` para el
  # invariante 4, que no es un comportamiento sino una forma del código.
  - internal/domain/marketplace_situacion_test.go:TestReferenciaNuncaHabilitaAccion
  - internal/domain/marketplace_test.go:TestClaseDesconocidaDegradaAReferencia
  - internal/adapters/marketplace/catalogo_local_test.go:TestLectorLocalNoEscribeEnElCheckout
  - docs/architecture/fitness/marketplace_shape_test.go:TestDominioNoAdoptaShapeAjeno
  - internal/domain/traer_test.go:TestPlanificarTraerRechazaReferencia
severity: critical
---

# Un marketplace de referencia es solo procedencia: read-only en el dominio, no solo en la UI

## L1 · Principio (estándar de industria)

**Un modelo que no controlás entra traducido y acotado, o te contamina el tuyo.** Dos formulaciones
maduras del mismo principio:

- **Bounded Context (DDD).** Un sistema grande no admite un modelo único: cada contexto tiene su
  propio modelo unificado, y la integración entre contextos pasa por un mapeo explícito, nunca por
  la adopción del vocabulario ajeno. *(experto: Fowler — «total unification of the domain model for
  a large system will not be feasible or cost-effective»)*
- **Anti-Corruption Layer (Evans).** Entre el sistema propio y uno externo va una capa de
  traducción; el sistema propio habla SIEMPRE su propio modelo, y la capa contiene toda la lógica
  de conversión. Sin ella, el modelo propio termina soportando esquemas, campos y semánticas del
  ajeno «para poder interoperar». *(estándar: Azure Architecture Center — «the core purpose of an
  anti-corruption layer is to protect the domain model»)*

**Corolario de producto (doctrina propia, `vision.md` §Qué mutó + `CLAUDE.md`):** *muere la idea de
operar/cargar arneses de terceros; solo arneses propios.* Entonces el catálogo de un marketplace
ajeno no es un agregado operable: es **dato de referencia** que existe para que un arnés pueda decir
de dónde viene. La distinción no puede vivir en un `disabled` de la UI — un `disabled` es una
decisión de render, y quien escriba el próximo widget no tiene por qué conocerla.

## L2 · Realización (este árbol Go+React)

Aterrizado en `internal/domain/marketplace{,_situacion}.go` +
`internal/adapters/marketplace/` + `internal/usecase/marketplace.go` (paquete
`docs/product/stories/2026-07-23-portafolio-agregar-marketplace/`, diseño en su `design.md`):

1. **La habilitación de acciones vive en el dominio, no en el widget.**
   `domain.AccionDeSituacion(situacion, clase)` devuelve `{Verbo, Habilitada, Motivo}` y para
   `ClaseReferencia` devuelve `Habilitada:false` en **toda** rama, con el motivo literal «no aplica:
   solo arneses propios». El widget renderiza lo que el dominio decidió; no tiene forma de habilitar
   un botón por su cuenta. ⇐ L1: el modelo propio decide, no la capa de presentación.
2. **La clase degrada al lado seguro.** `domain.ClaseSegura` mapea cualquier
   `ClaseMarketplace` fuera del enum a `referencia`, **nunca** a `propio`. Un registro editado a
   mano, un formato futuro o un typo no pueden habilitar operar lo ajeno. ⇐ L1: bounded context —
   lo que no reconozco entra en el modo más acotado, no en el más permisivo.
3-bis. **Sincronizar ≠ leer, y solo lo PROPIO se sincroniza (DD-2/E-bis, 2026-08-01).** «↻
   Refrescar» sobre un marketplace de clase `propio` hace `git pull --ff-only` del checkout ANTES
   de leer — vía el puerto `CatalogoSync` (adapter `SincronizadorGit`), un actor SEPARADO del
   lector: `LectorLocal` sigue sin escribir jamás (su enforcer queda intacto). Un marketplace de
   **referencia JAMÁS se sincroniza** — sobre lo ajeno solo se lee; la política (refresco
   explícito + clase propio) vive en el usecase y el pull fallido degrada VISIBLE en
   `Lectura.Motivo`. Enforcers: `TestRefrescarSincronizaSoloPropioYExplicito` ·
   `TestSincronizadorNoMergeaHistoriaDivergente`.
3. **Conocer ≠ adquirir.** `MarketplaceService.Registrar` escribe UNA fila en
   `~/.arnesia/marketplaces.json` y nada más: no clona, no instala, no descarga un árbol.
   `PortafolioService.AsignarOrigen` escribe el `home` declarado en el store y re-evalúa deriva: no
   trae ninguna copia. El lector de catálogo **nunca escribe** en `~/.claude/**` ni en el checkout
   del marketplace. ⇐ L1: la capa traduce, no toma posesión.
4. **El shape ajeno se traduce y se acota.** `plugins[].source` tiene cuatro formas reales
   (string de ruta relativa · objeto `git-subdir`/`url`/`github`) y las filas traen hasta 13 claves
   opcionales (`strict`, `skills`, `lspServers`, `homepage`, `author`, `displayName`, `keywords`,
   `category`, `tags`…). El adapter normaliza a `domain.SourceCatalogo` y **descarta el resto**:
   `internal/domain/marketplace.go` no menciona `git-subdir`, `$schema`, `installLocation` ni
   `renames` sin normalizar, y no crece un campo por cada extra del formato de Claude Code. Una
   forma no reconocida no se descarta en silencio: queda `SourceDesconocido` con su crudo visible.
   ⇐ L1: anti-corruption layer.

**Por qué es un nodo propio y no una extensión de
[`portafolio-identidad-y-deriva-honesta`](./portafolio-identidad-y-deriva-honesta.md):** ese nodo
enforça *no inventar cuando falta el dato* (honestidad del dato). Éste enforça *no adoptar ni operar
el modelo de un sistema que no controlo* (alcance del producto). Son dos L1 distintos; fusionarlos
diluye los dos. Los dos nodos se tocan en el mismo paquete y se leen juntos.

## Checklist evaluable

| id | qué chequea | severidad | señal en el mapa | enforcer |
|----|-------------|-----------|------------------|----------|
| referencia-no-opera | `AccionDeSituacion` devuelve `Habilitada:false` para clase `referencia` en las 6 situaciones, con el motivo literal | error | «catálogo de terceros con un botón que opera (traer/publicar/reparar)» | internal/domain/marketplace_situacion_test.go:TestReferenciaNuncaHabilitaAccion |
| clase-degrada-a-referencia | una `ClaseMarketplace` desconocida degrada a `referencia`, nunca a `propio` | error | «clase inválida habilitó operar un marketplace ajeno» | internal/domain/marketplace_test.go:TestClaseDesconocidaDegradaAReferencia |
| registrar-no-adquiere | `Registrar`/`AsignarOrigen` no clonan, no instalan y no escriben fuera de `~/.arnesia/`; el lector no escribe en `~/.claude/**` ni en el checkout ajeno | error | «registrar un marketplace descargó/escribió en un árbol ajeno» | internal/adapters/marketplace/catalogo_local_test.go:TestLectorLocalNoEscribeEnElCheckout |
| dominio-no-adopta-shape-ajeno | `internal/domain/marketplace*.go` no expone como campo NINGUNA clave cruda del formato ajeno (`$schema`, `lspServers`, `displayName`, `strict`, `keywords`, `category`, `tags`, `homepage`, `renames`, `metadata`, `installLocation`, `enabledPlugins`) — lo normalizado (`Tipo: "git-subdir"`, `install_location`) SÍ es del modelo propio | warn | «el dominio creció un campo por cada extra del catálogo ajeno» | docs/architecture/fitness/marketplace_shape_test.go:TestDominioNoAdoptaShapeAjeno |
| traer-referencia-rechazado | `PlanificarTraer` rechaza clase `referencia` en el DOMINIO, con cero I/O — el botón `disabled` de la UI no es control de acceso, y un POST puede llegar sin pasar por él | error | «se materializó un arnés de un marketplace de terceros» | internal/domain/traer_test.go:TestPlanificarTraerRechazaReferencia |

## Changelog

- 2026-08-01 · v1.3 · **DD-2/E-bis: nace `CatalogoSync`.** «Refrescar» releía el clone stale de CC
  y sincronizar era `git pull` manual (deuda E-bis del dogfood developer-vitalia). El pull vive en
  un actor separado del lector (invariante 3 intacto), solo clase `propio`, solo refresco
  explícito, `--ff-only` (divergencia = error visible, jamás merge silencioso). +2 enforcers.

- 2026-07-25 · v1.0 · Nodo fundacional (paquete
  `docs/product/stories/2026-07-23-portafolio-agregar-marketplace/`, etapa 3 · diseño técnico).
  Nace del reencuadre firmado en AG-D8: leyendo `vision.md` («muere operar arneses de terceros» ·
  el marketplace es el seam de entrega · ArnesIA es dueña única del observar y el modificar), el
  plano Marketplaces **no es una tienda** — es el estante de lo que vendemos y el espejo de si el
  cliente coincide. Sin la partición `propio`/`referencia` el plano degeneraba en gestor universal
  de plugins, justo lo que `vision.md` §Lo que NO es prohíbe. L1 = Bounded Context (Fowler) +
  Anti-Corruption Layer (Evans/Azure Architecture Center), las dos fuentes **verificadas en vivo el
  2026-07-25**. L2 = 4 invariantes sobre el dominio/usecase/adapter de marketplaces. **Nace
  `proposed`, no `enforced`**: el código todavía no existe (etapa 4 del paquete), y los 4 enforcers
  son tests colocados que shippean con él — el patrón sancionado en
  `portafolio-identidad-y-deriva-honesta` v1.1 (test colocado > test-proxy en `fitness/`). Los
  enforcers NO se listan en `enforced_by:` ni con su nombre `archivo_test.go:TestX` en la tabla a
  propósito: el parser del ruleset (`reGoTest`/`reTestName`) los levantaría e intentaría correrlos
  contra un paquete inexistente, convirtiendo un diferido honesto en un FAIL. 4 checks.
- 2026-07-25 · v1.1 · **+1 check `traer-referencia-rechazado`** (mismo día, `AG-D17` FIRMADA 🧑‍⚖️
  habilitó `↧ Traer canónico` completo — `design.md` §13). Con una acción que **materializa en disco**
  en juego, el invariante «`referencia` es read-only» deja de ser una cuestión de qué botón se pinta:
  `PlanificarTraer` tiene que rechazar la clase **en el dominio y con cero I/O**, porque un `POST
  /api/marketplaces/{nombre}/traidos` puede llegar sin haber pasado nunca por la UI. Un `disabled` es
  render, no control de acceso — y esa distinción es exactamente el L1 de este nodo (la decisión vive
  en el modelo propio, no en la capa de presentación). Sigue `proposed`; enforcer pendiente con la
  implementación. 4 → **5 checks**.
- 2026-07-25 · v1.2 · `proposed → enforced` (etapa 4 del paquete, MISMO día). El código existe y
  los **5 enforcers están cableados**: 4 tests COLOCADOS junto al código (el patrón sancionado en
  `portafolio-identidad-y-deriva-honesta` v1.1) + 1 source-scan en `fitness/` para el invariante 4,
  que no es un comportamiento sino una FORMA del código (qué campos declara `internal/domain/
  marketplace*.go`). El scan se acotó a las claves CRUDAS del formato ajeno: la redacción original
  del check listaba `git-subdir` e `installLocation`, pero los dos existen a propósito en el
  dominio ya NORMALIZADOS (`TipoSource("git-subdir")` es un valor del enum propio, y
  `install_location` es un campo del modelo propio con doc-comment explicando que es la ruta
  contra la que `deriva` compara) — prohibirlos habría sido prohibir la traducción, no la
  contaminación. Anotado como C20 en el `design.md` del paquete. 5 checks, los 5 corriendo.
