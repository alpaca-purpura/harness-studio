# Spec funcional · Agregación consolidada al Portafolio (proyecto + marketplace)

> `tipo: spec` · paquete `2026-07-23-portafolio-agregar-marketplace` · 2026-07-25.
> Etapa 3 del flujo (METODOLOGIA §10). Habilitado por **Gate 1 🧑‍⚖️** (mockup firmado).
> Insumos vinculantes: [`decisiones.md`](./decisiones.md) AG-D1..D12 + PENDIENTE-01 ·
> [`auditoria-storybook-vs-app.md`](./auditoria-storybook-vs-app.md) ·
> [`mockup-agregacion-consolidada.html`](./mockup-agregacion-consolidada.html) (6 superficies firmadas).
> El diseño técnico (contratos Go, endpoints, capabilities, boundaries, plan de pruebas) va en
> [`design.md`](./design.md).

## 0 · Alcance y no-alcance

**Dentro:**

| # | Superficie | Estado hoy |
|---|---|---|
| S1 | Plano **Arneses** — toolbar con rótulos visibles | existe, se corrige (L1) |
| S2 | Plano **Marketplaces** — lista de marketplaces conocidos | nuevo |
| S3 | **Catálogo** de un marketplace propio — 5 situaciones × 5 acciones | nuevo |
| S4 | **Catálogo** de un marketplace de referencia — read-only | nuevo |
| S5 | **Wizard** rama Proyecto — máquina de estados + Atrás/Cancelar | existe, se corrige (W3/W4) |
| S6 | **Wizard** rama Marketplace — registrar | existe disabled, se construye |
| S7 | **Diálogo de reconciliación** — asignar origen | nuevo |
| S8 | **`↧ Traer canónico`** — materializar el arnés del estante para trabajarlo (AG-D17) | botón existe `disabled` en el drawer; se construye |

**Fuera (explícito):** ejecutar Publicar / Actualizar-mi-copia / Reparar. En este paquete esas
acciones se **pintan con su situación real y quedan `disabled` + tooltip** apuntando a su ítem del
outcome (3/4/5). Motivo: cada una es su propio paquete con su propio gate. Lo que este paquete SÍ
entrega es el **cálculo honesto de la situación** que las habilita.

**Fuera:** clonar por `git` un repo de proyecto (rama «Repositorio GitHub» del wizard sigue `disabled`
+ `S2`, sin cambio).

## 1 · Tokens (contrato `fe-tokens-contrato`)

Cero color literal. Todo `var(--…)` de `web/src/app/styles/theme.css` (generado de
`web/tokens/base.tokens.json`). Tokens usados por este paquete — **ninguno nuevo**:

| Uso | Token |
|---|---|
| fondo de superficie / fila | `--card`, `--background` |
| fondo secundario (chip, segmento, panel) | `--secondary`, `--popover` |
| texto / texto atenuado | `--foreground`, `--muted-foreground` |
| acento de marca, activo, foco | `--primary`, `--accent-soft`, `--ring` |
| bordes / borde de input | `--border`, `--input` |
| salud: al-hilo / atención / crítico | `--ok` + `--ok-soft` · `--warn` + `--warn-soft` · `--crit` + `--crit-soft` |
| radios / espaciado / tipografía | `--radius-{sm,md,lg,full}` · `--space-*` · `--text-{xs,sm,base,lg,xl}` · `--font-{display,sans,mono}` |
| sombra de modal | `--shadow-md`, `--shadow-lg` |

**Reglas duras de token en este paquete:**

- **`sin-senal` nunca reusa el color de `ok`.** «No sé» ≠ «sano» (regla ya enforced en
  `.pf-dot-salud.sin-senal`: transparente + borde dashed `--muted-foreground`). Aplica igual a los
  nuevos estados de marketplace (`no leído aún`, `sin acceso`).
- **`--warn` para «necesita acción», `--crit` reservado a fallo.** Un marketplace sin acceso es
  `--warn` (recuperable: autenticar), no `--crit`.
- **Deuda conocida heredada:** `.text-warn` no cumple contraste 4.5:1 (`#c96a2e` sobre blanco, axe
  `color-contrast`, en `BACKLOG.md`). Este paquete **no la arregla** pero **no la agrava**: los
  textos nuevos sobre `--warn-soft` usan `--foreground`, no `--warn`.

## 2 · Átomos y componentes (contrato `fe-taxonomia-componentes` v1.1)

La taxonomía enforça **dirección de import**, no un ladder atómico. Ubicación por capa:

### 2.1 Reusar tal cual (cero cambio)

| Componente | Path | Por qué |
|---|---|---|
| `DerivaChip` | `entities/portafolio/ui/chips.tsx` | 3 estados `al-hilo`/`en-deriva`/`deriva-no-evaluable` ya firmados |
| `TipoInstalacionChip` | idem | descriptivo, sin tono de salud |
| `AvisoChip` | idem | tono `warn`, texto completo nunca truncado (BR-8) |
| `DotSaludPortafolio` | idem | 3 ramas, `sin-senal` con clase propia (G9/S1-D4) |
| `PortafolioDrawer` | `widgets/portafolio/ui/portafolio-drawer.tsx` | el botón `↧ Traer canónico` ya existe ahí, hoy `disabled` — se **habilita**, no se duplica |
| `trapTabKeyDown` | `shared/lib/focus-trap.ts` | a11y de modal, patrón S1-D18 (sin portal, por `within(canvasElement)`) |

### 2.2 Entidad nueva — `entities/marketplace/`

Capa: **entities** (organismo de dominio). No importa canvas, no importa otra feature, no importa
transporte (`fe-transporte-independiente`).

```
web/src/entities/marketplace/
  index.ts                      barrel público
  model/types.ts                MarketplaceConocido · ClaseMarketplace · EstadoLectura
                                EntradaCatalogo · SituacionCatalogo
  model/selectors.ts            situacionDe() · agrupar/ordenar · pendientesDe()
  model/selectors.test.ts       tabla de casos (colocado, patrón sancionado)
  ui/chips.tsx                  ClaseChip · EstadoLecturaChip · SituacionChip
  ui/chips.stories.tsx          una story por estado, con play()
  testing/marketplaces.ts       fixtures con shape REAL (AG-D9/D10)
```

**Por qué entidad propia y no dentro de `entities/portafolio`:** un marketplace no es un arnés; tiene
identidad (`nombre`), ciclo de lectura y clase propios. `entities/portafolio` seguiría creciendo con
vocabulario ajeno. Los dos se cruzan solo en selectores de página.

### 2.3 Widgets nuevos

| Widget | Path | Rol |
|---|---|---|
| `MarketplaceList` | `widgets/marketplace/ui/marketplace-list.tsx` | S2 — filas de marketplaces conocidos + puerta cruzada |
| `MarketplaceCatalogo` | `widgets/marketplace/ui/marketplace-catalogo.tsx` | S3/S4 — catálogo, buscador+filtro+lazy, acción por situación |
| `ResolverOrigenDialog` | `widgets/portafolio/ui/resolver-origen-dialog.tsx` | S7 — vive en `portafolio` porque actúa **sobre un arnés** |

### 2.4 Compositions (moléculas) a extraer en `shared/ui/`

Estas nacen de duplicación **ya existente** más la nueva — extraerlas es parte del alcance:

| Molécula | Path | Reemplaza |
|---|---|---|
| `GrupoControl` | `shared/ui/grupo-control.tsx` | rótulo visible + grupo de botones (AG-D4). Lo usan toolbar de Arneses y de Catálogo |
| `ListaLazy` | `shared/ui/lista-lazy.tsx` | lista con centinela `IntersectionObserver` (AG-D3). La usan catálogo y candidatos |
| `BuscadorFiltro` | `shared/ui/buscador-filtro.tsx` | input search + N `FiltroDisclosure`. Idem |

`FiltroDisclosure` hoy vive privado dentro de `portafolio-list.tsx`: **se promueve** a
`shared/ui/filtro-disclosure.tsx` sin cambiar su comportamiento (mismo patrón sin-portal,
`aria-expanded` en el botón). Es refactor de movimiento, no de conducta.

### 2.5 Página

`pages/shell/ui/portafolio-view.tsx` gana el conmutador de **planos** (`Arneses` | `Marketplaces`)
y es **la única** que hace transporte (`fe-transporte-independiente`: los widgets reciben props
puras y callbacks; cero `fetch` dentro de un widget).

## 3 · Modelo de datos (FE)

```ts
type ClaseMarketplace = "propio" | "referencia"

// AG-D9 — de dónde salió el conocimiento de este marketplace. Collect-all, no exclusivo.
type EslabonMarketplace = "cc-known-marketplaces" | "declarado-por-operador"

type EstadoLectura =
  | { tipo: "leido"; cuando: string; entradas: number }
  | { tipo: "no-leido" }                                  // registrado, nunca leído
  | { tipo: "sin-acceso"; motivo: string }                 // gh no auth / 404 / privado
  | { tipo: "url-no-resuelve"; motivo: string }

interface MarketplaceConocido {
  nombre: string                    // clave de merge (AG-D9)
  repo: string                      // "owner/repo"
  clase: ClaseMarketplace
  eslabones: EslabonMarketplace[]   // ≥1; los dos = detectado + confirmado
  install_location?: string         // de CC; es la ref de `deriva`
  lectura: EstadoLectura
  discrepancias?: string[]          // repo distinto entre eslabones → se MUESTRAN, no se eligen
}

// AG-D10/D11 — la fila del catálogo es la ENTRADA DE ÍNDICE (canal instalable)
interface EntradaCatalogo {
  nombre: string                    // lo instalable: `nombre@marketplace`
  source: string                    // ruta relativa cruda, p.ej. "./plugins/harness/0.5.3"
  version?: string                  // DERIVADA de `source`; ausente si no se puede
  descripcion?: string
  comparte_source_con?: string[]    // otras entradas con el mismo `source` (canales)
  estado_canal?: "habilitada" | "deprecada"  // solo si hay catalogo.json (prenter)
}

// La situación cruza catálogo × portafolio. NUNCA se calcula en el widget.
type SituacionCatalogo =
  | { tipo: "no-lo-tengo" }
  | { tipo: "al-hilo" }
  | { tipo: "mi-copia-adelantada"; mia: string; estante: string }
  | { tipo: "estante-adelantado"; mia: string; estante: string }
  | { tipo: "instalaciones-en-deriva"; cuantas: number }
  | { tipo: "no-comparable"; motivo: string }   // ← 6ta rama, ver §4.3
```

## 4 · Lógica y flujo

### 4.1 Plano Marketplaces (S2)

1. Al entrar, la página pide la lista. El backend hace **collect-all** (AG-D9): lee
   `known_marketplaces.json` de CC **y** el registro propio; mergea por `nombre`.
2. Cada fila muestra: `repo` · `ClaseChip` · nº de entradas de catálogo · `EstadoLecturaChip`.
3. **`propio`** → fila clickeable al catálogo. **`referencia`** → clickeable a catálogo read-only.
4. **`sin-acceso` / `url-no-resuelve` / `no-leido`** → fila NO navega; muestra el motivo textual y
   un botón (`Reintentar` / `Leer catálogo`). **Jamás catálogo vacío** — ver BR-4.
5. Contador cruzado «N arneses sin origen resuelto» → navega al plano Arneses filtrado.

### 4.2 Lectura de catálogo (S3/S4)

**Cacheada, con refresco explícito** (AG-D8 decisión 3). La app es conductor, no proxy.

1. Fuente primaria: `<install_location>/.claude-plugin/marketplace.json` si CC ya lo tiene clonado
   (camino barato, cero red, funciona offline). Si no hay `install_location`, se lee del remoto.
2. Se parsea `plugins[]` → `EntradaCatalogo[]`. `version` se **deriva** del último segmento de
   `source` cuando parsea como semver; si no, queda `undefined` (nunca se inventa).
3. Se agrupan entradas con el mismo `source` → `comparte_source_con` (AG-D11).
4. **Enriquecimiento opcional:** si existe `catalogo.json` en la raíz, se toma `canales` y
   `versiones[].estado`. Si no existe, se degrada sin ruido — no es parte del estándar.
5. Se persiste el resultado con su timestamp. La fila del plano muestra `leído hace <t>`.

### 4.3 Cálculo de situación (S3) — la pieza que habilita los ítems 3/4/5

Para cada `EntradaCatalogo`, se cruza contra el Portafolio por identidad `(home, id)` donde
`home` = nombre del marketplace, `id` = `EntradaCatalogo.nombre`:

| condición | situación |
|---|---|
| no hay entrada en el portafolio con esa identidad | `no-lo-tengo` |
| hay canónico y `version(mía) > version(estante)` | `mi-copia-adelantada` |
| hay canónico y `version(mía) < version(estante)` | `estante-adelantado` |
| hay ≥1 instalación con `deriva = en-deriva` | `instalaciones-en-deriva` |
| hay entrada, versiones iguales, todas las instalaciones `al-hilo` | `al-hilo` |
| **falta un insumo para comparar** (sin `version` del estante, sin canónico, `deriva-no-evaluable`) | **`no-comparable` + motivo** |

La 6ta rama es obligatoria: sin ella el sistema tendría que elegir entre mentir (`al-hilo` por
defecto) o callarse. Se pinta con `DotSaludPortafolio.sin-senal` + motivo textual. Precedencia cuando
varias aplican: `instalaciones-en-deriva` > divergencia de versión > `al-hilo`; y `no-comparable`
gana sobre cualquier afirmación positiva.

> **Corregido 2026-07-25 (design.md §2 C7):** una versión anterior de este párrafo afirmaba que la
> 6ta rama «no está en el mockup». **Sí está** (línea 670, `sin-senal` + motivo), igual que el chip
> de canales de AG-D11 (línea 618) — el dibujo firmado ya respalda las dos, así que no hay superset
> que justificar acá. El texto había quedado de antes de actualizar el mockup.

> **Dónde se calcula (design.md §2 C1, normativo):** en el **dominio Go**, y viaja resuelta en el
> wire. NO en `entities/marketplace/model/selectors.ts` como decía §2.2 de este spec: cruzar catálogo
> × portafolio en el FE exige importar `EntradaPortafolio` desde `entities/marketplace`, que es
> cross-import entity↔entity y `steiger fsd/no-cross-imports` lo rompe (y `pnpm run fsd` está dentro
> de `verify`). El selector FE solo **presenta** (`etiquetaDeSituacion` / `tonoDeSituacion`).
> Coherente con la regla que este mismo spec ya daba: «NUNCA se calcula en el widget».

### 4.4 Wizard rama Proyecto (S5) — corrección de W3/W4

Máquina de estados **conservada** (AG-D6), con:

- `Atrás` en cada estado; **`disabled` visible** (no oculto) en el primer paso.
- `Cancelar` en todos: sale completo, **cero efectos** (S1-D9). Sigue **bloqueado** durante
  `agregando` (S1-D19: hay POST en vuelo).
- En `candidatos`: rótulo **«Paso 2 de 2 · escaneo — encontrados N arneses»** (W4) + la **ruta
  escaneada visible** con botón `cambiar ruta` que vuelve a `fuente` preservando el texto.
- **Nada premarcado** (AG-D7). Botón `Agregar 0…` nace `disabled`.
- Buscador + filtro + lazy sobre los hallazgos (AG-D3) — obligatorio *porque* no se premarca.

### 4.5 Wizard rama Marketplace (S6) — solo registrar

1. Input de git url + `Validar`. La validación **exige respuesta real**: se busca
   `.claude-plugin/marketplace.json` legible. Sin respuesta, **cero ✓** (criterio G3, el fix central
   del Slice 1 — no se relaja).
2. Éxito → se muestran `name` + `owner.name` + nº de `plugins[]` **leídos del archivo real**.
3. El operador elige **clase** (`propio` / `referencia`). Sin default silencioso que importe: arranca
   en `propio` porque es el caso dominante, y la etiqueta explica la diferencia a la vista.
4. `Registrar y ver catálogo` → persiste y **aterriza en S3/S4** (cierra el modal).
5. Si el nombre ya existe: **no se duplica ni se pisa**. Se informa que ya está y se ofrece ir a él.

### 4.6 Reconciliación (S7)

Acción **in-situ** sobre la fila del arnés (patrón «Identificar» del Slice 2). Opciones = marketplaces
conocidos ordenados por señales blandas (coincidencia de autor/nombre del `plugin.json`), más la
opción explícita **«ninguno — dejarlo sin origen»**. Nada premarcado; `Confirmar` nace `disabled`.
Confirmar escribe el `home` declarado; **no** clona ni instala nada.

### 4.7 `↧ Traer canónico` (S8) — AG-D17

Materializa el arnés del estante como **canónico editable** en nuestro propio territorio, para poder
mejorarlo. **No** lo instala para usarlo — eso lo hace Claude Code en el proyecto del cliente
(`vision.md` §Ecosistema: ArnesIA es dueña única del modificar; las apps de rol solo ejecutan).

**Precondición:** clase `propio`. En `referencia` la acción **no existe** (BR-1) — no se pinta
habilitada ni deshabilitada-con-esperanza: `disabled` con el motivo literal ya firmado en el mockup
(«no aplica: solo arneses propios»).

**Destino:** `~/.arnesia/checkouts/<home-slug>/<id>/`. Nada se escribe fuera de `~/.arnesia`
(`superficie-local-confinada`); en particular **jamás** dentro de `~/.claude`, que ya está en la lista
de protegidos de `validarRootPortafolio`.

**Dos caminos, elegidos por la forma de `source` (AG-D13):**

| camino | cuándo | mecanismo |
|---|---|---|
| **A · local** | `installLocation` existe en disco **y** `source` es ruta relativa | copia de la subcarpeta. Sin red, sin auth, sin git. Es el caso de `prenter-marketplace` |
| **B · externo** | `source` es objeto (`git-subdir` / `url` / `github`) | clone shallow **pineado por `sha`** → se verifica que el commit materializado sea ese → se extrae solo `path`. Auth `gh` → PAT fallback |

> **Corregido 2026-07-25 (design.md §2 C16 · evidencia verificada).** La v1 de este párrafo decía
> «clone con `ref`, y si viene `sha` se verifica». Eso **abortaría el 100 % del camino B**. Medido
> sobre el catálogo oficial real (273 entradas, 220 con `source` objeto):
>
> - **`sha` está en 220/220; `ref` solo en 77.** El pin universal es el `sha`.
> - En una entrada **sana** los dos apuntan a commits **distintos**:
>   `42crunch-api-security-testing` declara `ref: v1.5.5` + `sha: 30287f5e…`, y contra el remoto real
>   `refs/tags/v1.5.5` = `faf53053…` mientras `30287f5e…` es **HEAD del default branch**. O sea: el
>   `sha` del catálogo sigue a HEAD, no al tag.
> - La mayoría de los `ref` son `main` — una rama que se mueve. Pinear por `ref` no es reproducible.
>
> **Regla:** el `sha` es la autoridad. Una divergencia `ref` vs `sha` es **aviso visible**, nunca
> aborto — el dato se muestra citando los dos. `ref` se usa solo como fallback de fetch cuando el
> servidor rechaza el fetch-por-sha, y BR-16 sigue aplicando sobre el `sha`.
>
> **`source: "github"` trae DOS hashes** (`commit` y `sha`) con valores distintos y sin semántica
> documentada (verificado: `fullstory`, `jfrog`). Gana **`sha`** — es el único presente en las 220
> formas objeto. La divergencia se muestra como aviso; no se adivina ni se rechaza la entrada.

**Secuencia (los dos caminos):**

1. Se valida precondición y destino. Si el destino **ya tiene contenido**, se **aborta sin tocar nada**
   (BR-14) — pisar un canónico destruiría trabajo, y la ley anti-drift dice que el canónico es la única
   copia editable.
2. Se materializa en un directorio **temporal**, no en el destino final.
3. Camino B: se verifica `sha` si el catálogo lo declara. Si no coincide, se aborta (BR-16).
4. Se mueve del temporal al destino (operación atómica). Si cualquier paso anterior falló, se limpia el
   temporal y **no queda canónico parcial ni entrada registrada** (BR-15).
5. Se registra la entrada como **canónico**. **Ruta e identidad NO son lo mismo** (design.md §2 C17):
   - **ruta** = `~/.arnesia/checkouts/<slug(nombre-del-marketplace)>/<slug(id)>/` — legible, y única
     porque el nombre del marketplace es la clave del registro.
   - **identidad** = `(home, id)` con **`home` = repo canonicalizado** (`CanonicalizarRepo(repo)`,
     RN-IDENT-1), **no** el nombre del marketplace.

   Si se registrara por nombre, el canónico recién traído **no cruzaría con su propia fila del
   catálogo** y la situación daría `no-lo-tengo` para siempre. Es el bug que E-106 vigila.
6. Se **evalúa `deriva` inmediatamente** y se muestra (BR-17). Recién traído debería dar `al-hilo`; si
   da otra cosa, es un dato honesto que el operador tiene que ver, no un error a esconder.

**Auth (camino B):** `gh` si está autenticado, PAT como fallback — user-owned, por el auth-terms ya
firmado. **ArnesIA nunca pide, guarda ni proxya credenciales propias**: es conductor (BR-18).

## 5 · Reglas de negocio (BR)

| id | regla | por qué |
|---|---|---|
| BR-1 | Un marketplace `referencia` **jamás** habilita `↧ Traer canónico` | «operar arneses de terceros» está muerto por visión (`vision.md` §Qué mutó · `CLAUDE.md` «solo arneses propios») |
| BR-2 | `version` del catálogo sale por **precedencia con procedencia anotada**: (1) `plugins[].version` no-nulo → `campo-version` · (2) último segmento de la ruta de `source` que parsee semver → `derivada-de-source` · (3) nada → ausente, **jamás inventada**. *(v2 por AG-D14: `version` SÍ es campo estándar, 273/273 en el catálogo oficial y con `$schema` publicado — hay una fuente mejor que la ruta y estaba a la vista. El espíritu de la v1 queda intacto.)* | AG-D10 + AG-D14; inventarla sería un dato fabricado |
| BR-2b | `source` se acepta como **`string` u objeto** (`git-subdir`/`url`/`github`) y se normaliza; un `source` objeto lleva `ref`/`sha`, que **no** son semver → suele dar `no-comparable`, no una versión falsa | AG-D13: 220 de 273 entradas del catálogo oficial usan la forma objeto |
| BR-3 | Catálogo **cacheado** + refresco explícito; la fila siempre dice **cuándo** se leyó | app = conductor, no proxy; legible sin red |
| BR-4 | Marketplace inalcanzable muestra **motivo textual**, nunca lista vacía | una lista vacía se lee como «no tiene arneses» = pass fabricado (mismo mal que G3 mató) |
| BR-5 | `Validar` no pinta ✓ sin `marketplace.json` real leído | criterio G3, no se relaja |
| BR-6 | Los hallazgos **no** se premarcan; el operador tilda | AG-D7 |
| BR-7 | Registrar un marketplace existente **no** pisa ni duplica | anti-drift: el operador decide, el sistema no sobreescribe solo |
| BR-8 | Discrepancias entre eslabones se **muestran**, no se resuelven solas | idéntico a `discrepancias` de `ResolverOrigen` (C-OR-6) |
| BR-9 | Situación `no-comparable` es una rama de primera clase con motivo | §4.3; alternativa = mentir o callar |
| BR-10 | Publicar / Actualizar / Reparar quedan `disabled` + tooltip a su ítem | fuera de alcance §0, con gate propio |
| BR-11 | Reconciliar escribe `home` declarado; **no** clona, no instala | asignar origen ≠ traer |
| BR-12 | Toda superficie nueva es **superset**: nada firmado en PARIDAD Slice 1 se quita | disciplina de paquete |
| BR-13 | El destino de `Traer` es `~/.arnesia/checkouts/<home-slug>/<id>/`; **nada** se escribe fuera de `~/.arnesia`, y jamás dentro de `~/.claude` | `superficie-local-confinada`; `~/.claude` ya es protegido en `validarRootPortafolio` |
| BR-14 | Si el destino ya tiene contenido, `Traer` **aborta sin tocar nada** e informa | pisar el canónico destruye trabajo; es la única copia editable (ley anti-drift) |
| BR-15 | Materialización **atómica**: temporal → mover. Un `Traer` interrumpido no deja canónico parcial **ni** entrada registrada | media-copia registrada como canónico es peor que no tenerla |
| BR-16 | El pin es el **`sha`**: si el commit materializado no es ese, se **aborta**. Una divergencia `ref` vs `sha` (o `commit` vs `sha`) es **aviso visible**, no aborto | jamás registrar como canónico algo que no es lo declarado — pero `ref` no es autoridad: sigue a una rama móvil, y en el catálogo real diverge del `sha` en entradas sanas (v2 por C16/C18) |
| BR-17 | Tras traer, `deriva` se evalúa **de inmediato** y se muestra tal cual salga | si recién traído no da `al-hilo`, es un dato honesto que hay que ver |
| BR-18 | Auth = `gh` → PAT fallback, user-owned. ArnesIA **no** pide, guarda ni proxya credenciales | app = conductor (auth-terms firmado) |

## 6 · Escenarios (dan los casos de prueba)

Con **datos reales de esta máquina** (AG-D9/D12). `gh` autenticado como `alpacapurpura`.

| id | escenario | insumo real | resultado esperado |
|---|---|---|---|
| E-01 | plano nace poblado por CC | `known_marketplaces.json` (5 entradas) | 5 filas, eslabón `cc-known-marketplaces` |
| E-02 | catálogo propio legible offline | `~/.claude/plugins/marketplaces/prenter-marketplace` | 2 entradas (`harness`, `harness-beta`) |
| E-03 | dos entradas, mismo `source` | ambas → `./plugins/harness/0.5.3` | `comparte_source_con` poblado; chip explicativo |
| E-04 | versión derivada de la ruta | `./plugins/harness/0.5.3` | `version = "0.5.3"` |
| E-05 | enriquecimiento `catalogo.json` | `canales{estable,beta}` + 4 `versiones` | `estado_canal` presente; `0.5.1` = `deprecada` |
| E-06 | marketplace sin `catalogo.json` | `claude-plugins-official` | degrada sin error, `estado_canal` ausente |
| E-07 | clase referencia sin Traer | `anthropics/claude-plugins-official` | `Traer canónico` `disabled` + tooltip (BR-1) |
| E-08 | 0 hallazgos honesto | escanear `/home/chalreme/Proyectos/vitalia` | mensaje honesto + `PasoFuente` re-montado (AG-D12) |
| E-09 | hallazgos reales positivos | escanear `/home/chalreme/Proyectos/luana-platform` | ≥5 candidatos, nada premarcado, botón `Agregar 0` disabled |
| E-10 | aviso real de eslabón | escanear este repo | aviso «enabledPlugins declara harness@prenter-marketplace pero sin record de instalación» |
| E-11 | Atrás desde candidatos | E-09 → `Atrás` | vuelve a `fuente`, ruta preservada |
| E-12 | `cambiar ruta` | E-09 → `cambiar ruta` | idem E-11 (misma transición, otra puerta) |
| E-13 | Cancelar = cero efectos | E-09 → `Cancelar` | portafolio sin cambios (S1-D9) |
| E-14 | cierre bloqueado en `agregando` | POST en vuelo | ✕/Esc no cierran (S1-D19) |
| E-15 | validar url real | `https://github.com/alpacapurpura/prenter-marketplace` | ✓ con `name`+`owner`+2 plugins reales |
| E-16 | validar url inexistente | `https://github.com/alpacapurpura/no-existe-xyz` | error textual, cero ✓ (BR-5) |
| E-17 | validar repo sin `marketplace.json` | `https://github.com/alpacapurpura/vitalia` | error textual explícito («no es un marketplace») |
| E-18 | registrar duplicado | registrar `prenter-marketplace` otra vez | no duplica, informa y ofrece ir (BR-7) |
| E-19 | registrar y aterrizar | url nueva válida | modal cierra, catálogo abierto |
| E-20 | situación `mi-copia-adelantada` | canónico 0.5.4 vs estante 0.5.3 | `Publicar` visible, `disabled` + tooltip ítem 3 |
| E-21 | situación `estante-adelantado` | canónico 0.5.2 vs estante 0.5.3 | `Actualizar mi copia` disabled + tooltip ítem 4 |
| E-22 | situación `instalaciones-en-deriva` | instalación con hash ≠ ref | `Reparar` disabled + tooltip ítem 5 |
| E-23 | situación `no-comparable` | entrada sin `version` derivable | `sin-senal` + motivo (BR-9) |
| E-24 | discrepancia entre eslabones | repo distinto CC vs declarado | ambas señales visibles (BR-8) |
| E-25 | reconciliar con confirmación | arnés sin `home` | `Confirmar` nace disabled; al elegir, habilita; escribe `home` sin clonar |
| E-26 | reconciliar «ninguno» | idem | queda sin origen, honesto, sin `home` inventado |
| E-27 | toolbar sin ambigüedad | plano Arneses | rótulos `VER POR`/`FILTROS` visibles; las dos «Marketplace» distinguibles (L1) |
| E-28 | lazy loading | catálogo con >20 entradas | render incremental, sin truncar en silencio |
| E-29 | sin red | daemon sin red, catálogo cacheado | catálogo se muestra + `leído hace <t>` (BR-3) |
| E-30 | a11y de modales | S6, S7 | focus trap, Esc, `role=dialog` + `aria-modal` |

**`↧ Traer canónico` (AG-D17).** El arquitecto amplía y afina esta tabla en `plan-pruebas.md`.

| id | escenario | insumo real | resultado esperado |
|---|---|---|---|
| E-76 | camino A feliz | `harness@prenter-marketplace`, `source ./plugins/harness/0.5.3`, ya en disco | copia a `~/.arnesia/checkouts/prenter-marketplace/harness/`, registrado canónico, **sin red** |
| E-77 | deriva tras traer | E-76 | `deriva` evaluada de inmediato; se muestra el valor real (BR-17) |
| E-78 | canal comparte `source` | traer `harness-beta` tras E-76 | destino propio por `id`; **no** pisa el de `harness` |
| E-79 | destino ya poblado | repetir E-76 | **aborta sin tocar nada** + informa; el canónico previo intacto (BR-14) |
| E-80 | clase `referencia` | `frontend-design@claude-plugins-official` | acción `disabled` + literal «no aplica: solo arneses propios» (BR-1) |
| E-81 | camino B feliz | entrada `git-subdir` real del catálogo oficial | clone shallow por `ref`, extrae solo `path`, registra canónico |
| E-82 | `sha` que no coincide | `sha` alterado a mano en el caché | **aborta**, nada registrado (BR-16) |
| E-83 | `ref` inexistente | `ref` inventado | error textual, nada registrado |
| E-84 | sin `gh` ni PAT | credencial ausente, repo privado | error de auth distinguible de «no existe»; ArnesIA no pide credenciales (BR-18) |
| E-85 | 401/403 vs 404 | repo privado vs inexistente | motivos **distintos**, ninguno genérico |
| E-86 | interrupción a mitad | matar el clone en vuelo | temporal limpio, **cero** canónico parcial, **cero** entrada registrada (BR-15) |
| E-87 | destino fuera de `~/.arnesia` | intento de escape por `..` en `home-slug`/`id` | rechazado (BR-13) |
| E-88 | jamás escribe en `~/.claude` | cualquier camino | verificado por aserción de ruta, no por confianza (BR-13) |
| E-89 | disco lleno / sin permiso de escritura | destino no escribible | error textual, temporal limpio, nada registrado |
| E-90 | `source` no reconocido | forma desconocida | acción `disabled` + motivo con el `source` crudo visible; **no** se intenta adivinar |

## 7 · Mecanismos de validación y verificación

Cinco capas, ninguna opcional. Detalle ejecutable (comandos, fixtures, ubicación de cada test) en
[`design.md`](./design.md) §Plan de pruebas.

1. **Dominio (Go, tabla)** — derivación de `version` desde `source`, merge collect-all con
   discrepancias, `situacionDe()` con las 6 ramas y su precedencia, degradado de `catalogo.json`
   ausente. Tests colocados junto al código (patrón sancionado en la auditoría de carga, pass 42→46).
2. **Selectores FE (vitest)** — `entities/marketplace/model/selectors.test.ts`, tabla de casos con
   fixtures de **shape real** (copiados de la máquina, no inventados).
3. **Stories con `play()` (Storybook + vitest-browser)** — una story por estado de cada superficie,
   incluidos los degradados (`sin-acceso`, `no-leido`, `no-comparable`, `referencia` sin Traer). Gate
   a11y en `error` (ya configurado en `.storybook/preview.ts`).
   ⚠ **Gotcha conocido:** vitest-browser **no corre en background** (Chromium no headless) — se corre
   `pnpm --dir web run verify` y las stories en sesión interactiva.
4. **E2E vivo contra la máquina real** — los 30 escenarios de §6 con los insumos reales listados.
   Prohibido mock donde hay dato real disponible.
5. **As-code / fitness** — capabilities nuevas (R1 integridad · R2 cobertura · R3 gate-commit · R4
   estado-generado), `arnesia conformance --todo` sin regresión, `depcruise` + `steiger` para la
   dirección de imports de la entidad/widgets nuevos, `stylelint` strict-value para el cero-color-literal.

## 8 · Riesgos y decisiones diferidas

- ~~AG-D11 (fila = canal) está PROPUESTA~~ → **FIRMADA 🧑‍⚖️ 2026-07-25**: una fila por canal, con chip
  `mismo contenido que <otra>` en las que comparten `source`. El nombre instalable
  (`<entrada>@<marketplace>`) queda **visible y copiable en cada fila** — es el dato operativo, no se
  esconde detrás de un chip. Descartado deduplicar por `source`. Ya no hay punto de cambio abierto acá.
- **`catalogo.json` es convención de prenter.** Si mañana se estandariza, el enriquecimiento pasa de
  opcional a primario. Hoy: opcional, degradado silencioso.
- **Auth.** `gh` está autenticado en esta máquina; el fallback PAT existe por auth-terms firmado
  (app = conductor, nunca proxya login). El E2E cubre el camino `gh`; el camino PAT se cubre con
  test de unidad del selector de credencial, no con un login real.
- **Deuda a11y `.text-warn`** sigue abierta en `BACKLOG.md`; este paquete no la agrava (§1).
