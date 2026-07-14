# Plan de implementación — Portafolio · Slice 1 «FE Portafolio» (3 superficies sobre cimientos)

> `tipo: plan` · paquete `2026-07-13-portafolio-slice1-fe` · autor: Fable 5 (arquitecto/líder técnico) 2026-07-13.
> Ejecuta: **Sonnet 5** (constructor). El modelo funcional está FIRMADO (HS-22), los cimientos FIRMADOS
> (HS-23), y las decisiones técnicas cerradas en [`decisiones.md`](./decisiones.md) (S1-D1..D15) — acá NO
> se decide, se CONSTRUYE. Specs fuente: `../2026-07-10-spike-carga-arneses/spec-funcional.md` (§10) ·
> `spec-usabilidad.md` · `casuistica.md` §I · `revision-adversaria.md` **G1-G9 (ítems de build
> OBLIGATORIOS, no opcionales)** · BACKLOG «Programa Portafolio» item 1. Contratos = el código REAL de
> Slice 0 (la spec pre-build ya demostró desviarse ~60%; acá los shapes están copiados del código).

## 0 · Goal (definición de terminado — verificable, no negociable)

Slice 1 está TERMINADO cuando **todo** esto es cierto (nada fabricado; gap = visible + documentado):

1. La ruta global `#/portafolio` renderiza la vista REAL (muere su `ComingSoon` en `GlobalView`):
   Lista con lente empresa (+ grupo «sin empresa») y lente plano, buscar, entradas del
   `~/.arnesia/portafolio.json` real vía `GET /api/portafolio`, **corruptas visibles** (BR-11),
   estados vacío/cargando/error/filtro-sin-resultados (G5).
2. El Wizard rama Proyecto/carpeta-local funciona E2E contra el backend real: `POST
   /api/portafolio/escaneos` (spinner + cancelar) → candidatos honestos (`v?`, «origen desconocido»,
   avisos, `es_canonico`, «ya en el portafolio») → `POST /api/portafolio/proyectos` → la Lista se
   refresca. Rama GitHub y tab Marketplace = disabled + tooltip S2, SIN validación fingida (G3).
3. El Drawer READ muestra por entrada: identidad calificada (home+id o provisional+scope, G6), facetas
   empresas/marketplaces honestas (G4, S1-D3), callout ley anti-drift, zona Canónico (ausente → CTA
   disabled S2), zona Instalaciones (tipo · deriva REAL con detalle · aviso · origen-de-la-copia con
   discrepancias y trazabilidad de eslabones), «update: no-verificado» (G2), acciones staged con
   tooltip de slice (S2/S3/S5), línea de frescura «datos del último escaneo» (S1-D15).
4. Desvincular funciona E2E con confirmación (G7): dialog explícito + checkbox «…y borrar clon local»
   disabled S2 → `DELETE /api/portafolio/arneses/{clave}` → 404 honesto si repite.
5. «Abrir en Mapa» funciona E2E **incluida una instalación `referenciada-cc` real bajo
   `~/.claude/plugins/cache/…`** (la prueba del GAP-1): `POST /api/portafolio/arneses/{clave}/mapa`
   (S1-D1) indexa sin registrar cwd → el Mapa (CAP-58/61) renderiza el grafo REAL del plugin →
   la tab fuente del inspector degrada con su mensaje honesto existente. Colisión de bare-id → diálogo
   de confirmación (S1-D2). Sin sesión activa → botón disabled + tooltip (S1-D13).
6. **Los 9 fixes G1-G9 aplicados y evidenciados** — cada uno mapeado a ticket en la tabla de §3.0 y a
   una story/test nombrado que lo prueba. La a11y (G8) es criterio de aceptación CON test (play asserts
   de rol/foco/Esc + axe del addon-a11y que rompe CI), no una nota.
7. Storybook = SSoT: stories nuevas `portafolio-list` · `portafolio-drawer` · `portafolio-wizard` ·
   chips de entidad, cada una con `play()` que asserta (story = test, fe-visual-fitness);
   `pnpm exec vitest --project=storybook run` verde EN FOREGROUND (gotcha HS-20) y
   `--project=unit run` verde (selectores G9/G4/G6, S1-D7).
8. Gates verdes: `go build ./... && go test ./... -race` · `golangci-lint run` · go-arch-lint ·
   `go test ./docs/architecture/fitness/...` · `bash scripts/estado.sh --check` ·
   `conformance --todo` sin regresión (pass ≥ 42 · fail 0) · `cd web && pnpm run verify`.
9. Capabilities registradas (T8): 4 nuevas `fe-portafolio/*` + `portafolio/observar-en-mapa` + extends;
   `mockups/arnesia-portafolio.html` corregido (G1-G9) + fila de `mockups/INDEX.md` re-estampada
   (S1-D12); `paridad.md` escrita con tabla spec↔realidad + evidencia E2E viva (firma humana PENDIENTE
   para el operador — no la simules).
10. Un commit Conventional por ticket en `main` (`feat(portafolio): S1-T<N> — <qué>`), gate §M verde
    ANTES de cada commit.

## 1 · Contexto mínimo que debés leer antes de tocar código

1. [`decisiones.md`](./decisiones.md) de este paquete — S1-D1..D15 + los 3 GAPS (cierra todo lo abierto).
2. `../2026-07-10-spike-carga-arneses/spec-funcional.md` §10 + `spec-usabilidad.md` (completa) +
   `revision-adversaria.md` §G (G1-G9) + `casuistica.md` §A/§I (15 min).
3. **`mockups/INDEX.md` PRIMERO** (regla dura), después `mockups/arnesia-portafolio.html` (la propuesta
   aprobada — vocabulario visual y layout; sus DATOS están fabricados, no los copies).
4. Código ancla backend: `internal/usecase/portafolio.go` · `internal/adapters/portafolio/store.go`
   (`Upsert`) · `internal/adapters/transport/http/portafolio.go` + `router.go` · `internal/ports/{portafolio,index}.go`
   · `cmd/arnesia/main.go` (`newPortafolioService`, `loadArnesDir`).
5. Código ancla FE: `web/src/pages/shell/ui/global-view.tsx` (patrón AjustesView = composition-root
   local) · `workspace-stage.tsx` (picker/viewedId del Mapa) · `shared/api/client.ts` + `types.ts` ·
   `shared/store/app-store.ts` · `features/self-update/ui/update-card.stories.tsx` (EL patrón de
   story-as-test a calcar) · `app/styles/ajustes.css` (patrón de css scoped por vista).
6. Boundaries FE: `docs/architecture/boundaries/fe-{topologia-fsd,taxonomia-componentes,
   transporte-independiente,visual-fitness,tokens-contrato}.md`.
7. `docs/product/_templates/capability.template.yaml` + `docs/architecture/boundaries/codigo-traza-a-capability.md` (R1-R4).
8. `../2026-07-13-portafolio-slice0-cimientos/paridad.md` §E2E — los DATOS REALES de esta máquina
   (harness@prenter-marketplace v0.5.2 `en-deriva`, identidades provisionales, avisos «sin record») —
   la materia prima de tus fixtures.

## 2 · Diseño técnico (contratos exactos)

### 2.1 Wire vigente (Slice 0, copiado del código — NO lo cambies salvo lo que T1 agrega)

```
GET    /api/portafolio                      → {entradas: [Entrada+{clave}], corruptas?: [{motivo}]}
POST   /api/portafolio/escaneos {path}      → [Candidato]            | 400 {error}
POST   /api/portafolio/proyectos {path, elegidos[]} → [Entrada SIN clave] | 400 {error}
DELETE /api/portafolio/arneses/{clave}      → {desvinculado: true}   | 404 {error}
```
`Candidato` (usecase/portafolio.go): `{clave, identidad{home?,id,scope?}, nombre?, descripcion?,
empresas?, instalacion, es_canonico?}`. `Instalacion`: `{proyecto_path, install_path, tipo, origen,
deriva, deriva_detalle?, aviso?}` con `tipo ∈ {materializada, referenciada-cc, proyecto-instalado}`,
`deriva ∈ {al-hilo, en-deriva, deriva-no-evaluable}`, `origen = {registry?, version?, eslabones?[],
discrepancias?[]}`. Tras un POST /proyectos exitoso el FE **re-fetchea el GET** (la respuesta del POST
no trae `clave` — nota menor de decisiones.md, no la «arregles»).

### 2.2 Backend nuevo — T1 (S1-D1 + S1-D3, lo ÚNICO de Go en este slice)

```go
// internal/usecase/portafolio.go
type PortafolioService struct { store …; scan …; cargar …; deriva …; indice ports.IndexPort } // 5° puerto
func NewPortafolioService(store, scan, cargar, deriva, indice) *PortafolioService // cmd: serve→idx real, CLI→nil

// ObservarEnMapa publica una presencia PERSISTIDA al índice del Mapa, read-only (S1-D1):
// clave→entrada (no está → domain.ErrNoEncontrado-style para que el handler dé 404);
// installPath ∈ {instalaciones[].InstallPath, canonico.Path} comparado canónico (EvalSymlinks+Clean)
// — ajeno → error 400; indice==nil → error «requiere el daemon»; cargar.Load(dir) err o Arnes==nil →
// error con motivo. OK → indice.Upsert(ctx, g) y devuelve g.Arnes.ID (el bare id EFECTIVO — S1-D2).
func (s *PortafolioService) ObservarEnMapa(ctx context.Context, clave, installPath string) (string, error)

// AgregarProyecto — delta S1-D3: al construir la entrada,
//   if r := c.Instalacion.Origen.Registry; r != "" {
//       canon, ok := domain.CanonicalizarRepo(r); if !ok { canon = r } // crudo visible
//       e.Registries = []string{canon}
//   }
```

```go
// internal/adapters/portafolio/store.go — Upsert merge (S1-D3): facetas por UNIÓN
merged.Empresas   = unionDedup(existing.Empresas, e.Empresas)     // orden estable, vacía no borra
merged.Registries = unionDedup(existing.Registries, e.Registries)
```

```go
// internal/adapters/transport/http/portafolio.go + router.go
// POST /api/portafolio/arneses/{clave}/mapa   body {"install_path": "..."}
//   200 {"id": "...", "indexed": true} · 404 clave desconocida · 400 path ajeno / no cargable /
//   500 sin índice. Registrar en router.go junto a los otros 4.
```

`docs/architecture/contracts/api/openapi.yaml`: agregar el path (mismo nivel de detalle que los
existentes; bump del comment de versión como hizo Slice 0).

### 2.3 Tipos FE — `web/src/entities/portafolio/model/types.ts` (espejo EXACTO del wire)

```ts
export type EstadoDeriva = "al-hilo" | "en-deriva" | "deriva-no-evaluable"
export type TipoInstalacion = "materializada" | "referenciada-cc" | "proyecto-instalado"
export interface EslabonOrigen { fuente: string; campo: string; valor: string }
export interface OrigenCopia { registry?: string; version?: string; eslabones?: EslabonOrigen[]; discrepancias?: string[] }
export interface IdentidadArnes { home?: string; id: string; scope?: string }
export interface Instalacion {
  proyecto_path: string; install_path: string; tipo: TipoInstalacion
  origen: OrigenCopia; deriva: EstadoDeriva; deriva_detalle?: string; aviso?: string
}
export interface Canonico { path: string; version?: string }
export interface EntradaPortafolio {
  clave: string; identidad: IdentidadArnes; nombre?: string; descripcion?: string
  empresas?: string[]; registries?: string[]; canonico?: Canonico
  instalaciones?: Instalacion[]; agregado?: string
}
export interface EntradaCorrupta { motivo: string }
export interface PortafolioListado { entradas: EntradaPortafolio[]; corruptas?: EntradaCorrupta[] }
export interface Candidato {
  clave: string; identidad: IdentidadArnes; nombre?: string; descripcion?: string
  empresas?: string[]; instalacion: Instalacion; es_canonico?: boolean
}
export type SaludPortafolio = "ok" | "atencion" | "sin-senal"
export type LentePortafolio = "empresa" | "plano"
```
(El tipo TS se llama `OrigenCopia` para no chocar con el `Origen` L0 de `entities/arnes` — mismo
motivo que `domain.OrigenPortafolio`, S1-D14. El campo json sigue siendo `origen`.)

### 2.4 Selectores puros — `entities/portafolio/model/selectors.ts` (unit-tested, S1-D7)

```ts
export function saludDe(e: EntradaPortafolio): SaludPortafolio        // regla EXACTA S1-D4
export function registriesDe(e: EntradaPortafolio): string[]          // entry.registries ∪ inst.origen.registry, dedup
export function agruparPorEmpresa(es: EntradaPortafolio[]): { grupo: string; entradas: EntradaPortafolio[] }[]
  // N:M (una entrada en N grupos); grupo «sin empresa» SIEMPRE al final si aplica (S1-D8)
export function filtrarEntradas(es: EntradaPortafolio[], q: string): EntradaPortafolio[] // id/nombre/descripcion, case-insensitive
export function idsColisionados(es: EntradaPortafolio[]): Map<string, string[]>          // id → claves (S1-D2)
```

### 2.5 API client — `shared/api/client.ts` (genéricos, domain-free, patrón `getGraph`)

```ts
listPortafolio: <T = unknown>() => req<T>("/api/portafolio"),
escanearProyecto: <T = unknown>(path: string, signal?: AbortSignal) =>
  req<T>("/api/portafolio/escaneos", { method: "POST", body: JSON.stringify({ path }), signal }),
agregarProyecto: <T = unknown>(path: string, elegidos: string[]) =>
  req<T>("/api/portafolio/proyectos", { method: "POST", body: JSON.stringify({ path, elegidos }) }),
desvincularDelPortafolio: <T = unknown>(clave: string) =>
  req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}`, { method: "DELETE" }),
observarEnMapa: <T = unknown>(clave: string, installPath: string) =>
  req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/mapa`,
    { method: "POST", body: JSON.stringify({ install_path: installPath }) }),
```

### 2.6 Widgets (props puras — el transporte vive en la página; contratos exactos)

```ts
// widgets/portafolio/ui/portafolio-list.tsx
interface PortafolioListProps {
  estado: "cargando" | "error" | "datos"
  error?: string | undefined
  entradas: EntradaPortafolio[]
  corruptas: EntradaCorrupta[]
  lente: LentePortafolio; onLente: (l: LentePortafolio) => void
  busqueda: string; onBusqueda: (q: string) => void
  seleccionada?: string | undefined                 // clave de la fila abierta en el drawer
  onAbrir: (clave: string) => void                  // fila = <button>, Enter/click
  onAgregar: () => void                             // abre el wizard
  onReintentar: () => void                          // estado error
}
```
Fila: emblema (inicial, color determinista por token `--c-*`) · id mono + nombre/descr · presencia
(`◆ canónico vX` | `◇ sin canónico` · `▣ N instalac.`) · chips honestos (`en-deriva` si aplica ·
`origen?` si ningún registry · aviso) · dot salud con `aria-label` (S1-D4). **Sin flag de update**
(G2). Topbar de la vista: título + contadores REALES (`N arneses · M empresas · lente: X`) + `＋ Agregar`.
Banner de corruptas cuando `corruptas.length > 0`: «N entrada(s) corrupta(s) en el registro — <motivo>»
(BR-11, visible, no modal). Lentes proyecto/marketplace + filtros estado/marketplace: disabled+tooltip.

```ts
// widgets/portafolio/ui/portafolio-drawer.tsx
interface PortafolioDrawerProps {
  entrada: EntradaPortafolio
  onClose: () => void
  onObservar?: ((installPath: string) => void) | undefined // undefined ⇒ disabled + tooltip (S1-D13)
  observarDisabledMotivo?: string | undefined
  desvinculando?: boolean | undefined
  desvincularError?: string | undefined
  onDesvincular: () => void                                // la página hace el DELETE
}
```
Estructura (spec-usabilidad §4 + S1-D5/D14/D15): header (emblema·id·nombre·✕) · línea identidad
(`home` canonicalizado o «identidad provisional (sin home) · scope: <scope>») · facetas empresas
(chips o «desconocida») / marketplaces (`registriesDe()` o «desconocido») · callout ley anti-drift ·
«update: no-verificado» (tooltip S4) · zona Canónico (presente: path+version+«estado del checkout: no
evaluado en este slice», acciones `◉ Abrir en Mapa` real / `✎ Mejorar` S2 / `▲ Publicar` S3; ausente:
CTA `↧ Traer canónico` disabled S2) · zona Instalaciones (por cada una: tipo badge · proyecto ·
install_path · `v<version>|v?` · chip deriva + `deriva_detalle` como `title` · chip aviso · origen de
la copia: registry o «desconocido» + discrepancias VISIBLES + `<details>` trazabilidad con los
eslabones crudos · acciones `◉ Observar en Mapa` real / `⚒ Reparar` S5 / `↩ Backport` S5) · línea
frescura «datos del último escaneo — agregado <fecha>» · footer Desvincular con **confirm interno**
(G7): dialog con copy «solo lo saca del portafolio — no desinstala ni borra nada del disco» + checkbox
«…y borrar clon local» disabled+tooltip S2 + Confirmar/Cancelar. A11y (G8): `role="dialog"`
`aria-modal="true"` `aria-labelledby`, foco inicial al cerrar-botón, trap de Tab, Esc cierra, al cerrar
devuelve el foco a la fila (la página pasa `seleccionada` y la lista re-enfoca).

```ts
// widgets/portafolio/ui/portafolio-wizard.tsx
interface PortafolioWizardProps {
  abierto: boolean
  onClose: () => void                                    // cancelar = CERO efectos
  estado: "fuente" | "escaneando" | "candidatos" | "agregando"
  error?: string | undefined                             // motivo 400 del backend, textual
  candidatos?: Candidato[] | undefined
  clavesExistentes: ReadonlySet<string>                  // badge «ya en el portafolio» (S1-D10)
  onEscanear: (path: string) => void
  onCancelarEscaneo: () => void                          // aborta el fetch (S1-D9)
  onAgregar: (elegidos: string[]) => void
  onElegirCarpeta?: (() => Promise<string | undefined>) | undefined // solo Tauri (S1-D9)
}
```
Paso 0 tabs Proyecto/Marketplace (Marketplace disabled+tooltip S2, G3) · paso 1 fuente (input path +
«Elegir carpeta…» si hay picker; radio GitHub disabled S2) · paso 2 candidatos (checkbox por candidato,
render S1-D10; 0 hallazgos → copy C-P-4; error → motivo C-P-3) · paso 3 «Agregar N al portafolio» +
nota espejos-read-only. A11y: `role="dialog"` `aria-modal`, tabs con `role="tab"`/`aria-selected`, Esc
cierra, foco atrapado.

### 2.7 Página + Mapa — composition root y peek

- **`pages/shell/ui/portafolio-view.tsx`** (patrón AjustesView): TODO el transporte; máquina de estados
  de la vista (cargando→datos/error, refetch tras agregar/desvincular); estado del wizard (S1-D9:
  `AbortController` para cancelar); confirmación de colisión (S1-D2: `idsColisionados` + `window`-less
  dialog propio, mismo primitivo del confirm de desvincular); `onObservar` = confirmar-si-colisión →
  `api.observarEnMapa` → `setMapaPeek(id)` → `parkView("Mapa")` (store de sesiones) → `setView("mapa")`
  (app-store, ruta no-global). `onObservar` es `undefined` si `useSessions` no tiene sesión activa
  (S1-D13). `onElegirCarpeta` solo `isTauri()` (patrón RF-110).
- **`shared/store/app-store.ts`**: campo nuevo `mapaPeek: string | null` + `setMapaPeek` (doc: puente
  Portafolio→Mapa, un solo consumo).
- **`pages/shell/ui/workspace-stage.tsx`**: `useEffect` — si `mapaPeek` y `isMapa`: re-fetch de
  `listHarnesses` (el guard actual rechaza ids que no estaban en el snapshot), `setViewedId(mapaPeek)`,
  `setMapaPeek(null)`. NO toques nada más del stage.
- **`pages/shell/ui/global-view.tsx`**: `if (route === "portafolio") return <PortafolioView />`; borrá
  la entrada `portafolio` del mapa `GLOBAL` (el ComingSoon muere — G5 lo reemplazan estados reales).
- **`app/styles/portafolio.css`** scoped `.arnesia-portafolio` + import en `app/styles/index.css`;
  tokens DTCG only; `@media (prefers-reduced-motion: reduce)` apaga transiciones (G8).

### 2.8 Fixtures honestas — `entities/portafolio/testing/entradas.ts`

Calcadas de las salidas REALES del E2E de Slice 0 (`paridad.md` §E2E): (a) entrada
`harness@prenter-marketplace` — identidad provisional (el caso real: sin `arnes.l0.json`), registry
`github.com/alpacapurpura/prenter-marketplace`, `version: "0.5.2"`, `deriva: "en-deriva"` con detalle,
tipo `referenciada-cc`, install_path bajo `~/.claude/plugins/cache/…`; (b) entrada provisional
`proyecto-instalado` sin registry (origen desconocido, `deriva-no-evaluable` con motivo); (c) entrada
sintética PERO con shape real para los estados restantes: `al-hilo` sin avisos (salud ok), con
`empresas: ["alpacapurpura"]`, con `canonico` poblado (para las stories de zona canónico), con
`discrepancias` y con `aviso: "declarada-en-lock, ausente"`. Cada fixture comenta de qué salida real
sale o por qué es sintética-con-shape-real. PROHIBIDO inventar campos que el wire no tiene
(`drift: 2`, `update: "1.2.0"`, `dirty: false` — los vicios G1/G2 del mockup).

## 3 · Tickets (ordenados por dependencia — ejecutá EN ORDEN, un commit por ticket)

> Formato por ticket: (1) leé los anclajes; (2) escribí PRIMERO los tests/stories del ticket (rojos);
> (3) implementá hasta verde; (4) corré el GATE LOCAL (§M); (5) commit `feat(portafolio): S1-T<N> — <qué>`.

### 3.0 · Mapa G1-G9 → ticket (los 9 son OBLIGATORIOS; esta tabla se copia a paridad.md con evidencia)

| G | Fix | Ticket(s) | Evidencia (test/story nombrado) |
|---|---|---|---|
| G1 deriva fabricada | chips SOLO desde `EstadoDeriva` real + detalle; muere `drift:N` | T4·T5·T6 | `ConDatos` (list) · `InstalacionEnDerivaReal` (drawer) · `CandidatosReales` (wizard) |
| G2 update fabricado | fila SIN flag; drawer «update: no-verificado» + tooltip S4 | T4·T5 | `ConDatos` asserts ausencia `⬆` · `UpdateNoVerificado` (drawer) |
| G3 marketplace ✓ hardcodeado | tab + rama GitHub disabled + tooltip S2, cero resultado fingido | T6 | `Paso1Fuente` asserts disabled + title |
| G4 empresa del path | empresa SOLO del manifiesto; «sin empresa»/«desconocida» | T2·T4·T5 | `agruparPorEmpresa` unit · `ConDatos` grupo «sin empresa» · `IdentidadProvisional` |
| G5 estados faltantes | vacío/cargando/error/0-hallazgos/corruptas/sin-instalaciones/sin-resultados | T4·T5·T6 | `Vacia`·`Cargando`·`ErrorDeCarga`·`CorruptasVisibles`·`FiltroSinResultados`·`SinInstalaciones`·`SinHallazgos`·`ErrorDePath` |
| G6 keyeado por id desnudo | key = `clave` en todo el FE; identidad provisional VISIBLE | T2·T4·T5 | types + `IdentidadProvisional` (drawer) · `onAbrir(clave)` asserts |
| G7 borrar-clon sin confirmación | confirm dialog; checkbox borrar-clon disabled S2 | T5 | `ConfirmarDesvincular` |
| G8 a11y | drawer dialog+trap+Esc · wizard aria-modal · filas focuseables · reduced-motion · axe=CI | T3-T6 | `FocoYTeclado` (drawer) · `A11yModal` (wizard) · asserts de teclado en `ConDatos` · addon-a11y en TODAS |
| G9 dot de salud sin regla | regla S1-D4 definida + testeada + aria-label | T2·T3·T4 | `saludDe` unit (todas las ramas) · `chips.stories` dot · `ConDatos` |

### T1 · Backend: observar-en-mapa + facetas pobladas (S1-D1/D2/D3 — lo único de Go)
- **Archivos:** `internal/usecase/portafolio.go` (5° puerto `indice`, `ObservarEnMapa`, delta
  `AgregarProyecto`→Registries) · `internal/adapters/portafolio/store.go` (merge unión) ·
  `internal/adapters/transport/http/portafolio.go` + `router.go` (endpoint) · `cmd/arnesia/main.go`
  (wiring: `serve` pasa `idx`, subcomando CLI pasa `nil`) · `docs/architecture/contracts/api/openapi.yaml`
  · tests en `usecase/portafolio_test.go`, `adapters/portafolio/store_test.go`, `transport/http`.
- **Tests obligatorios:** `TestObservarEnMapaIndexaSinRegistro` (fake index recibe Upsert con el grafo;
  cero interacción con ArnesRegistry — estructural: ni siquiera es dep) ·
  `TestObservarEnMapaInstallPathAjeno` · `TestObservarEnMapaClaveInexistente` ·
  `TestObservarEnMapaSinIndice` (nil → error honesto) · `TestObservarEnMapaNoCargable` (loader err →
  error con motivo, jamás grafo inventado) · `TestAgregarProyectoPueblaRegistries` (canonicaliza; crudo
  si no parsea) · `TestUpsertMergeUneFacetas` (unión Empresas/Registries; upsert con faceta vacía NO
  borra) · HTTP `TestPortafolioObservar` (200/404/400 con httptest).
- **Capability:** alta `docs/product/capabilities/portafolio/observar-en-mapa.yaml` (pointers reales:
  `usecase.ObservarEnMapa`, handler HTTP; `valida:` = estos tests; scenarios citando S1-D1/S1-D2 y
  C-N-12). `cap_num` = seguí del máximo actual (verificá con grep).
- **Gate extra:** `conformance --todo` sin regresión; boot del daemon intacto (nada del portafolio
  puede impedir `serve`, BR-11).

### T2 · Contrato FE: entity model + api client + infra unit (S1-D6/D7 + stubs as-code)
- **Archivos:** `web/src/entities/portafolio/model/{types,selectors}.ts` (§2.3/§2.4) ·
  `entities/portafolio/testing/entradas.ts` (§2.8) · `entities/portafolio/index.ts` ·
  `shared/api/client.ts` (§2.5) · `web/vitest.config.ts` (proyecto `unit`: environment node, include
  `src/**/*.test.ts`, extends vite.config) · `.github/workflows/ci.yml` (job `ts`: paso
  `pnpm exec vitest --project=unit run` antes del fitness visual) ·
  `entities/portafolio/model/selectors.test.ts`.
- **Tests obligatorios (unit):** `saludDe` — TODAS las ramas de S1-D4 (en-deriva→atencion ·
  aviso→atencion · discrepancias→atencion · todo al-hilo limpio→ok · todo no-evaluable→sin-senal ·
  0 instalaciones→sin-senal — G9) · `agruparPorEmpresa` (N:M duplica en N grupos · «sin empresa» al
  final — G4) · `registriesDe` (unión entry∪instalaciones, dedup) · `filtrarEntradas` ·
  `idsColisionados` (dos claves mismo id → detectado; ids únicos → mapa vacío — S1-D2).
- **As-code (R3):** alta de las 4 capabilities `fe-portafolio/*` con `status: stub` + pointers al
  primer símbolo real (patrón Slice 0 §M: el SSoT crece con el código; T8 las completa):
  `lista-del-portafolio` · `drawer-detalle-read` · `wizard-agregar-proyecto` · `abrir-en-mapa`.
- **Gate:** `pnpm run verify` + `vitest --project=unit run` verdes.

### T3 · Entity UI: chips/dot del dominio + css base (G8/G9 visuales)
- **Archivos:** `entities/portafolio/ui/chips.tsx` (`DerivaChip` — 3 estados, `deriva_detalle` como
  `title` · `TipoInstalacionChip` · `AvisoChip` · `DotSaludPortafolio` — hueco para `sin-senal`,
  `aria-label="salud: <x>"` · `EmblemaInicial` — color determinista sobre la lista de tokens `--c-*`
  existentes, función pura) + `chips.stories.tsx` · `app/styles/portafolio.css` (scoped
  `.arnesia-portafolio`, tokens only, reduced-motion) + import en `app/styles/index.css`.
- **Stories/tests:** `chips.stories.tsx` con play: los 3 `DerivaChip` renderizan el literal exacto
  (`al-hilo`/`en-deriva`/`deriva-no-evaluable`), el dot expone su `aria-label` (G9), `sin-senal` NO usa
  el color de ok (assert de clase). El addon-a11y corre sobre todas (G8).
- **Gate:** verify + storybook run en foreground.

### T4 · Widget Lista (G1/G2/G4/G5/G6/G8/G9 en la superficie 1)
- **Archivos:** `widgets/portafolio/ui/portafolio-list.tsx` (§2.6) + `portafolio-list.stories.tsx` +
  `widgets/portafolio/index.ts`; css que necesite en `portafolio.css`.
- **Stories obligatorias (cada una con play; fixtures de §2.8):**
  - `Vacia` — «Tu portafolio está vacío» + CTA `＋ Agregar` (dispara `onAgregar`).
  - `Cargando` — skeleton/spinner; sin filas fantasma.
  - `ErrorDeCarga` — motivo textual + botón Reintentar habilitado (dispara `onReintentar`).
  - `ConDatos` — lente empresa: grupo `alpacapurpura` + grupo «sin empresa» (G4); fila con chip
    `en-deriva` (G1) y SIN flag `⬆` (G2: assert `queryByText(/⬆/) === null`); dots con `aria-label`
    (G9); contadores reales; click y Enter en fila llaman `onAbrir(clave)` (G6/G8).
  - `LentePlano` — sin headers de grupo; mismas filas.
  - `CorruptasVisibles` — banner «1 entrada corrupta …» con el motivo (BR-11).
  - `FiltroSinResultados` — búsqueda sin match → «Ningún arnés coincide» + limpiar.
  - Lentes/filtros diferidos: assert disabled + `title` (BR-8).
- **Gate:** verify + unit + storybook en foreground.

### T5 · Widget Drawer READ + confirm Desvincular (G1/G2/G4/G5/G6/G7/G8 en la superficie 3)
- **Archivos:** `widgets/portafolio/ui/portafolio-drawer.tsx` (§2.6) + `portafolio-drawer.stories.tsx`;
  css en `portafolio.css`.
- **Stories obligatorias:**
  - `IdentidadResuelta` — home visible; facetas; zona canónico AUSENTE con CTA disabled+tooltip S2;
    acciones staged Reparar/Backport disabled+`title` «próximo · S5» (S1-D11); línea frescura (S1-D15).
  - `IdentidadProvisional` — «identidad provisional (sin home)» + scope visible (G6); facet
    marketplaces «desconocido»; empresa «desconocida» (G4).
  - `InstalacionEnDerivaReal` — chip `en-deriva` + detalle en `title` (G1); tipo `referenciada-cc`;
    origen de la copia: registry + `v0.5.2` (fixture real).
  - `UpdateNoVerificado` — el texto exacto «update: no-verificado» + tooltip S4 (G2); assert de
    AUSENCIA de cualquier `⬆ v`.
  - `ConDiscrepancias` — discrepancias del origen VISIBLES como texto (C-OR-6).
  - `TrazabilidadEslabones` — `<details>` se expande y lista fuente·campo·valor crudos (BR-3, S1-D14).
  - `ConCanonico` — fixture con canónico: path+version, «estado del checkout: no evaluado en este
    slice», `Abrir en Mapa` habilitado (dispara `onObservar(canonico.path)`).
  - `SinInstalaciones` — «sin instalaciones registradas» (G5).
  - `ObservarSinSesion` — `onObservar` undefined → botón disabled + `title` con
    `observarDisabledMotivo` (S1-D13).
  - `ConfirmarDesvincular` — click Desvincular → dialog: copy exacto «solo lo saca del portafolio…»,
    checkbox borrar-clon **disabled**+tooltip S2 (G7), Cancelar cierra sin llamar, Confirmar llama
    `onDesvincular`; `desvincularError` visible cuando se pasa.
  - `FocoYTeclado` — asserts a11y (G8): `role="dialog"` + `aria-modal` + `aria-labelledby`; el foco
    inicial vive dentro del drawer; Tab desde el último focusable vuelve al primero (trap); Escape
    llama `onClose`.
- **Gate:** verify + unit + storybook en foreground.

### T6 · Widget Wizard rama Proyecto (G3/G5/G8 en la superficie 2)
- **Archivos:** `widgets/portafolio/ui/portafolio-wizard.tsx` (§2.6) + `portafolio-wizard.stories.tsx`;
  css en `portafolio.css`.
- **Stories obligatorias:**
  - `Paso1Fuente` — radio Carpeta local activo con input path + botón Escanear; radio GitHub
    **disabled** + `title` «próximo · S2»; tab Marketplace **disabled** + `title` (G3 — assert que NO
    existe ningún «✓ marketplace válido» en el DOM); botón «Elegir carpeta…» presente solo si
    `onElegirCarpeta` viene (S1-D9).
  - `Escaneando` — spinner + botón Cancelar habilitado (dispara `onCancelarEscaneo`); input bloqueado.
  - `CandidatosReales` — fixtures de candidatos (§2.8): checkbox por candidato; `v?` cuando no hay
    versión; «origen desconocido» cuando no hay registry; chip deriva honesto (G1); badge «ya en el
    portafolio» cuando `clave ∈ clavesExistentes` con checkbox HABILITADO (S1-D10); candidato
    `es_canonico` con copy de canónico (RN-IDENT-4); contador del botón «Agregar N al portafolio»
    sigue a los checkboxes; nota «espejos read-only».
  - `SinHallazgos` — copy C-P-4 «No encontré arneses instalados aquí» + elegir otra carpeta (G5).
  - `ErrorDePath` — el motivo 400 REAL del backend, textual (C-P-3/G5).
  - `Agregando` — botón bloqueado «Agregando…»; sin checklist inventada.
  - `A11yModal` — `role="dialog"` + `aria-modal`; tabs `role="tab"`/`aria-selected`; Esc llama
    `onClose`; foco atrapado (G8).
- **Gate:** verify + unit + storybook en foreground.

### T7 · Página composition-root + wiring real + Abrir en Mapa (S1-D6/D13 + G5 vivos)
- **Archivos:** `pages/shell/ui/portafolio-view.tsx` (§2.7 — transporte completo: GET al montar,
  refetch tras agregar/desvincular, AbortController del escaneo, confirmación de colisión S1-D2,
  `onElegirCarpeta` Tauri) · `pages/shell/ui/global-view.tsx` (swap ruta) ·
  `shared/store/app-store.ts` (`mapaPeek`) · `pages/shell/ui/workspace-stage.tsx` (consumo del peek,
  §2.7) · `shared/api/types.ts` solo si necesitás exportar algo transversal (NO muevas los tipos del
  portafolio ahí — viven en la entity).
- **Verificación en este ticket** (la página no se storiea — transporte): `pnpm run verify` verde +
  smoke manual contra el daemon real (`go run ./cmd/arnesia serve` + `pnpm dev`): la ruta
  `#/portafolio` lista lo real, el wizard escanea un proyecto real, Abrir en Mapa navega y renderiza.
  El E2E exhaustivo y con evidencia es T8 — acá alcanza el smoke para commitear honesto.
- **Cuidado:** NO toques `viewedId`/guards del stage más allá del efecto del peek; el chip de alcance
  (RF-111/118) y el picker existentes deben seguir funcionando (regresión = correr las stories del
  map-canvas en el gate).

### T8 · As-code + mockup G1-G9 + E2E vivo + PARIDAD + cierre
1. **Capabilities:** completar las 4 `fe-portafolio/*` (status según R4; `valida:` = stories/tests
   REALES de T3-T7; scenarios BDD citando G-fixes y C-XX) · `change_log: extend` en
   `fe-shell/navegacion-global.yaml` (CAP-75: portafolio ya no es ComingSoon) y en
   `portafolio/desvincular.yaml` (superficie FE + confirmación G7) · regenerar índice:
   `python3 scripts/cap_doctor.py --index`. Si `cap_doctor`/schema exige registrar el módulo nuevo
   `fe-portafolio`, registralo donde el validador lo pida (mirá cómo está `fe-mapa`).
2. **Mockup (S1-D12):** corregir `mockups/arnesia-portafolio.html` EN SITIO — fuera `drift:N`,
   `⬆ v1.2.0`, `empresa:"Vitalia"` inferida, «✓ marketplace válido» activo; entra `deriva`/`update:
   no-verificado`/keys por clave/«sin empresa». Actualizar la fila en `mockups/INDEX.md`: estado →
   «✅ construido (Slice 1) — SSoT = stories `portafolio-*` @<commit>; snapshot corregido».
3. **E2E vivo (máquina real, browser + curl — pegá las salidas REALES recortadas en `paridad.md`):**
   - `go run ./cmd/arnesia serve` + `cd web && pnpm dev` → `#/portafolio`.
   - Estado inicial honesto (vacío o lo que haya). Wizard → escanear `~/Proyectos/luana-vitalia`
     (proyecto real): candidatos reales (el `harness` 0.5.2 con registry y `en-deriva` REALES como en
     Slice 0), 0-hallazgos y error-de-path probados con paths reales inválidos.
   - Agregar 2 candidatos → Lista real (grupo «sin empresa» esperable) → Drawer del `harness`:
     deriva real, update no-verificado, trazabilidad de eslabones.
   - **Abrir en Mapa de la instalación `referenciada-cc`** (cache CC bajo `~/.claude/…`): el Mapa
     renderiza el grafo REAL del plugin harness; la tab fuente del inspector muestra el mensaje honesto
     de sin-registro (GAP-1 resuelto y demostrado). Verificá con `curl GET /api/arneses` que NO se
     registró ningún cwd nuevo.
   - Corromper `~/.arnesia/portafolio.json` a mano → banner de corrupta visible + el resto vivo +
     `serve` re-bootea (BR-11).
   - Desvincular con confirmación → DELETE real; segundo DELETE por curl → 404.
   - Teclado: recorrer Lista→fila→Enter→drawer→Esc sin mouse (G8 manual + el axe automático ya corrió).
   - **Restaurar `~/.arnesia/portafolio.json`** a su estado previo (el E2E no deja basura).
4. **`paridad.md`:** tabla spec §10 ↔ realidad + **tabla G1-G9 ↔ evidencia** (de §3.0, con los nombres
   REALES de stories/tests que quedaron) + S1-D1..D15 estado + desviaciones visibles. Firma 🧑‍⚖️
   PENDIENTE (gate humano del operador — no la simules).
5. **Cierre documental:** `INDEX.md` de este paquete («Retomar aquí» → gate PARIDAD pendiente) ·
   `docs/product/checkpoint.md` (paquete activo; cifras via `estado.sh`, NO tecleadas) · BACKLOG item 1
   → construido-pendiente-firma + deuda que siga viva (re-key índice S0-D6 sigue abierta; lentes
   proyecto/marketplace si alguien las quiere) · memoria de sesión si tu harness la tiene.

## M · GATE LOCAL (correr tras CADA ticket; TODO verde antes de commitear)

```bash
# Go (obligatorio en T1; en T2-T8 corre igual — es barato y caza regresiones de imports)
go build ./... && go test ./... -race
golangci-lint run
go run github.com/fe3dback/go-arch-lint@latest check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml
go test ./docs/architecture/fitness/...
bash scripts/estado.sh --check || bash scripts/estado.sh   # regenera cifras si driftaron

# FE (desde T2)
cd web && pnpm run verify                                  # typecheck+biome+depcruise+steiger+stylelint
cd web && pnpm exec vitest --project=unit run              # selectores (S1-D7)
cd web && pnpm exec vitest --project=storybook run         # ⚠ FOREGROUND: Chromium no corre en bg (HS-20)

# T1 y T8 además:
go run ./cmd/arnesia conformance --todo                    # pass ≥ 42 · fail 0 (sin regresión)
```

El hook lefthook `capabilities` exige trazabilidad por commit: T1 crea su capability completa; T2 crea
las 4 `fe-portafolio/*` como stub con pointers reales (los commits T3-T7 tocan esas hojas si suman
archivos nuevos al módulo); T8 las completa. Mismo esquema que Slice 0 (§M) — el SSoT crece con el código.

## P · Prohibiciones (anti-drift del ejecutor)

1. NO relitigar lo firmado (HS-22/HS-23) ni las S1-D de este paquete — si algo no cierra en código,
   documentá una S1-D nueva EN EL MISMO TURNO y tomá la alternativa más cercana al espíritu
   (honestidad > limpieza).
2. NO fabricar datos: sin dato = `v?` / «desconocido» / «no-verificado» / «sin señal» / disabled+tooltip
   (BR-8). Las fixtures se calcan de salidas reales del backend — jamás `drift:N`, `⬆ vX`, empresas
   inferidas del path, ni validaciones ✓ de ramas no construidas.
3. NO tocar Slices 2-5: nada de clonar, marketplace, publicar, update-check real, reparar, backport,
   auth GitHub, borrar-clon funcional. Solo sus botones disabled con tooltip del slice.
4. NO deps nuevas: `go.mod` intacto; `web/package.json` sin entradas nuevas (primitivos = los
   existentes + copy-in shadcn sobre `@base-ui-components/react` YA presente si falta Dialog/Checkbox).
5. NO re-keyear el índice ni tocar `ArnesRegistry`/`arneses.json`/`checkProtected`: observar NO
   registra cwd (S1-D1); la colisión bare-id se confirma en FE, no se «arregla» en backend (S1-D2).
6. Transporte SOLO en `pages/*` (fe-transporte-independiente): widgets y entities reciben props puras;
   `shared/api` sigue domain-free (genéricos `<T>`, cero imports de entities).
7. Estilos SOLO con tokens DTCG (stylelint lo rompe); light+dark; reduced-motion (G8). Nada de estilos
   inventados fuera del vocabulario del Storybook vigente.
8. NO pisar vocabulario L0: `origen` (estandar/del-puesto) · `procedencia` · `canal` · `insumos` ·
   `banda` no se reusan para conceptos nuevos; el del Portafolio se rotula «origen de la copia» (S1-D14).
9. NO teclear cifras (checkpoint/INDEX — las genera `estado.sh`); NO teclear `status:` de caps contra
   R4; NO firmar PARIDAD ni gates humanos (eso es del operador).
10. Los tests NO tocan la máquina real (fixtures/inyección); SOLO el E2E de T7-smoke/T8 usa el daemon
    real, y lo único que escribe es `~/.arnesia/portafolio.json` — restaurálo al cerrar.
11. `vitest --project=storybook` SIEMPRE en foreground (Chromium no headless en bg — gotcha HS-20);
    ejecutá `pnpm run verify` antes de cada commit FE.
12. La ley anti-drift en la UI: NINGÚN botón «Editar» sobre una instalación, nunca (INV-1/C-BK-3) — ni
    siquiera disabled: ese camino NO existe.
