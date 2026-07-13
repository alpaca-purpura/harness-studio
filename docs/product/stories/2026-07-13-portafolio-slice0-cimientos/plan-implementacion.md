# Plan de implementación — Portafolio · Slice 0 «Cimientos» (dominio/backend, SIN FE)

> `tipo: plan` · paquete `2026-07-13-portafolio-slice0-cimientos` · autor: Fable 5 (arquitecto/líder técnico) 2026-07-13.
> Ejecuta: **Sonnet 5** (constructor). El modelo funcional está FIRMADO (HS-22) y las decisiones técnicas
> cerradas en [`decisiones.md`](./decisiones.md) (S0-D1..D11) — acá NO se decide, se CONSTRUYE.
> Specs fuente: `../2026-07-10-spike-carga-arneses/spec-funcional.md` (§3-§9, §11) · `casuistica.md`
> (C-P/C-OR/C-ID/C-N) · BACKLOG «Programa Portafolio» item 0.

## 0 · Goal (definición de terminado — verificable, no negociable)

Slice 0 está TERMINADO cuando **todo** esto es cierto (nada fabricado; gap = visible + documentado):

1. Los 7 deltas de dominio D-DOM-1..7 implementados según S0-D1..D11.
2. `go build ./... && go test ./... -race` verde · `golangci-lint run` limpio · go-arch-lint verde.
3. `go test ./docs/architecture/fitness/...` verde (R1-R4 capabilities + boundary nuevo).
4. `bash scripts/estado.sh --check` verde (cifras regeneradas); `go run ./cmd/arnesia conformance --todo`
   sin regresión (pass ≥ 42 · fail 0); `conformance --arnes dogfood/dev-full-cycle` sin fail NUEVO
   (el warn honesto `art-es-path` preexistente se queda).
5. `cd web && pnpm run verify` verde (la migración `empresas[]` no rompe types/stories).
6. **E2E vivo contra la máquina real** (§T9): `arnesia portafolio escanear` sobre un proyecto real
   detecta las 3 formas de instalación, resuelve origen vía eslabón CC con versión REAL de
   `installed_plugins.json`, y la deriva del plugin del kit se evalúa contra el checkout local del
   marketplace (o reporta `deriva-no-evaluable` con motivo). Salidas HONESTAS (`?`, `desconocido`,
   `deriva-no-evaluable`) donde no hay dato — jamás inventadas.
7. 5 capabilities nuevas + extends registradas; boundary nuevo enforced; `paridad.md` escrita con la
   tabla spec↔realidad (firma humana queda PENDIENTE para el operador — no la simules).
8. Un commit Conventional por ticket en `main` (`feat(portafolio): T<N> — <qué>`), cada uno con los gates
   de §M verdes ANTES de commitear.

## 1 · Contexto mínimo que debés leer antes de tocar código

1. `docs/product/stories/2026-07-10-spike-carga-arneses/spec-funcional.md` — el modelo (30 min).
2. `docs/product/stories/2026-07-10-spike-carga-arneses/casuistica.md` — §A, §B, §C, §G, §I.
3. [`decisiones.md`](./decisiones.md) de este paquete — S0-D1..D11 (cierra todo lo que la spec dejaba abierto).
4. Código ancla: `internal/domain/graph.go` · `internal/adapters/loader/loader.go` ·
   `internal/adapters/store/arnes_registry.go` (patrón atomic-save + checkProtected) ·
   `internal/adapters/index/store.go` · `cmd/arnesia/main.go` (composition root) ·
   `internal/adapters/transport/http/{router,arneses}.go`.
5. `docs/product/_templates/capability.template.yaml` + `docs/architecture/boundaries/codigo-traza-a-capability.md` (R1-R4).

## 2 · Diseño técnico (contratos exactos)

### 2.1 Dominio nuevo — `internal/domain/portafolio.go`

```go
// IdentidadArnes — clave del Portafolio (BR-1, S-D10/F-A). Home = marketplace autor-declarado
// CANONICALIZADO (host/owner/repo, minúsculas); Home=="" ⇒ identidad provisional y Scope discrimina
// (RN-IDENT-2: install-path relativo al proyecto en local, o remote canonicalizado del proyecto en git).
type IdentidadArnes struct {
    Home  string `json:"home,omitempty"`
    ID    string `json:"id"`
    Scope string `json:"scope,omitempty"`
}
func (i IdentidadArnes) Provisional() bool
// Clave: slug estable p/ URL y dedup: "<home-slug>~<id>~<scope-slug>" (slug = [a-z0-9-_], resto '-').
func (i IdentidadArnes) Clave() string

type EstadoDeriva string
const (
    DerivaAlHilo      EstadoDeriva = "al-hilo"
    DerivaEnDeriva    EstadoDeriva = "en-deriva"
    DerivaNoEvaluable EstadoDeriva = "deriva-no-evaluable"
)

type TipoInstalacion string
const (
    InstMaterializada     TipoInstalacion = "materializada"      // raíz/.claude/plugins/<id>/ (HS-12)
    InstReferenciadaCC    TipoInstalacion = "referenciada-cc"    // enabledPlugins → cache CC (S0-D1)
    InstProyectoInstalado TipoInstalacion = "proyecto-instalado" // el .claude/ del proyecto mismo
)

// EslabonOrigen — un dato crudo de la cadena §6 con su fuente anotada (BR-3, trazabilidad).
type EslabonOrigen struct {
    Fuente string `json:"fuente"` // manifiesto | lock-devstudio | cc-plugins | git-plugin | git-proyecto | no-legible
    Campo  string `json:"campo"`  // home | registry | version | empresa | canal | proyecto-remote
    Valor  string `json:"valor"`
}

// Origen — reconciliación collect-all (spec §6). Registry/Version vacíos = desconocido/? honestos.
type Origen struct {
    Registry      string          `json:"registry,omitempty"`
    Version       string          `json:"version,omitempty"`
    Eslabones     []EslabonOrigen `json:"eslabones,omitempty"`
    Discrepancias []string        `json:"discrepancias,omitempty"` // C-OR-6: visibles, jamás elección silenciosa
}

type Instalacion struct {
    ProyectoPath string          `json:"proyecto_path"`
    InstallPath  string          `json:"install_path"` // canonicalizado (EvalSymlinks+Clean, C-N-3)
    Tipo         TipoInstalacion `json:"tipo"`
    Origen       Origen          `json:"origen"`
    Deriva       EstadoDeriva    `json:"deriva"`
    DerivaDetalle string         `json:"deriva_detalle,omitempty"` // p.ej. "sin referencia local de home@v"
    Aviso        string          `json:"aviso,omitempty"`          // C-P-5 no-reconocible · C-P-14 declarada-ausente · C-P-9 no-encontrada
}

type Canonico struct { // Slice 0: forma + clasificación RN-IDENT-4; el clone llega en Slice 2
    Path    string `json:"path"`
    Version string `json:"version,omitempty"`
}

// EntradaPortafolio — 1 card por identidad (BR-1). Registries = facetas de adquisición (N:M).
type EntradaPortafolio struct {
    Identidad     IdentidadArnes `json:"identidad"`
    Nombre        string         `json:"nombre,omitempty"`
    Descripcion   string         `json:"descripcion,omitempty"`
    Empresas      []string       `json:"empresas,omitempty"`
    Registries    []string       `json:"registries,omitempty"`
    Canonico      *Canonico      `json:"canonico,omitempty"`
    Instalaciones []Instalacion  `json:"instalaciones,omitempty"`
    Agregado      string         `json:"agregado,omitempty"` // RFC3339
}
```

Funciones PURAS (mismo archivo o `origen.go` del dominio):

```go
// ResolverOrigen aplica el orden de autoridad S0-D1 sobre eslabones crudos:
// procedencia-de-copia: lock-devstudio ≙ cc-plugins (install-time) > git-plugin > (manifiesto NO da registry).
// version: lock/cc > plugin.json. Conflictos → Discrepancias[] (no elige en silencio).
func ResolverOrigen(eslabones []EslabonOrigen) Origen

// ResolverIdentidad deriva la clave: home = CanonicalizarRepo(manifiesto.Marketplace) si existe;
// id según RN-IDENT-3 (arnes.l0.id → plugin.json.name → dirname; discrepancia → aviso, no silencio);
// sin home → provisional con scope (RN-IDENT-2).
func ResolverIdentidad(a *Arnes, idFallback, scopeLocal, scopeRemote string) (IdentidadArnes, aviso string)
```

### 2.2 Canonicalización — `internal/domain/repo_ref.go` (D-DOM-7, RN-IDENT-1)

```go
// CanonicalizarRepo normaliza cualquier forma de referencia git a "host/owner/repo" (minúsculas,
// sin esquema, sin .git, sin trailing). Formas aceptadas: "owner/repo" (asume github.com),
// "https://host/owner/repo(.git)(/)", "git@host:owner/repo(.git)", "ssh://git@host/owner/repo".
// ok=false si no parsea a host/owner/repo — el caller conserva el crudo como dato visible.
func CanonicalizarRepo(ref string) (canon string, ok bool)
```

### 2.3 Cambios a dominio existente (`graph.go`) — D-DOM-1/D-DOM-2 (S0-D3)

- `Arnes.Version string \`json:"version,omitempty"\`` (nuevo; fuente: `plugin.json.version` vía loader).
- `Arnes.Empresa string` → `Empresas []string \`json:"empresas,omitempty"\`` + `UnmarshalJSON` custom que
  acepta legacy `"empresa": "x"` → `["x"]` (Marshal emite solo `empresas`).
- `Arnes.Marketplace` queda tal cual (key `marketplace` = home crudo autor-declarado).
- NUEVO `Arnes.FuenteManifiesto string \`json:"fuente_manifiesto,omitempty"\`` — `"arnes.l0.json"` |
  `"plugin.json"` (BR-3: anotar fuente; deja detectar «degradado» sin el hack `Arnes==nil`).

### 2.4 Loader (`internal/adapters/loader/loader.go`) — D-DOM-4

`leerManifiesto(dir)`:
1. `arnes.l0.json` existe → como hoy + estampa `FuenteManifiesto:"arnes.l0.json"`; ADEMÁS si existe
   `.claude-plugin/plugin.json`, parsea y completa `Version` (y `Nombre`/`Descripcion` si vacíos —
   cadena bendecida nomenclatura §2); si `arnes.l0.id ≠ plugin.json.name` → el graph gana un nodo-aviso
   NO (eso es para elementos) — se anota discrepancia devolviéndola al caller: firma nueva
   `leerManifiesto(dir) (*domain.Arnes, aviso string, err error)` (LoadArnes la propaga en un nuevo
   retorno o la loguea — mantené `LoadArnes(dir) (domain.Graph, error)` y agregá
   `LoadArnesInfo(dir) (domain.Graph, Info, error)` si necesitás el aviso; `LoadArnes` delega).
2. Sin `arnes.l0.json` pero SÍ `.claude-plugin/plugin.json` (**C-N-14, el caso más común**) → fallback:
   `Arnes{ID: name, Nombre: displayName||name, Descripcion: description, Version: version,
   FuenteManifiesto:"plugin.json"}`. plugin.json corrupto → (nil, aviso, nil) degradado visible.
3. Ninguno → (nil, "", nil) como hoy (forma-instalada sin manifiesto sigue legal).

### 2.5 Adapter nuevo — `internal/adapters/portafolio/` (S0-D4)

**`scanner.go`** — walker físico (D-DOM-3 + D-DOM-6 + S0-D1/D2):

```go
type Scanner struct {
    CCPluginsDir string // default ~/.claude/plugins — INYECTABLE para tests
    MaxDepth     int    // default 4 — C-P-11 monorepo acotado
}
type Hallazgo struct {
    Dir       string                 // dir a cargar con el loader (forma-plugin o raíz de proyecto)
    Tipo      domain.TipoInstalacion
    Eslabones []domain.EslabonOrigen // crudos: lock, cc-plugins, git-plugin, git-proyecto
    Aviso     string                 // C-P-5 / C-P-14
}
func (s *Scanner) Escanear(ctx context.Context, root string) ([]Hallazgo, error)
```

Reglas del walk (TODAS con test):
- `root/.claude/` poblado → 1 hallazgo `proyecto-instalado` (dir=root).
- `root/.claude/plugins/<id>/` con `.claude-plugin/plugin.json` → `materializada`; sin plugin.json →
  hallazgo con `Aviso` no-reconocible (C-P-5), no crash, resto sigue.
- `root/.devstudio/arneses.yaml` (lock, formato HS-12: entradas `id·versión·canal·registry`) → eslabones
  `lock-devstudio` por entrada; entrada cuyo dir no existe → hallazgo `Aviso:"declarada-en-lock, ausente"`
  (C-P-14 análogo; jamás omitir).
- `root/.claude/settings.json#enabledPlugins` con `"<id>@<mkt>": true` → cruzar con
  `<CCPluginsDir>/installed_plugins.json` (guard `"version":2`; filtrar `projectPath == root`) +
  `known_marketplaces.json` → hallazgo `referenciada-cc` con Dir = installPath del cache, eslabones
  `cc-plugins` (registry = repo del marketplace, version). Metadata ilegible → eslabón `no-legible` visible.
- Subcarpetas: descender buscando `.claude/` anidados hasta MaxDepth; NO seguir symlinks fuera de root
  (C-P-12); saltar `node_modules`, `.git`, dirs ocultos salvo `.claude`/`.devstudio`; `ctx` cancelable (C-P-13).
- Detección git del proyecto: `git-proyecto` eslabón con remote `origin` (fallback: nada); resolver
  `.git` archivo `gitdir:` (worktree, C-N-1); `origin` gana, `upstream` = eslabón secundario (C-N-2);
  bare repo → no escaneable (error honesto). Parser manual de `.git/config` (S0-D8), en `gitconfig.go`.
- `root/.claude/plugins/<id>/.git` existe → eslabón `git-plugin` con su remote (C-OR-3, C-N-16).

**`store.go`** — persistencia (S0-D5, BR-11):

```go
type EntradaCorrupta struct { Raw json.RawMessage; Motivo string }
func NewStore(path string) (*Store, error) // path default ~/.arnesia/portafolio.json; NO falla por contenido corrupto
func (s *Store) Listar() ([]domain.EntradaPortafolio, []EntradaCorrupta)
func (s *Store) Upsert(e domain.EntradaPortafolio) error      // merge por Identidad; instalaciones dedup por InstallPath
func (s *Store) Desvincular(clave string) (bool, error)       // quita del registro; NO toca disco de proyectos (C-UNL-3)
func (s *Store) Checkouts() []string                          // paths de canónicos conocidos (RN-IDENT-4)
```
- Envelope `{"version":1,"entradas":[…]}`; decode entry-wise; corrupta → visible + saltada; save atómico.
- `Upsert` RECHAZA fusionar una identidad provisional con una resuelta (C-ID-2): son claves distintas, y
  no existe (en Slice 0) operación de fusión — solo el error/aviso explícito.
- Dos canónicos para la misma identidad → error explícito (C-N-5).

**`deriva.go`** — evaluador (S0-D7, BR-4):

```go
// HashFormaPlugin: sha256 agregado del árbol (paths relativos ordenados con '/' + contenido);
// excluye .git/, .in_use, .orphaned_at. Determinista (test con fixture golden).
func HashFormaPlugin(dir string) (string, error)

type Referencias struct{ CCPluginsDir string } // lee known_marketplaces.json
// RutaReferencia: home canonicalizado + id + version → dir local "<installLocation>/plugins/<id>/<version>/"
// si el marketplace está clonado localmente y el dir existe; ok=false si no hay referencia accesible.
func (r *Referencias) RutaReferencia(home, id, version string) (dir string, ok bool)

// EvaluarDeriva: sin version o sin referencia → (DerivaNoEvaluable, motivo). Con referencia:
// hash(install) == hash(ref) → al-hilo; ≠ → en-deriva. Commits locales del espejo = deriva, no autoría (C-N-16).
func EvaluarDeriva(installDir string, refs *Referencias, home, id, version string) (domain.EstadoDeriva, string)
```

### 2.6 Puertos — `internal/ports/portafolio.go`

```go
type ArnesLoader interface { Load(dir string) (domain.Graph, error) } // satisface adapter loader vía wrapper en cmd
type PortafolioScanner interface { Escanear(ctx context.Context, root string) ([]portafolioHallazgo…, error) }
type PortafolioStore interface { Listar() (…); Upsert(…); Desvincular(…); Checkouts() []string }
```
(El tipo Hallazgo vive en `domain` o en `ports` — decidí `domain.HallazgoInstalacion` para no duplicar;
el adapter lo emite, el usecase lo consume. Mantené UNA definición.)

### 2.7 Usecase — `internal/usecase/portafolio.go`

```go
type PortafolioService struct { store ports.PortafolioStore; scan ports.PortafolioScanner; cargar ports.ArnesLoader; refs … }
// Escanear: walk → por hallazgo: cargar (loader, fallback plugin.json) → ResolverIdentidad →
// ResolverOrigen(eslabones) → clasificar RN-IDENT-4 (dir ∈ store.Checkouts() ⇒ candidato CANÓNICO,
// jamás instalación) → EvaluarDeriva → []Candidato. NO persiste (el user elige — spec §7.1).
func (s *PortafolioService) Escanear(ctx, root string) ([]Candidato, error)
// AgregarProyecto: re-escanea root y persiste SOLO los candidatos cuyos Clave() ∈ elegidos (TOCTOU-safe);
// idempotente por identidad+installPath (C-P-8: re-agregar = re-escanear, sin duplicar).
func (s *PortafolioService) AgregarProyecto(ctx, root string, elegidos []string) ([]domain.EntradaPortafolio, error)
func (s *PortafolioService) Listar(ctx) ([]domain.EntradaPortafolio, []EntradaCorrupta, error)
func (s *PortafolioService) Desvincular(ctx, clave string) (bool, error)
```

Validación de `root` reusa la política de `arnes_registry.validate` (abs · existe · no-protegido) —
extraé el helper compartible o duplicá la política mínima con test (NO registres el proyecto en
`arneses.json`: son stores separados; el registro cwd solo pasa cuando se abra en Mapa, Slice 1).

### 2.8 Transporte + CLI (S0-D9)

- `internal/adapters/transport/http/portafolio.go`: los 4 endpoints; router.go los registra bajo `/api`.
  Respuestas honestas: corruptas visibles en `GET /api/portafolio` como bloque aparte
  `{"entradas":[…],"corruptas":[{"motivo":…}]}`. Errores 400 con motivo (path relativo/no existe).
- `cmd/arnesia/main.go`: subcomando `portafolio` (`escanear <dir>` imprime candidatos JSON ·
  `listar` · `agregar <dir> <clave>…` · `desvincular <clave>`), cablea Scanner/Store/loader-wrapper.
- `docs/architecture/contracts/api/openapi.yaml`: agregar los paths (mismo nivel de detalle que los existentes).

## 3 · Tickets (ordenados por dependencia — ejecutá EN ORDEN, un commit por ticket)

> Formato de trabajo por ticket: (1) leé los anclajes; (2) escribí PRIMERO los tests del ticket (rojos);
> (3) implementá hasta verde; (4) corré el GATE LOCAL (§M); (5) commit `feat(portafolio): T<N> — <qué>`.

### T1 · Dominio: identidad + origen + deriva + canonicalización (D-DOM-1/2/7 parte pura)
- **Archivos:** `internal/domain/portafolio.go` · `internal/domain/repo_ref.go` · `internal/domain/graph.go`
  (Version, Empresas+Unmarshal tolerante, FuenteManifiesto) · tests `internal/domain/portafolio_test.go`,
  `repo_ref_test.go`, `graph_test.go` (ampliar).
- **Tests obligatorios (table-driven):**
  - `TestCanonicalizarRepo` — ≥10 casos: slug, https, https+.git, git@, ssh://, trailing /, mayúsculas,
    no-parseable (ok=false), host no-github, path con 3+ segmentos.
  - `TestResolverOrigen` — lock>manifiesto para registry (F1) · cc-plugins aporta version · conflicto
    lock-vs-cc y manifiesto-home-vs-lock-registry → Discrepancias pobladas (C-OR-6) · solo git-proyecto
    → registry sigue vacío (BR-5/C-OR-4) · nada → todo vacío honesto (C-OR-5).
  - `TestResolverIdentidad` — con marketplace → (home,id) · sin → provisional+scope (RN-IDENT-2) ·
    precedencia id RN-IDENT-3 + aviso en discrepancia · `Clave()` estable y sin `/`.
  - `TestArnesEmpresasTolerante` — unmarshal `"empresa":"x"` → `["x"]`; `"empresas":["a","b"]` pasa;
    marshal emite solo `empresas`.
- **Gate:** dominio sigue sin imports fuera de stdlib (go-arch-lint verde).

### T2 · Contrato de datos: schema + manifiestos + seed + FE types (S0-D3 propagación)
- **Archivos:** `docs/architecture/contracts/schema/graph.l0.schema.json` (arnes: `empresas[]` + anyOf legacy
  `empresa` · `version` · `fuente_manifiesto`) · `dogfood/*/arnes.l0.json` y los del kit (`empresa`→`empresas`)
  · `internal/adapters/index/store.go` seed (`Empresas: []string{"alpacapurpura"}`) ·
  `web/src/entities/arnes/model/types.ts` + `web/src/shared/api/types.ts` (`empresa?: string` →
  `empresas?: string[]`, `version?`, `fuente_manifiesto?`) + los ~10 usos FE (grep `empresa` en `web/src`;
  display = `empresas[0] ?? '—'` donde hoy muestra el escalar; fixtures `entities/arnes/testing/*`).
- **Verificación:** `go run ./cmd/arnesia conformance --arnes dogfood/dev-full-cycle` sin fail nuevo ·
  `cd web && pnpm run verify` verde (⚠ correlo en foreground: vitest-browser no corre headless en bg).

### T3 · Loader: fallback `plugin.json` + version + fuente (D-DOM-4, C-N-14 — bloqueante de todo lo demás)
- **Archivos:** `internal/adapters/loader/loader.go` (§2.4) + `loader_test.go` + fixtures en
  `internal/adapters/loader/testdata/` (plugin-solo-pluginjson/ · plugin-ambos-manifiestos/ ·
  plugin-pluginjson-corrupto/ · plugin-ids-discrepantes/).
- **Tests:** `TestLoaderFallbackPluginJSON` (id/nombre/descr/version poblados, FuenteManifiesto="plugin.json")
  · `TestLoaderVersionDesdePluginJSON` (con arnes.l0 presente igual toma version) ·
  `TestLoaderPluginJSONCorrupto` (degradado visible, no error) · `TestLoaderIDsDiscrepantes` (aviso, gana arnes.l0).
- **Regresión:** TODOS los tests previos del loader intactos.

### T4 · Scanner: walker + lock + metadata CC + git (D-DOM-3/6, S0-D1/D2/D8)
- **Archivos:** `internal/adapters/portafolio/{scanner,gitconfig}.go` + tests + `testdata/` con proyectos
  fixture: `proyecto-simple/` (.claude/ poblado) · `proyecto-materializado/` (.claude/plugins/x/ + uno roto
  C-P-5) · `proyecto-lock/` (.devstudio/arneses.yaml + entrada ausente) · `proyecto-cc/` (settings
  enabledPlugins + fixture de CCPluginsDir con installed_plugins.json v2 + known_marketplaces.json + cache) ·
  `monorepo/` (2 .claude/ anidados + symlink circular) · `worktree/` (.git archivo gitdir:) ·
  fixture con `installed_plugins.json` `"version": 99` (→ eslabón no-legible).
- **Tests:** `TestScannerProyectoInstalado` · `TestScannerMaterializada` (+`TestScannerPluginRotoVisible`) ·
  `TestScannerLockDevstudio` (+entrada ausente visible) · `TestScannerReferenciadaCC` (cruza projectPath,
  version, marketplace→repo) · `TestScannerCCMetadataIlegible` (no-legible, no crash) ·
  `TestScannerMonorepoAcotado` (depth, symlinks no seguidos, cancelación ctx) · `TestGitConfigParse`
  (origin/upstream/worktree-gitdir/bare) · `TestScannerRootInvalido` (no existe / relativo → error honesto).

### T5 · Deriva: hasher + referencias locales (S0-D7, BR-4)
- **Archivos:** `internal/adapters/portafolio/deriva.go` + tests + fixtures (mini-árbol golden; par
  install/ref idénticos y con 1 byte distinto; ref con `.in_use`).
- **Tests:** `TestHashFormaPluginDeterminista` (golden estable, excluye .git/.in_use/.orphaned_at) ·
  `TestEvaluarDerivaAlHilo` · `TestEvaluarDerivaEnDeriva` · `TestEvaluarDerivaNoEvaluable` (sin version ·
  sin marketplace local · version inexistente en plugins/<id>/) · `TestDerivaNuncaSemver` (dos árboles
  distintos con MISMA version en plugin.json ⇒ en-deriva — el veredicto sale del hash, no del string).

### T6 · Store + usecase + puertos (S0-D5/D6, BR-11, RN-IDENT-4)
- **Archivos:** `internal/adapters/portafolio/store.go` · `internal/ports/portafolio.go` ·
  `internal/usecase/portafolio.go` + tests de ambos.
- **Tests:** `TestStoreDegradaHonesto` (archivo con 1 entrada corrupta y 2 sanas → 2 listadas + 1 corrupta
  visible; save posterior NO pierde la corrupta cruda — o la descarta EXPLÍCITAMENTE documentado; decidí
  conservarla) · `TestStoreUpsertMergePorIdentidad` (misma identidad 2 proyectos → 1 entrada 2 instalaciones,
  C-P-10; re-upsert mismo installPath → no duplica, C-P-8/C-N-3) · `TestStoreNoFusionaProvisional` (C-ID-2)
  · `TestStoreDobleCanonico` (error, C-N-5) · `TestStoreSaveAtomico` ·
  `TestServiceEscanearClasificaCheckout` (dir ∈ Checkouts() ⇒ canónico, jamás instalación — RN-IDENT-4/C-N-12)
  · `TestServiceAgregarSoloElegidos` · `TestServiceDesvincularNoTocaDisco` (C-UNL-3: los archivos del
  proyecto fixture siguen intactos) · `TestServiceRootProtegido` (rechaza $HOME/raíz — política registry).

### T7 · HTTP + CLI + openapi + wiring (S0-D9)
- **Archivos:** `internal/adapters/transport/http/portafolio.go` + router.go · `cmd/arnesia/main.go`
  (subcomando + wiring: `portafolio.NewStore` default `~/.arnesia/portafolio.json`, Scanner default,
  wrapper `ports.ArnesLoader` sobre `loader.LoadArnes`) · `docs/architecture/contracts/api/openapi.yaml`.
- **Tests:** handlers con `httptest` (`TestPortafolioListIncluyeCorruptas` · `TestPortafolioEscanearNoPersiste`
  · `TestPortafolioAgregar` · `TestPortafolioDesvincular404`); el subcomando se prueba en T9 vivo.
- **Nota:** boot del daemon: NADA del portafolio puede impedir `serve` (BR-11) — store corrupto entero ⇒
  log warn + portafolio vacío + corruptas visibles.

### T8 · As-code: capabilities + boundary + seams (S0-D11)
- **Capabilities** (`docs/product/capabilities/portafolio/*.yaml`, template obligatorio; `cap_num` = seguí
  del máximo actual, verificá con grep): `registrar-identidad` (pointers: domain.IdentidadArnes,
  CanonicalizarRepo, store.Upsert; valida: TestResolverIdentidad, TestStoreUpsertMergePorIdentidad…) ·
  `escanear-proyecto` · `resolver-origen` · `evaluar-deriva` · `desvincular` — cada una con scenarios BDD
  citando los C-XX que cubre y `valida:` = tests REALES de T1-T7 (R4: vivo ⟹ valida no vacío).
  `change_log: extend` en CAP-15/16/17 (loader fallback). Regenerá el índice: `python3 scripts/cap_doctor.py --index`.
- **Seam:** `project.config.yaml` → `domain_modules` += `portafolio`.
- **Boundary nuevo:** `docs/architecture/boundaries/portafolio-identidad-y-deriva-honesta.md` (frontmatter
  como los existentes; status `enforced`; ledger HS-22): L1 = content-addressable verification (hash-based
  integrity, git/Nix/OCI) + living documentation; L2 = este árbol; checklist 4 checks →
  `identidad-calificada-unica` (TestStoreNoFusionaProvisional) · `deriva-nunca-semver` (TestDerivaNuncaSemver)
  · `store-degrada-honesto` (TestStoreDegradaHonesto) · `procedencia-anotada` (TestResolverOrigen).
  Actualizá la tabla de `docs/architecture/INDEX.md` (+1 nodo, conteo de checks — el bloque de cifras del
  checkpoint lo regenera estado.sh, NO lo tecleés).
- **go-arch-lint:** componente `portafolio` + deps (S0-D4). **_coverage.yaml:** sumá
  `internal/adapters/transport/http/portafolio.go` a support_files SOLO si no lo reclamás por puntero de
  capability (preferí puntero de capability — es funcional, no soporte).

### T9 · Verificación final E2E viva + PARIDAD + cierre
1. **Batería completa** (§M) + `go run ./cmd/arnesia conformance --todo` (pass ≥42 · fail 0).
2. **E2E vivo (máquina real, read-only — no persistas nada fuera de `~/.arnesia/portafolio.json`):**
   - `go run ./cmd/arnesia portafolio escanear ~/Proyectos/harness-studio` → esperado: hallazgo
     `proyecto-instalado` + `referenciada-cc` por cada enabledPlugins:true real; para el plugin del kit
     (`harness@prenter-marketplace`): version REAL (cotejá vos contra `~/.claude/plugins/installed_plugins.json`),
     home/registry `github.com/alpacapurpura/prenter-marketplace`, deriva evaluada contra
     `~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/<version>/` (si el layout
     por-versión existe ahí ⇒ al-hilo/en-deriva REAL; si no ⇒ `deriva-no-evaluable` con motivo — cualquiera
     de los dos es válido, lo INVÁLIDO es fabricar).
   - `… escanear ~/Proyectos/luana-vitalia` → `proyecto-instalado` con identidad provisional
     (⟂, "luana-vitalia", scope) + origen `desconocido` + deriva-no-evaluable — TODO honesto.
   - `agregar` + `listar` + `desvincular` E2E; verificá `~/.arnesia/portafolio.json` a mano; corrompé una
     entrada a mano y verificá degradación honesta en `listar` y que `serve` bootea.
   - `bin/arnesia serve` (o `go run ./cmd/arnesia serve`) + `curl` los 4 endpoints.
   - **Pegá las salidas REALES (recortadas) en `paridad.md`** como evidencia.
3. **`paridad.md`** en este paquete: tabla spec-item (§9 + D-DOM-1..7 + BR-1..11 aplicables + S0-D1..D11) →
   estado (✅/desviación/deferred) → evidencia (test/salida). Desviaciones = visibles, jamás maquilladas.
   Firma 🧑‍⚖️ queda PENDIENTE (gate humano del operador).
4. **Cierre documental:** `INDEX.md` de este paquete («Retomar aquí» → gate PARIDAD pendiente) ·
   `docs/product/checkpoint.md` (paquete activo: Slice 0 construido, pendiente firma; cifras via estado.sh)
   · BACKLOG: item 0 → marcar construido-pendiente-firma + alta deuda «re-key índice calificado (S0-D6)»
   y «conservación de entradas corruptas: política definitiva» si aplica · memoria (si tu sesión la tiene).

## M · GATE LOCAL (correr tras CADA ticket; TODO verde antes de commitear)

```bash
go build ./... && go test ./... -race
golangci-lint run
go run github.com/fe3dback/go-arch-lint@latest check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml
go test ./docs/architecture/fitness/...
bash scripts/estado.sh --check || bash scripts/estado.sh   # regenera cifras si driftaron
# T2 y T9 además:  cd web && pnpm run verify
```
El hook lefthook `capabilities` fallará tus commits desde T1 hasta que T8 registre las caps — para los
commits intermedios T1-T7 tocá las capability-YAML del módulo en el MISMO commit que crea archivos nuevos
(alta mínima con pointers reales; T8 las completa) O creá las 5 hojas en T1 con `status: stub` y pointers
al primer símbolo real, completándolas por ticket. Elegí la segunda (más honesta: el SSoT crece con el código).

## P · Prohibiciones (anti-drift del ejecutor)

1. NO relitigar decisiones firmadas (HS-22) ni las S0-D de este paquete — si algo no cierra en código,
   documentá en `decisiones.md` y tomá la alternativa MÁS CERCANA al espíritu (honestidad > limpieza).
2. NO deps nuevas en go.mod. NO go-git. NO SQLite.
3. NO tocar FE más allá de la migración de tipos/usos `empresas[]` (T2). Cero superficie FE nueva (Slice 1).
4. NO clonar, NO publicar, NO reparar, NO auth GitHub (Slices 2-5). NO red en ninguna ruta de Slice 0.
5. NO editar instalaciones ni escribir dentro de proyectos escaneados: el scan es READ-ONLY estricto.
6. NO fabricar estados: sin dato = `?`/`desconocido`/`deriva-no-evaluable` + motivo. Jamás «al día» sin evidencia.
7. NO teclear cifras (checkpoint/INDEX): las genera `estado.sh`. NO teclear `status:` de caps que contradiga R4.
8. NO firmar PARIDAD ni gates humanos — eso es del operador.
9. Los tests NO tocan la máquina real (todo por fixtures/inyección de CCPluginsDir); SOLO T9-E2E lee la
   máquina real, y solo lee.
