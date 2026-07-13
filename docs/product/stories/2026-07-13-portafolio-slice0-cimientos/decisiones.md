# Decisiones — Portafolio · Slice 0 «Cimientos» (arquitectura del build)

> `tipo: decisiones` · paquete `2026-07-13-portafolio-slice0-cimientos` · deriva del modelo FIRMADO HS-22
> (`stories/2026-07-10-spike-carga-arneses/`). Decisiones TÉCNICAS del arquitecto (Fable 5, 2026-07-13)
> que cierran los huecos que S-D11 difirió al build. El ejecutor (Sonnet 5) NO las relitiga; si una
> resulta inviable en código, documenta el porqué AQUÍ (nueva S0-D) y elige la alternativa más cercana.

## S0-D1 · Eslabón CC de procedencia (F5) — RESUELTO con evidencia de máquina real

Investigado 2026-07-13 sobre la instalación CC local (lo que S-D11 exigía ANTES de firmar el orden fino):

- **`~/.claude/plugins/known_marketplaces.json`** — mapa `nombre-marketplace → {source:{source:"github",repo:"owner/repo"}, installLocation, lastUpdated}`. `installLocation` = checkout LOCAL del marketplace (`~/.claude/plugins/marketplaces/<nombre>/`).
- **`~/.claude/plugins/installed_plugins.json`** (`"version": 2`) — mapa `"<id>@<marketplace>" → [ {scope:"project", projectPath, installPath, version, installedAt, lastUpdated} ]`. `installPath` apunta al **cache global** `~/.claude/plugins/cache/<marketplace>/<id>/<versión>/` — un dir **forma-plugin válido** (tiene `.claude-plugin/plugin.json` con `name`/`version`/`description`/`displayName`), cargable con `LoadArnes`. `version` puede ser `"unknown"` (→ tratar como `?`).
- **`.claude/settings.json#enabledPlugins`** del proyecto — `"<id>@<marketplace>": true|false`. Es el lado-proyecto del registro CC.
- El cache CC mete archivos meta `.in_use` / `.orphaned_at` dentro del dir del plugin (excluir del hash de deriva).

**Realidad medida (10 proyectos reales):** NINGÚN proyecto tiene `.claude/plugins/<id>/` físico hoy; NINGÚN `.devstudio/arneses.yaml` existe aún. La instalación CC es **referenciada** (bytes en cache global + puntero en settings), no materializada dentro del proyecto.

**Orden fino de la cadena (cierra F5; ratifica el operador en el gate PARIDAD):** para *procedencia de la copia* los **records de install-time mandan**, cada uno autoritativo en su carril de gestión: lock DevStudio (carril DevStudio) y CC `installed_plugins`+`enabledPlugins` (carril CC). Después: remote git del dir del plugin (eslabón 3) → manifiesto (que solo declara `home` autor, JAMÁS procedencia de copia). Conflicto entre carriles/eslabones → discrepancia VISIBLE (C-OR-6), nunca elección silenciosa. Parse defensivo de los JSON de CC: `version` ≠ 2 o shape inesperada → eslabón `no-legible` visible; jamás crash (formato interno de CC, puede driftar).

## S0-D2 · Tres tipos de instalación (sale de S0-D1, refina el walker D-DOM-3)

`domain.TipoInstalacion`: **`materializada`** (dir `raíz/.claude/plugins/<id>/`, convención DevStudio HS-12 — hoy 0 casos reales, se construye igual: contrato firmado) · **`referenciada-cc`** (`enabledPlugins:true` → metadata CC → dir del cache global) · **`proyecto-instalado`** (el `.claude/` poblado del proyecto mismo = forma-instalada, nomenclatura §1). El walker enumera las tres. `enabledPlugins:false` NO se lista (no está activa en el proyecto).

## S0-D3 · D-DOM-2 aterrizado (empresas[] sí; `marketplace` NO se renombra en el manifiesto)

- `domain.Arnes.Empresa string` → **`Empresas []string`** con tag `json:"empresas"`. Unmarshal TOLERANTE (custom): acepta `empresa` escalar legacy y lo normaliza a `[valor]`; Marshal emite SOLO `empresas`. Schema `graph.l0`: `required` pasa de `empresa` a `anyOf(empresa | empresas)`; se agrega `version` (string, opcional).
- `domain.Arnes.Marketplace` **se queda** (key JSON `marketplace` = **home autor-declarado, CRUDO**, tal como lo tecleó el autor). El **`Home` canonicalizado** y las **`Registries []string`** (facetas de adquisición N:M) viven en `domain.EntradaPortafolio` — la procedencia de la copia jamás viaja en el manifiesto (viaja en los bytes, sería mentira). Esto satisface la INTENCIÓN de D-DOM-2 (N:M representable + home como identidad) sin romper todos los `arnes.l0.json` existentes.
- Migran a `empresas`: dogfood (`dev-full-cycle`, `content-studio-full`), kit, seed `index/store.go` (el hardcode `Empresa:"alpacapurpura"` → `Empresas:[]string{"alpacapurpura"}` — es seed demo, honesto como demo), FE `types.ts` + ~10 usos.

## S0-D4 · Layout de paquetes (hexagonal, sin dependencia adapter→adapter)

- **`internal/domain/portafolio.go` + `repo_ref.go`** — tipos + canonicalización + reconciliación PURA (cero I/O).
- **`internal/adapters/portafolio/`** — paquete adapter NUEVO: `store.go` (persistencia) + `scanner.go` (walker físico + lock + metadata CC + parser `.git/config`) + `deriva.go` (hasher + referencias locales). Depende SOLO de `domain`+`ports` (+`yaml` para el lock).
- El scanner NO llama `loader.LoadArnes` (adapters no se importan entre sí): devuelve **hallazgos** (dir + tipo + eslabones crudos); el **usecase `PortafolioService`** orquesta cargar/resolver/persistir vía puertos (`ports.ArnesLoader` nuevo, satisfecho por el loader; cablea `cmd`, como ya hace `loadArnesDir`).
- go-arch-lint: componente nuevo `portafolio: in: internal/adapters/portafolio/**`, `mayDependOn: [domain, ports]`, `canUse: [yaml]`; `cmd` lo suma a su lista.

## S0-D5 · Store del Portafolio: archivo y forma

`~/.arnesia/portafolio.json` (separado de `arneses.json` — A1). Envelope versionado `{"version":1,"entradas":[…]}`. Carga **entry-wise** (`[]json.RawMessage`, decode por entrada): entrada corrupta → se conserva cruda + visible como corrupta en el listado, se SALTA, jamás impide boot (BR-11/C-N-4 — el anti-patrón es `arneses.json` que hoy brickea `main.go`). Escritura atómica temp+rename (mismo patrón `arnes_registry.go#saveLocked`). Merge por identidad: `Upsert` dedup instalaciones por `install_path` canónico (C-N-3: EvalSymlinks + Clean antes de comparar).

## S0-D6 · Índice del Mapa NO se re-keyea en Slice 0 (A2 acotada — deuda explícita)

El índice in-memory (`index.Store`, key = id bare) es la superficie de SESIÓN/Mapa, no el Portafolio. El Portafolio keyea calificado `(home,id,scope)` en SU store (A1 ya separó). El re-key global del índice llega cuando «Abrir en Mapa» de una instalación lo exija (Slice 1) → registrar como `deuda` en BACKLOG al cerrar. Lo que SÍ entra ya: el store del portafolio JAMÁS fusiona provisional↔resuelto (C-ID-2).

## S0-D7 · Deriva en Slice 0 = solo referencias LOCALES (sin red, sin gh/PAT)

Referencia = `home/plugins/<id>/<versiónInstalada>/` buscada en: (1) checkout local del marketplace que
`known_marketplaces.json` conoce para ese `home` (su `installLocation`); (2) un registry local si la cadena
resolvió un path git local. Sin referencia accesible → `deriva-no-evaluable` (default honesto). Hash: sha256
por archivo sobre el árbol ordenado (path relativo con `/` + contenido), excluye `.git/`, `.in_use`,
`.orphaned_at`. NUNCA semver-string ni `git status` como veredicto (BR-4). Auth GitHub queda FUERA de Slice 0.

## S0-D8 · Sin dependencias nuevas

`.git/config` se parsea a mano (subset INI: secciones `[remote "x"]` + `url`) + resolución de `gitdir:` para
worktrees (C-N-1). NO se agrega go-git (el módulo tiene 2 deps y así se queda). Canonicalización acepta
`owner/repo` (asume `github.com`), `https://…`, `git@host:…`, `ssh://…`; minúsculas todo; sin `.git`/trailing.

## S0-D9 · Superficie observable de Slice 0: HTTP mínima + CLI (sin FE)

- HTTP: `GET /api/portafolio` (entradas + corruptas visibles) · `POST /api/portafolio/escaneos {path}` (escanea, NO persiste — devuelve candidatos) · `POST /api/portafolio/proyectos {path, elegidos[]}` (persiste elegidos; C-P-8 re-escaneo idempotente) · `DELETE /api/portafolio/arneses/{clave}` (desvincular; NO borra nada del disco — C-UNL-3). `clave` = slug estable derivado de la identidad (expuesto en el listado). openapi.yaml se actualiza.
- CLI: `arnesia portafolio <escanear|listar|agregar|desvincular>` — la vía de verificación E2E sin FE (y del operador). Reusa el MISMO usecase (cero lógica en cmd).
- Slice 1 (FE) consumirá exactamente esto; defaults honestos ya salen del backend.

## S0-D10 · E3 (ubicación del canónico) — cerrado en papel, se construye en Slice 2

Default `~/Arneses/<home-slug>/<id>/` (visible, editable, FUERA de toda ruta protegida por `checkProtected` →
registrable para «Abrir en Mapa»), override por flag `-checkouts-root` cuando Slice 2 lo construya. Slice 0 solo
implementa: campo `Canonico{Path,Version}` en el store + clasificación RN-IDENT-4 (path detectado ∈ checkouts
conocidos del store → CANÓNICO, jamás instalación) + test.

## S0-D11 · As-code: boundary nuevo + capabilities

- Boundary nuevo `docs/architecture/boundaries/portafolio-identidad-y-deriva-honesta.md` (🌱→enforced en el
  mismo slice): 4 checks — `identidad-calificada-unica` (BR-1/C-ID-2) · `deriva-nunca-semver` (BR-4) ·
  `store-degrada-honesto` (BR-11) · `procedencia-anotada` (BR-3), cada uno `enforced_by:` su test Go real.
- 5 capabilities nuevas módulo `portafolio` (las que firma spec-funcional §9): `registrar-identidad` ·
  `escanear-proyecto` · `resolver-origen` · `evaluar-deriva` · `desvincular`; + `change_log: extend` en
  CAP-15/16/17 (loader). `portafolio` se suma a `domain_modules` en `project.config.yaml` (seam).

## S0-D12 · Desviaciones de ejecución (Sonnet 5, T1) — anti-drift documentado

- **Colisión de nombre `Origen`:** el plan §2.1 nombra el tipo de reconciliación collect-all
  `domain.Origen`, pero ese identificador YA existe (`internal/domain/box.go:177`, `nodo.origen` L0:
  estandar/del-puesto). Alternativa más cercana al espíritu: `domain.OrigenPortafolio` (mismo shape,
  mismos campos `Registry/Version/Eslabones/Discrepancias`; el campo `Instalacion.Origen` lo tipa).
- **Fix de compilación forzado a T1 (fuera del file-list de T2):** `domain.Arnes.Empresa string` →
  `Empresas []string` rompe la compilación de dos sitios Go que el plan asignaba a T2/no mencionaba:
  `internal/adapters/index/store.go:102` (seed, SÍ estaba en T2 — se adelantó) e
  `internal/adapters/transport/http/router.go` `harnessSummary.Empresa` (NO estaba en el file-list de
  ningún ticket). Se cambió `harnessSummary.Empresa string` → `Empresas []string` (json
  `empresas,omitempty`) para mantener `go build ./...` verde en T1 — el tipo FE `HarnessSummary` en
  `web/src/shared/api/types.ts` migra recién en T2 junto con el resto de los tipos FE (hoy compila
  igual porque TS estructural no exige el campo).
