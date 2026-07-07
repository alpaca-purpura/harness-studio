# FASE 1 — revisión + diseño técnico (sellado antes de codear)

> 2026-07-07 · producto de la fase de revisión del PROMPT-implementacion.md.
> Contrasta RF-80..96 contra el código existente y fija DÓNDE vive cada pieza,
> cumpliendo arch/ (canvas⊥chrome · fe-transporte-independiente · fe-topologia-fsd ·
> superficie-local-confinada). Ninguna línea de código antes de este sello.

## Revisión RF ↔ existente (qué hay, qué falta)

| RF | Existe hoy | Falta |
|---|---|---|
| 80-83 shell | `inspector.tsx` = panel fijo 340px, un solo cuerpo, ✕ cierra | ⤢/⤡ + modo expandido + tabs + CSS expandido (sticky/rejilla/920) |
| 84 vacío | la página NO monta Inspector sin selección (ausencia) | estado vacío con affordance |
| 85-88 tooltips | nada (títulos/campos planos) | diccionario doctrinal + «i» + punteado + a11y |
| 89 viene-de | `selectors.ts` solo selectores de geografía | `selectVieneDe` (edges inversos) + sección |
| 90 chips | necesita/ruta = `<ul>` texto plano | chips tipados navegables vía `onSelect` + badge cond |
| 91 hallazgos | endpoint `GET …/conformance` vivo (HS-11); FE no lo consume | sección Hallazgos + filtro por nodo + client |
| 92 botonera | nada | footer staged disabled rotulado |
| 93-95 contenido | `fuente_path` viaja en L0; NO hay endpoint de lectura | endpoint Go confinado + tab viewer + acciones staged |
| 96 corridas | nada | tab honesta estática |

## Decisiones de ubicación (arquitectura)

1. **Diccionario doctrinal → `web/src/entities/arnes/model/doctrina.ts`** (nuevo).
   `SEC_TIP` · `DEF_CAMPO` · `DEF_VALOR` · `PROP_NOTE` · `tipDe(k,v)` — copia cementada
   del mockup v6 (mockup:309-360), fuentes doctrinales citadas en comentario. Cero
   imports (dato puro) → legal en entities/model (FSD, depcruise). La entity es la dueña
   (RF-87); el widget solo consume vía `@/entities/arnes`.
2. **Edges inversos → `selectors.ts`**: `selectVieneDe(g, nodeId): {de, tipo}[]` sobre
   `graph.edges` ya cargados (RF-89). Puro, sin transporte.
3. **Conformance por nodo → FILTRO EN FE** (decisión pedida por el prompt, justificada):
   los checks del scope `arnes` son de GRAFO con violaciones listadas en `detalle`
   («id (motivo); id2 (…)», `domain.veredictoDeLista`) — la atribución por nodo NO existe
   como dato en el reporte; un query param en el daemon haría el MISMO string-match en
   peor lugar (transporte con regla de filtrado) y rompería el cacheo 1-fetch-por-arnés.
   FE: tipo espejo mínimo `ConformanceResult` en `entities/arnes/model/types.ts`
   (transport-free) + `selectHallazgosConformance(results, nodeId)` en selectors.ts —
   filtra `veredicto ∈ {fail,error}` y matchea el id del nodo por word-boundary en
   `detalle`. Si mañana el motor atribuye nodo por dato, el selector se simplifica.
4. **Transporte queda en la PÁGINA** (patrón vigente de workspace-stage:
   «transport lives here, not in the entity/canvas»): la página fetchea
   `getConformance` al cargar el grafo y le INYECTA al Inspector `conformance` (dato) y
   `loadFuente(nodeId)` (callback sobre `api.getNodeFuente`). El Inspector sigue sin
   importar `shared/api` (hexagonal FE; stories triviales con fakes). `client.ts` gana
   `getConformance<T>(id)` y `getNodeFuente(id, nodeId)` (texto plano → helper `reqText`).
5. **Endpoint fuente → hexagonal completo, confinado S2 (RF-93):**
   - Puerto outbound nuevo `ports.FuenteReader { LeerConfinado(dir, ref) ([]byte, error) }`.
   - Adapter `internal/adapters/artifact/fuente.go` (la familia de lectores confinados ya
     vive ahí: mismo patrón `confinedPath` de artifact.Reader). Acepta ref relativa
     (join bajo dir) o absoluta (debe caer DENTRO de dir); Clean+Rel valida traversal.
   - Usecase `internal/usecase/fuente_service.go`: `FuenteService{index, workdir, files}`.
     Flujo: grafo→nodo→`fuente_path` (vacío → `ErrFuenteNoDeclarada`, 404) → 
     `WorkdirResolver.Resolve` (solo `registered=true`; si no → `ErrArnesSinDirectorio`,
     404 honesto: «Carga la carpeta») → lectura confinada.
   - Ruta `GET /api/harnesses/{id}/nodes/{nodeId}/fuente` en router.go → 200 `text/plain`
     · 404 (arnés/nodo/fuente/dir/archivo ausente) · 403 (ref escapa del dir).
   - **Changelog OpenAPI**: path nuevo en `arch/contracts/api/openapi.yaml` + bump
     `info.version` → `0.2.0-hs09`.
   - Composition root: `NewHandler` gana el servicio (firma cambia; main.go lo cablea).
   - El showcase (`content-studio-full`, embebido, sin disco) responde 404 honesto — la
     tab lo DICE, jamás inventa (RF-93 Gherkin «nodo sin fuente»).
6. **Shell del drawer → reescritura de `inspector.tsx` + `web/src/app/styles/inspector.css`**
   (patrón map.css: port VERBATIM del CSS del mockup scoped bajo `.arnesia-inspector`,
   tokens `var(--…)`, importado desde `app/styles/index.css`). Estado `expanded` + tab
   activa = estado local del Inspector (ninguna necesidad de estado global); `Esc`
   colapsa solo en expandido; ✕ = `onClose` SIEMPRE; tab se resetea a Resumen al cambiar
   de nodo. La página monta el Inspector SIEMPRE en vista Mapa (RF-84).
   - **Interpretación registrada (RF-81 ↔ RF-84):** «✕ cierra del todo» = desaparecen el
     overlay expandido Y el drawer del nodo; como «sin selección → affordance, no
     ausencia» (RF-84) también está firmado, tras ✕ queda el drawer VACÍO (affordance).
     Es la única lectura que mantiene vivos ambos RF. Se consulta en paridad.
7. **Stories = test por marca** (vitest --project=storybook, convención
   inspector.stories.tsx): expandir/colapsar/cerrar · tabs · tooltips (sección+campo) ·
   viene-de + chip navegable + cond + chip inerte · hallazgo gate:none · sin-hallazgos ·
   botonera disabled · contenido con fuente (fake) · sin fuente honesto · no-reconocido
   raw · corridas honesta (+nota caja) · vacío.

## Mini-plan de commits (orden de FASE 2)

1. `feat(HS-09): drawer — diccionario doctrinal + selectores viene-de/hallazgos (entities/arnes)`
   → doctrina.ts · selectVieneDe · ConformanceResult + selectHallazgosConformance · exports.
2. `feat(HS-09): drawer shell — tabs, ⤢/⤡/✕, expandido y estado vacío (RF-80..84)`
   → inspector.tsx (shell) · inspector.css · página (drawer persistente) · stories shell.
3. `feat(HS-09): drawer — tooltips doctrinales por sección y campo (RF-85..88)`
   → Section/Field con «i»/punteado + a11y · stories tooltips.
4. `feat(HS-09): drawer — Resumen: viene-de, chips navegables, hallazgos, botonera staged (RF-89..92)`
   → panes del Resumen · client.getConformance · página · stories.
5. `feat(HS-09): fuente confinada — GET /api/harnesses/{id}/nodes/{nodeId}/fuente (RF-93)`
   → ports.FuenteReader · adapters/artifact/fuente.go · usecase/fuente_service.go ·
   router.go · main.go · openapi.yaml (changelog) · tests Go (usecase+adapter, -race).
6. `feat(HS-09): drawer — tab Contenido real + tab Corridas honesta (RF-93..96)`
   → client.getNodeFuente/reqText · página loadFuente · panes Contenido/Corridas · stories.
7. `docs(HS-09): PARIDAD verificada + retomar-aquí` (tras FASE 3 con Chrome DevTools).

Gates que deben quedar verdes en cada commit y al cierre: `tsc --noEmit` · `biome ci src`
· `depcruise` · `steiger` · `stylelint` · story-tests · `go build/vet/test -race ./...` ·
`golangci-lint` · `go-arch-lint`.
