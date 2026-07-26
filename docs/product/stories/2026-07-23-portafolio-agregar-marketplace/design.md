# Diseño técnico · Agregación consolidada al Portafolio (proyecto + marketplace)

> `tipo: design` · paquete `2026-07-23-portafolio-agregar-marketplace` · 2026-07-25.
> Etapa 3 del flujo (METODOLOGIA §10), mitad técnica del par [`spec.md`](./spec.md) + este archivo.
> Insumos vinculantes: `spec.md` · [`decisiones.md`](./decisiones.md) AG-D1..D12 ·
> [`auditoria-storybook-vs-app.md`](./auditoria-storybook-vs-app.md) ·
> [`mockup-agregacion-consolidada.html`](./mockup-agregacion-consolidada.html) (FIRMADO 🧑‍⚖️).
> Plan de pruebas ejecutable en [`plan-pruebas.md`](./plan-pruebas.md).
>
> **Este documento es el contrato para quien implementa.** Cada nombre de tipo, cada firma, cada
> ruta de archivo y cada código HTTP están decididos acá. Si algo falta, es un bug de este
> documento — no una licencia para improvisar: **preguntá**, no inventes.

## 0 · Cómo leer esto

| § | Qué |
|---|---|
| 1 | **Evidencia nueva de la máquina real** que cambia el diseño respecto de AG-D9/D10 |
| 2 | **Contradicciones detectadas** (spec ↔ mockup ↔ boundary ↔ código) + recomendación |
| 3 | Contratos Go: `domain/` · `ports/` · `usecase/` · `adapters/` · composition root |
| 4 | Persistencia (dónde vive el registro y el caché, y por qué no en SQLite) |
| 5 | Lectura de catálogo (local-primero, remoto por `gh`, degradado honesto) |
| 6 | Cálculo de situación: tabla de verdad completa + precedencia + acción |
| 7 | Endpoints HTTP (método · ruta · JSON exacto · códigos y su semántica) |
| 8 | Frontend: árbol exacto, moléculas a `shared/ui/`, orquestación, gates de lint |
| 9 | Máquina de estados del wizard (las dos ramas) con `Atrás`/`Cancelar` |
| 10 | Tokens y CSS |
| 11 | Arquitectura as-code: capabilities, boundaries, fitness, contratos |
| 12 | Orden de construcción sugerido y riesgos |

---

## 1 · Evidencia nueva de la máquina real (verificada 2026-07-25)

AG-D9/D10 se firmaron «por evidencia», pero la muestra fue **un solo marketplace**
(`prenter-marketplace`, 2 entradas). Inspeccionando **los 5** de esta laptop aparecen cuatro
hechos que **cambian el contrato** y que ninguna decisión del paquete había capturado. Se
proponen como **AG-D13..AG-D16** (evidencia, no preferencia — el operador solo tiene que
ratificar que se leyeron bien).

### AG-D13 · `source` tiene CUATRO formas reales, no una

`claude-plugins-official/.claude-plugin/marketplace.json` (159 KB, **273** entradas):

| forma de `source` | cuántas | ejemplo |
|---|---|---|
| string ruta relativa | 53 | `"./plugins/agent-sdk-dev"` · `"./external_plugins/asana"` · `"./"` (caveman, ponytail) |
| objeto `source:"git-subdir"` | 78 | `{"source":"git-subdir","url":"https://github.com/42Crunch-AI/claude-plugins.git","path":"plugins/api-security-testing","ref":"v1.5.5","sha":"30287f5e…"}` |
| objeto `source:"url"` | 140 | idem con `url` |
| objeto `source:"github"` | 2 | idem con `repo` |

**Consecuencia:** `AG-D10 punto 2` («`source` es una ruta relativa que lleva la versión adentro»)
es cierto para prenter y falso para el 80 % del catálogo oficial. El parser **debe** aceptar
`string | objeto` y normalizar a `domain.SourceCatalogo` (§3.1.3). Un `source` objeto **no** tiene
versión en la ruta: la lleva en `ref`/`sha`, que NO son semver.

### AG-D14 · `version` SÍ es un campo estándar de `plugins[]`

`"version"` aparece en **273/273** entradas del catálogo oficial: **259 con `null` explícito** y
**14 con semver real** (`clangd-lsp: "1.0.0"`, `warp: "2.1.0"`…). Y el archivo declara
`"$schema": "https://anthropic.com/claude-code/marketplace.schema.json"` — hay un esquema oficial.

**Consecuencia:** BR-2 («`version` se **deriva** de `source`») pasa de *derivación exclusiva* a
**precedencia con procedencia anotada**:

1. `plugins[].version` no vacío y ≠ `null` → `procedencia = "campo-version"`.
2. último segmento de la ruta de `source` que parsee semver → `procedencia = "derivada-de-source"`.
3. nada → `version = ""`, `procedencia = ""` (**ausente, jamás inventada**).

El espíritu de BR-2 se conserva intacto: nunca se fabrica. Lo que cambia es que hay una fuente
mejor que la ruta y estaba a la vista.

### AG-D15 · el catálogo ajeno trae campos que el spec no modela

Claves reales presentes en `plugins[]` del catálogo oficial: `name`, `source`, `version`,
`description`, `strict`, `skills`, `lspServers`, `homepage`, `author`, `displayName`, `keywords`,
`category`, `tags`. Y a nivel raíz: `$schema`, `name`, `description` **(top-level, no
`metadata.description`)**, `owner`, `renames`, `plugins`.

- `owner` es `{name, email?}` (prenter, Anthropic) **o** `{name, url?}` (caveman, ponytail, warp).
  El spec asumía `owner.name` + email.
- `description` vive **top-level** en 4 de 5 marketplaces y bajo `metadata.description` solo en
  prenter. El lector acepta las dos y prefiere la top-level.
- `renames` es un mapa `nombre-viejo → nombre-nuevo` (6 entradas reales en el oficial). Es
  **exactamente** el insumo del escenario «arnés renombrado en el remoto» (§ plan-pruebas E-38).

**Consecuencia:** el adapter **traduce y descarta** lo que no modela (anti-corruption layer, §11.3);
`domain.EntradaCatalogo` no crece un campo por cada extra del formato ajeno. `renames` sí se
modela porque tiene efecto sobre identidad.

### AG-D16 · el catálogo oficial es 273 entradas / 159 KB, no 41

El mockup dibuja «41 arneses en catálogo» para `claude-plugins-official`. El real es **273**. La
cifra del dibujo es ilustrativa; el FE **jamás** la teclea (sale del wire). Consecuencias duras:

- `ListaLazy` (AG-D3) deja de ser confort y pasa a ser **requisito**: 273 filas × (emblema + chips
  + situación + acción) no se montan de una.
- El plano Marketplaces **no puede** parsear los 5 `marketplace.json` en cada `GET` para contar
  entradas → justifica el caché (§4.2), que no es «offline nice-to-have» sino O(1) por fila.
- Hay un techo de tamaño a defender en el lector remoto (§5.3).

---

## 2 · Contradicciones detectadas

Ninguna se resuelve en silencio. Para cada una: qué choca, opciones, recomendación.

| id | contradicción | opciones | recomendación |
|---|---|---|---|
| **C1** | **`spec.md §2.2` pone `situacionDe()` en `entities/marketplace/model/selectors.ts`**, pero cruzar catálogo × portafolio ahí exige importar `EntradaPortafolio` de `entities/portafolio` = **cross-import entity↔entity**, que `steiger` `fsd/no-cross-imports` rompe (y `pnpm run fsd` está dentro de `verify`; el override de `steiger.config.ts` solo exime `./src/entities/*/@x/**`). | (a) la situación se calcula en el **dominio Go** y viaja en el wire; el selector FE solo *presenta* · (b) abrir `entities/portafolio/@x/marketplace.ts` · (c) apagar la regla | **(a)**. Es además lo que el propio spec ordena («NUNCA se calcula en el widget») y lo que §7.1 exige (tabla Go de las 6 ramas). El selector FE se renombra a `etiquetaDeSituacion`/`tonoDeSituacion` — presentación, no cálculo. Cero cross-import, cero regla apagada. |
| **C2** | ~~**`spec.md §2.1` dice que `↧ Traer canónico` «se habilita»**, pero §0 no lo declaraba en alcance, ningún §4/§5/§6 definía su mecanismo y ningún escenario lo ejercitaba.~~ | (a) construir completo · (b) diferir con motivo | **RESUELTA por el operador · `AG-D17` FIRMADA 🧑‍⚖️.** Se construye **completo, los dos caminos**. La asimetría que lo justifica (y por la que Publicar/Actualizar/Reparar siguen fuera): `Traer` escribe **solo en `~/.arnesia/checkouts/`**; los otros tres escriben en cosas del cliente o en el estante remoto. Era un hueco del spec, no del diseño. Mecanismo → **§13**. |
| **C3** | **BR-11 «Reconciliar escribe el `home` declarado»** + el precedente `Identificar` (S1-D28) escribe el sello `arnes.l0.json` in-situ. Pero la instalación dominante real es `referenciada-cc`, cuyo dir vive en `~/.claude/plugins/cache/…`, y **`~/.claude` está en la lista `protegidos` de `validarRootPortafolio`** (`internal/usecase/portafolio.go`): escribir el sello ahí es **imposible por diseño**. | (a) escribir el sello (rompe en el caso dominante) · (b) declarar el `home` en el **store del Portafolio** + re-key + eslabón `declarado-por-operador` · (c) relajar el guardrail | **(b)**. Nada se escribe fuera de `~/.arnesia/` (respeta `superficie-local-confinada`), la procedencia queda anotada (BR-3: el eslabón dice «declarado-por-operador», no «manifiesto»), y funciona para las 3 formas de instalación. (c) queda descartado: relajar el guardrail para escribir en `~/.claude` es peor que el problema. |
| **C4** | **BR-4 «nunca lista vacía»** no tiene realización en el wire si `entradas` viaja siempre como array: `[]` es indistinguible de «no pude leer». | (a) `entradas: null` cuando no hay lectura vs `[]` cuando el marketplace declara `plugins: []` de verdad · (b) un flag aparte | **(a)**, y es normativo: en Go `Entradas []EntradaCatalogo` **sin** `omitempty` (nil → `null`); en TS `entradas: EntradaCatalogo[] \| null`. `null` = «no sé», `[]` = «leí y no declara ninguno, hace 4 min». Es el mecanismo anti-pass-fabricado a nivel cable. |
| **C5** | **Mockup vs BR-10:** los botones `Publicar` (línea 601), `Actualizar mi copia` (617) y `Reparar` (634) del mockup **no** llevan `disabled` ni `title` — solo un `<span class="badge nuevo">ítem N</span>` al lado. BR-10 exige `disabled` + tooltip. | (a) spec gana · (b) mockup gana | **(a) spec gana.** El mockup es un dibujo estático sin backend; BR-10 es la regla. La única fila del mockup con el par correcto es la de `referencia`: `disabled title="no aplica: solo arneses propios"` — **ese literal se conserva tal cual** (§6.3). |
| **C6** | **Mockup vs spec §4.5.2 / E-15:** el mockup **no dibuja** la pantalla de éxito de `Validar` (`name` + `owner.name` + nº de `plugins[]`); `registrarYAterrizar()` solo cierra el modal. | (a) spec gana (se amplía el dibujo) · (b) recortar el spec | **(a)**. Es superset del mockup, no contradicción de fondo: BR-5/G3 exige que el ✓ solo aparezca con un archivo real leído, y para eso hay que mostrar QUÉ se leyó. §9.2 define el estado `validado` con esos 3 datos. |
| **C7** | **El spec se contradice con el mockup sobre `no-comparable` y los canales — y el que está stale es el SPEC.** `spec.md §4.3` afirma «la 6ta rama **no está en el mockup**», pero el mockup **sí la dibuja** (línea 670: `<span class="pf-dot-salud sin-senal"></span> no comparable`) y **también dibuja el chip de canales** de AG-D11 (línea 618: `mismo contenido que harness`). El texto del spec quedó de una versión anterior del dibujo. *(Verificado por lectura directa del HTML; un grep por `no-comparable` con guion no la encuentra — en el dibujo va con espacio.)* | (a) corregir la frase del spec · (b) dejarla | **(a)**. No cambia ninguna decisión: `no-comparable` sigue siendo obligatoria y el diseño ya la trata como rama de primera clase; lo que cambia es que **el dibujo firmado ya la respalda**, así que no hay superset que justificar en este punto. Se pinta con `DotSaludPortafolio salud="sin-senal"` + motivo, exactamente como el mockup (cero vocabulario nuevo). Corregir `spec.md §4.3` y `§8`. |
| **C8** | **AG-D10 vs máquina real:** `source` no es siempre string y `version` SÍ es campo estándar (§1, AG-D13/D14). | — | AG-D13/AG-D14 supersede parcialmente AG-D10 **sin invertir su espíritu** (nada se inventa). Ratificación del operador pedida. |
| **C9** | **Vocabulario de eslabones:** AG-D9 nombra el eslabón `cc-known-marketplaces`, pero el enum Go vigente de procedencia (`domain.EslabonOrigen.Fuente`, `scanner.go`) usa `cc-plugins`. | (a) dos enums distintos y documentados · (b) unificar | **(a)**. `EslabonMarketplace` (de dónde salió el CONOCIMIENTO de un marketplace) y `EslabonOrigen.Fuente` (de dónde salió un dato de UNA COPIA) son ejes distintos; unificarlos confunde dos preguntas. Se documenta en el doc-comment de ambos tipos. |
| **C10** | **E-09 vs máquina real:** «`luana-platform` → 5 plugins de `claude-plugins-official`». Real: 6 claves en `enabledPlugins`, **4 en `true`**; y `harness@prenter-marketplace` **no tiene record** para `luana-platform` (su record es de `luana-vitalia`). | — | El aserto «≥5 candidatos» **se cumple** pero por otra composición: 3 `referenciada-cc` con record + 1 aviso «sin record de instalación» + 1 `proyecto-instalado` del root. `plan-pruebas.md` lo corrige con la cuenta real. |
| **C11** | **AG-D11 está PROPUESTA, no firmada.** | — | Se diseña para la propuesta (fila = entrada de índice/canal). **Punto de cambio exacto** si el operador prefiere deduplicar por `source`: §8.6. El backend **no cambia** en ninguno de los dos casos. |
| **C12** | ~~**AG-D11 aparece FIRMADA en dos lados y PROPUESTA en el otro.**~~ **RETIRADA — no existía.** Fue una **lectura stale mía**: leí `decisiones.md` mientras se registraba la firma. `AG-D11` está **FIRMADA 🧑‍⚖️** (línea 425: «el operador eligió “una fila por canal” tras ver las dos opciones renderizadas»). No hubo superposición de estado de firma. | — | **Sin acción.** Consecuencia real: el punto de cambio de **§8.7 queda CERRADO** — la fila del catálogo es el canal y **no se deduplica por `source`**. El hedge se retira; nada del diseño cambia (estaba hecho para esta opción). |
| **C13** | **La fila de `mockups/INDEX.md` está stale en dos campos.** Dice estado «🎨 etapa 1 — **esperando Gate 1 🧑‍⚖️**», pero `decisiones.md` registra `GATE 1 · Mockup FIRMADO 🧑‍⚖️ (2026-07-25)`; y afirma que el mockup cuenta «**6** situaciones» y «AG-D11 firmada» (ver C12). | — | Corregir la fila: Gate 1 firmado + etapa 3 en curso, y alinear la afirmación de AG-D11 con lo que resuelva el operador en C12. |
| **C14** | **Cifras internas del mockup inconsistentes entre sí y con el dato real.** La fila del plano dice «6 arneses en catálogo» para prenter (línea ~470) y la cabecera de su catálogo dice «8 entradas de catálogo» (línea 546); el real es **2**. Para `claude-plugins-official` dice «41» y el real es **273** (AG-D16). | — | **No es un defecto de diseño, es un dibujo:** todas esas cifras salen del wire en la implementación y **el FE no teclea ninguna**. Se deja anotado para que nadie las hardcodee «para que quede igual al mockup», y `plan-pruebas.md` E-28 usa 273 (el real) como caso de lazy. |
| **C15** | **Duplicación del decoder de `known_marketplaces.json`.** `internal/adapters/portafolio/scanner.go` ya lo lee (`leerKnownMarketplaces`, privado) y `deriva.go` lo reusa. El adapter nuevo lo necesita, pero **go-arch-lint prohíbe adapter→adapter** (allow-list, default-deny). | (a) duplicar 20 líneas en el adapter nuevo · (b) extraer un componente nuevo al grafo solo para una struct JSON · (c) refactorizar `Referencias` para recibir los marketplaces por parámetro | **(a)**, documentada como deuda con razón + entrada propuesta a `BACKLOG.md` («unificar el decoder de metadata CC detrás de un puerto»). (b) mete un componente al grafo cuyo único contenido es un DTO; (c) toca código ya firmado con firma de puerto fija (`ports.DerivaEvaluator`). |
| **C16** | **El mecanismo de camino B del spec, tomado literal, abortaría en la primera entrada real.** `spec.md §4.7` dice «clone shallow con `ref` → si viene `sha`, se **verifica**». Verificado en vivo: en `42Crunch-AI/claude-plugins` el `ref: v1.5.5` resuelve a `faf5305385de8afe…` y el `sha` declarado es `30287f5e3f122a64…` — **commits distintos**. Y `ref` está en solo **77/220** `source` objeto, contra `sha` en **220/220**. | (a) `sha` es la autoridad, `ref` es pista/fallback · (b) clonar por `ref` y relajar BR-16 · (c) abortar cuando difieren | **(a)**. Se hace `fetch --depth 1 --filter=blob:none origin <sha>` (**probado en vivo**: `HEAD == sha` exacto, 664 KB) y `ref` solo se usa si no hay `sha`, o como reintento único si el remoto no permite fetch por sha. La divergencia `ref`≠`sha` queda como `Aviso` visible, no como aborto. (b) mataría la única verificación de integridad que el formato ofrece; (c) rechazaría entradas sanas. **Corregir la redacción de `spec.md §4.7`.** |
| **C17** | **`home` de la identidad: nombre del marketplace vs repo canonicalizado.** `spec.md §4.7` paso 5 dice que la identidad es «(nombre del marketplace, nombre de la entrada)» y E-76 espera el destino `~/.arnesia/checkouts/prenter-marketplace/harness/`. Pero `IdentidadArnes.Home` está documentado y construido como el repo **canonicalizado** (RN-IDENT-1, `ResolverIdentidad`), y el cruce de §6.2 compara contra `Registries` que un escaneo puebla con `github.com/alpacapurpura/prenter-marketplace`. | (a) separar: **path** por nombre, **identidad** por repo canónico · (b) identidad por nombre (rompe el cruce) · (c) path por repo canónico (destino ilegible) | **(a)**. Si el canónico se registrara con `Home = "prenter-marketplace"` **no cruzaría con la fila del catálogo de la que vino** — la columna de situación quedaría diciendo `no-lo-tengo` sobre algo que acabamos de traer. El path es almacenamiento (legible, y el nombre es la clave única del registro); la identidad es el contrato. Explicitado en §13.4 con la advertencia, porque es exactamente lo que un implementador colapsa. |
| **C18** | **`source: "github"` declara DOS hashes distintos sin semántica documentada:** `commit` **y** `sha`, y verificado en vivo que **los dos son commits válidos del mismo repo** (`fullstorydev/fullstory-skills`: `1ec5865e…` y `b20614e2…`). BR-16 dice «si el catálogo declara `sha`…» sin contemplar que haya dos. | (a) gana `sha`, la divergencia se muestra · (b) gana `commit` · (c) declarar la forma no-materializable | **(a)**. `sha` es el **único campo presente en las 220 formas objeto**; `commit` es un extra de la forma `github` (2 entradas). Elegir el universal no es adivinar: es preferir el campo que el formato usa siempre. La divergencia se anota como `Aviso` VISIBLE (BR-8), no se silencia. (c) sería más honesto si `sha` no estuviera en el 100 % de los casos; con ese dato, rechazar sería sobre-cautela. |
| **C19** | **`SourceCatalogo` (§3.1.3) no tiene campo para el 2º hash, pero §13.2/E-95 exigen citar `commit` Y `sha` en el aviso.** El `commit` de la forma `github` se perdía en el parser, así que el aviso de divergencia era incomputable. | (a) `SourceCatalogo.Commit string \`json:"commit,omitempty"\`` · (b) generar el aviso en el adapter y meterlo en `EntradaCatalogo.Aviso` · (c) no avisar | **(a) APLICADA.** Un campo, `omitempty`, y sigue siendo un dato NORMALIZADO (un hash de commit), no una clave cruda del formato ajeno: el aviso queda computable en el dominio puro, que es donde E-95 lo asserta. (b) rompía la aserción de E-95 (`plan.Avisos`); (c) violaba BR-8/C18. |
| **C20** | **El check `dominio-no-adopta-shape-ajeno` prohíbe que el dominio «mencione» `git-subdir` e `installLocation` — pero el propio §3.1.3/§3.1.4 los pone ahí.** `TipoSource("git-subdir")` es un valor del enum propio y `install_location` es un campo del modelo propio con doc-comment. La letra del check prohibía la TRADUCCIÓN, no la contaminación. | (a) acotar el scan a las claves CRUDAS del formato ajeno · (b) renombrar los valores del enum · (c) dejar el check sin enforcer | **(a) APLICADA.** El enforcer (`fitness/marketplace_shape_test.go:TestDominioNoAdoptaShapeAjeno`) escanea los **tags `json:`** de `internal/domain/marketplace*.go` buscando `$schema`, `lspServers`, `displayName`, `strict`, `keywords`, `category`, `tags`, `homepage`, `renames`, `metadata`, `installLocation`, `enabledPlugins`, `projectPath`. La fila del boundary se corrigió con esa lista. (b) inventaría vocabulario para nombrar lo mismo. |
| **C21** | **`ports.Materializador` (§13.3) devuelve `(sha, error)`, sin canal para los avisos** — pero §13.6 paso 3 y E-92 exigen que un symlink que escapa o roto deje `Aviso` VISIBLE, y el único que lo detecta es el adapter de copia. Con la firma literal, esos hallazgos se PERDÍAN. | (a) `(shaEfectivo string, avisos []string, err error)` · (b) loguearlos y no mostrarlos · (c) no anotar nada | **(a) APLICADA.** Es un puerto que nace en ESTE paquete (no una firma fija preexistente como `ports.DerivaEvaluator`), y (b)/(c) violan BR-8 directamente: un archivo del canónico que NO se copió y nadie lo dice es un silencio. |
| **C22** | **AG-D14/§1 afirma «259 con `null` explícito» y el dato real es OTRO.** Medido de nuevo sobre el archivo: **14 filas con semver, 259 SIN la clave `version`, y CERO con `null` explícito.** | — | **Corrección de dato, cero cambio de comportamiento.** El argumento de AG-D14 se sostiene igual (`version` ES campo estándar: 14 semver reales + `$schema` publicado). El parser usa `*string` y trata ausente / `null` / `""` como AUSENTE — las tres ramas se prueban con un fixture propio (`dañados/version-null-explicito/`) porque el archivo real no trae el caso `null`. |
| **C23** | **Los centinelas de §3.5/§13.3 están declarados en `usecase`, pero los PRODUCE el adapter y los CLASIFICA el transporte** — y go-arch-lint (allow-list, default-deny) no permite `usecase → adapter` ni `adapter → usecase`. Con la letra del diseño habría DOS instancias del mismo `ErrNoEsMarketplace` y el `errors.Is` del handler fallaría. | (a) declararlos en `internal/domain` y re-exportarlos desde `usecase` · (b) duplicar y traducir en cada frontera · (c) mover el mapeo de status al usecase | **(a) APLICADA.** `ErrNoEsMarketplace`, `ErrSinViaDeLectura`, `ErrTraerRemotoNoTiene`, `ErrTraerSinAuth`, `ErrTraerLocal` viven en `domain` (son valores puros: `errors.New` es stdlib, cero transporte) y `usecase` los re-exporta con los nombres EXACTOS del contrato — el §3.5 se cumple literal desde afuera. (b) es la clase de bug que `errors.Is` existe para evitar. *(Nota menor del mismo tipo: §3.5 declara `Corruptas []corruptaWire`, un tipo unexported del paquete `httpapi`; se usa `usecase.CorruptaMarketplace`, mismo shape.)* |
| **C24** | **§5.1 calcula la situación SOLO en el camino fresco**, así que un catálogo servido desde el caché devolvía la situación de cuando se cacheó. Tras un `Traer`, la fila seguía diciendo `no-lo-tengo` sobre algo que acabás de traer — justo lo que **E-106** vigila. | (a) recalcular la situación en TODOS los caminos y no persistirla nunca · (b) invalidar el caché en cada escritura del Portafolio · (c) dejarlo | **(a) APLICADA.** El cruce es puro y baratísimo (`CruzarConPortafolio`+`CalcularSituacion` sobre N filas en memoria); el caché guarda la LECTURA del estante, no el veredicto. (b) acopla el caché del estante a cada escritura del Portafolio (escaneo, Identificar, AsignarOrigen, Traer) y se olvida en el primer camino nuevo. |
| **C25** | **E-16 y E-17 exigen mensajes DISTINGUIBLES, pero `gh api …/contents/.claude-plugin/marketplace.json` devuelve el MISMO 404** para «el repo no existe» y para «el repo existe y no tiene el archivo». Verificado en vivo: `alpacapurpura/no-existe-xyz` y `alpacapurpura/vitalia` daban texto idéntico. | (a) sondar `repos/<owner>/<repo>` SOLO en la rama de error · (b) dejar los dos mensajes iguales · (c) inferirlo del nombre | **(a) APLICADA.** Un request extra únicamente cuando ya hubo 404, con `--jq .full_name` (campo que siempre está). Si la SONDA misma falla, se dice que no se pudo comprobar — jamás se le atribuye el 404 al repo sin haberlo verificado. (b) incumple el plan de pruebas y deja al operador sin saber si escribió mal la url. |

---

## 3 · Contratos Go

Regla que gobierna todo lo que sigue: **`dominio-independiente-de-transporte`** — `internal/domain`
no importa `net/http`, `database/sql`, `modernc.org/sqlite` ni ningún paquete de adapter; los
puertos viven en `internal/ports`; el shape crudo del archivo ajeno se traduce en el adapter
(§11.3, boundary nuevo).

### 3.1 · `internal/domain/marketplace.go` (nuevo)

#### 3.1.1 · Clase y eslabón

```go
// ClaseMarketplace parte el universo en dos, y NO son simétricas (AG-D8 decisión 1):
// `propio` = publicamos ahí (participa de traer/publicar/actualizar/reparar);
// `referencia` = solo resuelve la procedencia de un arnés ajeno — catálogo read-only,
// jamás operable (boundary marketplace-referencia-es-solo-procedencia).
type ClaseMarketplace string

const (
	ClasePropio     ClaseMarketplace = "propio"
	ClaseReferencia ClaseMarketplace = "referencia"
)

// ClaseValida reporta si c es una de las dos clases conocidas. Una clase desconocida NUNCA
// degrada a `propio` (eso habilitaría operar lo ajeno): degrada a `referencia`, el lado seguro.
func ClaseValida(c ClaseMarketplace) bool

// ClaseSegura devuelve c si es válida, y ClaseReferencia si no lo es — fail-safe explícito.
func ClaseSegura(c ClaseMarketplace) ClaseMarketplace

// EslabonMarketplace dice de dónde salió el CONOCIMIENTO de que este marketplace existe
// (AG-D9, collect-all: no exclusivo, ≥1, los dos = detectado + confirmado). Eje DISTINTO de
// EslabonOrigen.Fuente (portafolio.go), que dice de dónde salió un dato de UNA COPIA
// instalada; no unificar (C9 de design.md): son dos preguntas.
type EslabonMarketplace string

const (
	// EslabonCCKnown: `~/.claude/plugins/known_marketplaces.json` — registro propio de Claude
	// Code. Da checkout local gratis (`installLocation`) y fecha de última actualización real.
	EslabonCCKnown EslabonMarketplace = "cc-known-marketplaces"
	// EslabonDeclarado: lo registró el operador por el wizard (S6).
	EslabonDeclarado EslabonMarketplace = "declarado-por-operador"
)
```

#### 3.1.2 · Estado de lectura

Go no tiene sum types: se modela como struct discriminada por `Tipo`, que serializa **exacto** a
la unión etiquetada del FE (`spec.md §3`).

```go
// TipoLectura son los 4 estados explícitos de AG-D8 decisión 4. `no-leido` es el CERO del tipo
// a propósito: un EstadoLectura sin poblar dice «no leído aún», nunca «leído con 0 entradas».
type TipoLectura string

const (
	LecturaNoLeida       TipoLectura = "no-leido"
	LecturaLeida         TipoLectura = "leido"
	LecturaSinAcceso     TipoLectura = "sin-acceso"
	LecturaURLNoResuelve TipoLectura = "url-no-resuelve"
)

// EstadoLectura es el veredicto de la última lectura de catálogo de un marketplace.
// Invariante (BR-4): Tipo==LecturaLeida ⟺ Entradas es una cuenta REAL y Cuando≠"".
// Cualquier otro Tipo EXIGE Motivo≠"" — un degradado sin motivo es un degradado mudo.
type EstadoLectura struct {
	Tipo TipoLectura `json:"tipo"`
	// Cuando es RFC3339 UTC de la última lectura EXITOSA (aunque Tipo ya no sea `leido`:
	// así la fila puede decir «leído hace 2 días · ahora sin acceso» — BR-3 + BR-4 juntas).
	Cuando string `json:"cuando,omitempty"`
	// Entradas es la cuenta de filas del catálogo de esa última lectura exitosa. Solo
	// significativa con Cuando≠"".
	Entradas int `json:"entradas,omitempty"`
	// Motivo es el texto que la UI muestra tal cual (gh no autenticado, 404, JSON inválido…).
	Motivo string `json:"motivo,omitempty"`
	// Fuente de esa lectura: "local" (checkout de CC) | "remoto" (gh api) | "" (nunca leído).
	Fuente string `json:"fuente,omitempty"`
}
```

#### 3.1.3 · `SourceCatalogo` — el shape ajeno YA normalizado

```go
// TipoSource son las formas reales que `plugins[].source` toma en los marketplaces de esta
// máquina (AG-D13, verificado sobre 273+2+1+1+1 entradas). El adapter traduce; el dominio
// solo ve esto normalizado.
type TipoSource string

const (
	SourceRutaRelativa TipoSource = "ruta-relativa" // "./plugins/harness/0.5.3", "./"
	SourceGitSubdir    TipoSource = "git-subdir"
	SourceURL          TipoSource = "url"
	SourceGitHub       TipoSource = "github"
	SourceDesconocido  TipoSource = "desconocido" // shape que el adapter no reconoció: VISIBLE, no descartado.
)

// SourceCatalogo es `plugins[].source` normalizado. Crudo es la representación textual estable
// que se usa para agrupar canales (AG-D11): para una ruta relativa es la ruta tal cual; para un
// objeto es "<tipo>:<url|repo>#<path>@<ref>" — determinista, comparable, y jamás se muestra como
// si fuera una ruta local.
type SourceCatalogo struct {
	Tipo  TipoSource `json:"tipo"`
	Crudo string     `json:"crudo"`
	Ruta  string     `json:"ruta,omitempty"` // ruta relativa, o `path` del git-subdir.
	URL   string     `json:"url,omitempty"`
	Repo  string     `json:"repo,omitempty"`
	Ref   string     `json:"ref,omitempty"`
	SHA   string     `json:"sha,omitempty"`
}
```

#### 3.1.4 · `MarketplaceConocido`

```go
// MarketplaceConocido es UNA fila del plano Marketplaces (S2), ya mergeada collect-all
// (AG-D9). La clave de merge es Nombre — el mismo nombre con el que Claude Code keyea
// `installed_plugins.json` (`"<id>@<nombre>"`), o sea el nombre INSTALABLE, no el repo.
type MarketplaceConocido struct {
	Nombre string `json:"nombre"`
	// Repo canonicalizado con domain.CanonicalizarRepo → "host/owner/repo" (RN-IDENT-1).
	// "" si ninguna fuente lo declara o ninguna canonicaliza (el crudo queda en Discrepancias).
	Repo  string           `json:"repo,omitempty"`
	Clase ClaseMarketplace `json:"clase"`
	// Eslabones ordenados y dedupeados (detectado antes que declarado), ≥1 siempre.
	Eslabones []EslabonMarketplace `json:"eslabones"`
	// InstallLocation es el checkout local que Claude Code ya mantiene — la ruta contra la que
	// `deriva` compara (`<InstallLocation>/plugins/<id>/<version>/`). Solo del eslabón CC.
	InstallLocation string `json:"install_location,omitempty"`
	// CCActualizado es el `lastUpdated` que CC declara (RFC3339). Dato de CC, NO nuestra lectura:
	// nunca se muestra como «leído hace…» (eso es Lectura.Cuando).
	CCActualizado string        `json:"cc_actualizado,omitempty"`
	Lectura       EstadoLectura `json:"lectura"`
	// Discrepancias entre eslabones (repo distinto, clase distinta): se MUESTRAN, jamás se
	// eligen en silencio (BR-8, mismo criterio que OrigenPortafolio.Discrepancias / C-OR-6).
	Discrepancias []string `json:"discrepancias,omitempty"`
	// Registrado es RFC3339 de cuándo el operador lo declaró; "" si solo lo detectó CC.
	Registrado string `json:"registrado,omitempty"`
}
```

#### 3.1.5 · `EntradaCatalogo`

```go
// ProcedenciaVersion dice de DÓNDE salió Version — BR-2 con AG-D14: nunca se inventa, y
// siempre se puede auditar cuál de las dos fuentes ganó.
type ProcedenciaVersion string

const (
	VersionDeCampo    ProcedenciaVersion = "campo-version"       // plugins[].version
	VersionDeSource   ProcedenciaVersion = "derivada-de-source"  // último segmento semver de la ruta
	VersionAusente    ProcedenciaVersion = ""                    // no hay dato: Version==""
)

// EntradaCatalogo es UNA fila del catálogo (AG-D11: la ENTRADA DE ÍNDICE, o sea el canal —
// lo que el cliente instala con `/plugin install <nombre>@<marketplace>`; dos entradas pueden
// compartir Source y NO son dos arneses).
type EntradaCatalogo struct {
	Nombre      string         `json:"nombre"`
	Source      SourceCatalogo `json:"source"`
	Version     string         `json:"version,omitempty"`
	VersionDe   ProcedenciaVersion `json:"version_de,omitempty"`
	Descripcion string         `json:"descripcion,omitempty"`
	// ComparteSourceCon son los OTROS Nombre del mismo catálogo con idéntico Source.Crudo
	// (AG-D11: canales). Vacío = fila única.
	ComparteSourceCon []string `json:"comparte_source_con,omitempty"`
	// EstadoCanal viene SOLO del enriquecimiento opcional `catalogo.json` (convención de
	// prenter, NO del estándar): "habilitada" | "deprecada" | "" si no hay archivo.
	EstadoCanal string `json:"estado_canal,omitempty"`
	// NombreAnterior: si `renames` del marketplace mapea algún nombre viejo a este, queda acá
	// (AG-D15) — así un arnés del Portafolio keyeado con el nombre viejo sigue cruzando.
	NombreAnterior []string `json:"nombre_anterior,omitempty"`
	// Aviso son problemas de CALIDAD DEL DATO de esta fila, no de nuestro cálculo: nombre
	// duplicado en el catálogo, `source` que apunta a un dir ausente del checkout, shape de
	// `source` no reconocido. Se MUESTRAN completos (mismo criterio que Instalacion.Aviso /
	// BR-8) y jamás se convierten en un descarte silencioso de la fila.
	Aviso []string `json:"aviso,omitempty"`
	// Situacion y Accion las calcula el dominio al cruzar con el Portafolio (§6). Nunca las
	// calcula el widget (spec §4.3).
	Situacion SituacionCatalogo `json:"situacion"`
	Accion    AccionCatalogo    `json:"accion"`
}
```

#### 3.1.6 · `Catalogo`

```go
// Catalogo es el catálogo leído de UN marketplace. Entradas==nil ⟺ NO hay lectura válida
// (C4 de design.md: viaja `null`, no `[]` — «no sé» ≠ «no tiene arneses», BR-4).
// Entradas==[]EntradaCatalogo{} ⟺ el marketplace declara `plugins: []` de verdad.
type Catalogo struct {
	Marketplace string            `json:"marketplace"`
	Clase       ClaseMarketplace  `json:"clase"`
	Repo        string            `json:"repo,omitempty"`
	OwnerNombre string            `json:"owner_nombre,omitempty"`
	OwnerEmail  string            `json:"owner_email,omitempty"`
	OwnerURL    string            `json:"owner_url,omitempty"`
	Descripcion string            `json:"descripcion,omitempty"`
	Lectura     EstadoLectura     `json:"lectura"`
	Entradas    []EntradaCatalogo `json:"entradas"` // SIN omitempty — nil debe viajar como null.
	// Canales/Versiones: enriquecimiento opcional de `catalogo.json` (prenter). nil = ausente,
	// degradado SIN ruido (AG-D11 sub-hallazgo).
	Canales   map[string]string  `json:"canales,omitempty"`
	Versiones []VersionCatalogo  `json:"versiones,omitempty"`
	// Truncado: cuántas entradas se descartaron por el techo de seguridad (§5.3). >0 es un dato
	// VISIBLE, jamás un recorte silencioso (E-28).
	Truncado int `json:"truncado,omitempty"`
}

// VersionCatalogo es una fila de `catalogo.json#versiones[]` (convención de prenter).
type VersionCatalogo struct {
	Version string `json:"version"`
	Estado  string `json:"estado"` // habilitada | deprecada (crudo, sin enum: es formato ajeno).
	Fuente  string `json:"fuente,omitempty"`
	Fecha   string `json:"fecha,omitempty"`
}
```

#### 3.1.7 · Funciones puras del dominio

```go
// MergeMarketplaces aplica el collect-all de AG-D9: mergea por Nombre las filas detectadas
// (CC) con las declaradas (store del operador). Reglas, en este orden:
//  1. Un Nombre presente en los dos lados produce UNA fila con Eslabones = [cc-known, declarado].
//  2. Repo: si los dos lados canonicalizan al MISMO valor, se usa. Si difieren, se usa el
//     DECLARADO (el operador es la autoridad de lo que él registró) y la divergencia queda en
//     Discrepancias con LOS DOS valores crudos visibles — jamás se elige en silencio (BR-8).
//  3. Clase: CC no modela clase. La declarada manda; sin declaración, ClaseReferencia
//     (fail-safe: un marketplace que solo apareció por procedencia NO es «propio»).
//  4. InstallLocation/CCActualizado: solo del lado CC (el operador no los declara).
//  5. Orden de salida: por Nombre ascendente (estable, sin depender del orden del map).
// Es pura: no lee disco, no consulta red. `detectados` y `declarados` los traen los puertos.
func MergeMarketplaces(detectados, declarados []MarketplaceConocido) []MarketplaceConocido

// VersionDeEntrada aplica la precedencia de AG-D14/BR-2. `declarada` es plugins[].version tal
// como vino (puede ser ""). Devuelve ("","") cuando no hay dato — jamás una versión fabricada.
func VersionDeEntrada(declarada string, src SourceCatalogo) (version string, de ProcedenciaVersion)

// EsSemver reporta si s parsea como MAJOR.MINOR.PATCH con pre-release/build opcionales
// (subconjunto laxo de semver 2.0.0, suficiente para comparar versiones de plugin). Sin
// dependencia externa: el repo no tiene golang.org/x/mod y no se agrega una por esto.
func EsSemver(s string) bool

// CompararSemver ordena a vs b como semver (−1/0/+1). ok=false si alguno no parsea — el caller
// NUNCA cae a comparación de strings (eso es justo el pass fabricado que BR-4 mata en deriva).
func CompararSemver(a, b string) (cmp int, ok bool)

// AgruparCanales puebla ComparteSourceCon en su: dos entradas con el mismo Source.Crudo son
// canales del mismo arnés físico (AG-D11), y cada una lista a las otras. Muta su in-place y lo
// devuelve para poder encadenar. Orden de los nombres: como vienen en el catálogo (estable).
func AgruparCanales(su []EntradaCatalogo) []EntradaCatalogo
```

### 3.2 · `internal/domain/marketplace_situacion.go` (nuevo)

```go
// TipoSituacion son las 6 ramas de spec §4.3. La 6ta (`no-comparable`) es de PRIMERA CLASE:
// sin ella el sistema tendría que elegir entre mentir (al-hilo por defecto) o callarse (BR-9).
type TipoSituacion string

const (
	SituacionNoLoTengo         TipoSituacion = "no-lo-tengo"
	SituacionAlHilo            TipoSituacion = "al-hilo"
	SituacionMiCopiaAdelantada TipoSituacion = "mi-copia-adelantada"
	SituacionEstanteAdelantado TipoSituacion = "estante-adelantado"
	SituacionEnDeriva          TipoSituacion = "instalaciones-en-deriva"
	SituacionNoComparable      TipoSituacion = "no-comparable"
)

// SituacionCatalogo es el veredicto del cruce catálogo × Portafolio para UNA fila.
// Invariante: Tipo==SituacionNoComparable ⟹ Motivo≠"" (una rama muda no informa nada).
type SituacionCatalogo struct {
	Tipo    TipoSituacion `json:"tipo"`
	Mia     string        `json:"mia,omitempty"`     // versión del canónico propio.
	Estante string        `json:"estante,omitempty"` // versión de la fila del catálogo.
	Cuantas int           `json:"cuantas,omitempty"` // instalaciones en-deriva.
	Motivo  string        `json:"motivo,omitempty"`
	// ClavePortafolio es la Identidad.Clave() de la entrada cruzada ("" si no-lo-tengo) — el FE
	// la usa para el link «ver ficha» sin re-derivar el slug.
	ClavePortafolio string `json:"clave_portafolio,omitempty"`
	// Via dice CÓMO se cruzó (BR-3, trazabilidad): "home-declarado" (identidad.home == el repo
	// del marketplace) | "faceta-registry" (identidad provisional, pero el registry de la copia
	// resuelve a este marketplace) | "rename" (cruzó por nombre_anterior). La UI muestra el
	// cruce débil ("faceta-registry") como tal — no lo presenta como identidad resuelta.
	Via string `json:"via,omitempty"`
}

// Accion es el verbo que la fila del catálogo ofrece (AG-D8 decisión 6). El vocabulario ya
// existe: `traer-canonico` es el `↧ Traer canónico` del drawer (portafolio-drawer.tsx) —
// dos puertas, un acto, cero vocabulario nuevo (mockups/INDEX.md regla dura 4).
type Accion string

const (
	AccionNinguna           Accion = ""
	AccionTraerCanonico     Accion = "traer-canonico"
	AccionPublicar          Accion = "publicar"
	AccionActualizarMiCopia Accion = "actualizar-mi-copia"
	AccionReparar           Accion = "reparar"
)

// AccionCatalogo es el verbo + si está habilitado + POR QUÉ no. En ESTE paquete Habilitada es
// SIEMPRE false (spec §0 + BR-10 + C2 de design.md): el entregable es la situación honesta, no
// la ejecución. El motivo es el tooltip LITERAL — vive acá, en el dominio, para que exista un
// solo texto testeable y para que ningún widget pueda pintar un botón habilitado por su cuenta.
type AccionCatalogo struct {
	Verbo      Accion `json:"verbo"`
	Habilitada bool   `json:"habilitada"`
	Motivo     string `json:"motivo,omitempty"`
}

// CalcularSituacion cruza UNA fila del catálogo contra las entradas del Portafolio que le
// corresponden. `coincidencias` viene de CruzarConPortafolio (0, 1 o N). Tabla de verdad y
// precedencia completas en design.md §6.1 — el orden de los `case` ES la precedencia.
func CalcularSituacion(fila EntradaCatalogo, coincidencias []CoincidenciaPortafolio) SituacionCatalogo

// CoincidenciaPortafolio es una entrada del Portafolio que cruzó con una fila del catálogo,
// con el CÓMO anotado.
type CoincidenciaPortafolio struct {
	Entrada EntradaPortafolio
	Via     string // home-declarado | faceta-registry | rename
}

// CruzarConPortafolio busca en `todas` las entradas que corresponden a (repoMarketplace, fila).
// Tres vías, en orden de fuerza (§6.2):
//  1. home-declarado: identidad.Home == repoMarketplace ∧ identidad.ID == fila.Nombre.
//  2. faceta-registry: identidad.Home == "" (provisional) ∧ identidad.ID == fila.Nombre ∧
//     repoMarketplace ∈ RegistriesDe(entrada).
//  3. rename: idem 1 o 2 pero con identidad.ID ∈ fila.NombreAnterior.
// Devuelve TODAS las coincidencias — el caller nunca elige una en silencio (C-ID-2).
func CruzarConPortafolio(repoMarketplace string, fila EntradaCatalogo, todas []EntradaPortafolio) []CoincidenciaPortafolio

// RegistriesDe es el equivalente Go del selector FE homónimo: unión de entrada.Registries ∪
// cada instalaciones[].Origen.Registry, canonicalizada y dedupeada, orden de primera aparición.
// Un crudo que no canonicaliza se conserva tal cual (S1-D3: visible, nunca descartado).
func RegistriesDe(e EntradaPortafolio) []string

// AccionDeSituacion mapea (situación, clase) → verbo + habilitación + tooltip literal.
// ES EL PUNTO DE ENFORCEMENT de BR-1 y del boundary marketplace-referencia-es-solo-procedencia:
// para ClaseReferencia devuelve Habilitada:false en TODA rama, sin excepción.
func AccionDeSituacion(s SituacionCatalogo, clase ClaseMarketplace) AccionCatalogo
```

### 3.3 · Extensión de `internal/domain/portafolio.go`

Un solo campo nuevo, **aditivo** (`omitempty` ⇒ todo `portafolio.json` existente sigue parseando
sin migración):

```go
// En EntradaPortafolio:

	// OrigenSinResolverDesde: RFC3339 de cuándo el operador CONFIRMÓ dejar esta entrada sin
	// origen (S7, opción «ninguno — dejarlo sin origen», BR-11/E-26). Distinto de "" , que
	// significa «todavía nadie lo miró»: el contador cruzado «N sin origen resuelto» cuenta
	// las provisionales con este campo VACÍO. Confirmar «ninguno» no inventa un home — solo
	// deja de reclamar atención.
	OrigenSinResolverDesde string `json:"origen_sin_resolver_desde,omitempty"`
```

### 3.4 · `internal/ports/marketplace.go` (nuevo)

```go
package ports

// MarketplaceStore persiste el lado DECLARADO del registro (los que el operador registró por el
// wizard). Degrada honesto igual que PortafolioStore: Listar jamás falla por una fila corrupta.
type MarketplaceStore interface {
	Listar() ([]domain.MarketplaceConocido, []domain.EntradaCorrupta)
	Upsert(m domain.MarketplaceConocido) error
	Olvidar(nombre string) (bool, error)
}

// MarketplaceDetector expone el lado DETECTADO: `~/.claude/plugins/known_marketplaces.json`.
// READ-ONLY sobre un árbol ajeno. Metadata ilegible ⇒ (nil, error) — el usecase lo convierte
// en una discrepancia VISIBLE, nunca en una lista vacía silenciosa.
type MarketplaceDetector interface {
	Detectados() ([]domain.MarketplaceConocido, error)
}

// CatalogoReader lee el `marketplace.json` de UN marketplace. Hay dos implementaciones y el
// usecase decide el orden (política en el usecase, mecanismo en el adapter): local primero
// (checkout de CC, cero red, offline) y remoto como fallback (gh api). `Fuente` se anota en
// EstadoLectura.Fuente para trazabilidad (BR-3).
type CatalogoReader interface {
	Leer(ctx context.Context, m domain.MarketplaceConocido) (domain.Catalogo, error)
	Fuente() string // "local" | "remoto"
}

// CatalogoValidador prueba una URL ANTES de registrar (S6, BR-5/G3): busca
// `.claude-plugin/marketplace.json` legible y devuelve lo que LEYÓ. Nunca persiste nada.
// Es una interfaz aparte de CatalogoReader porque el insumo es una URL, no un
// MarketplaceConocido que todavía no existe.
type CatalogoValidador interface {
	Validar(ctx context.Context, url string) (domain.Catalogo, error)
}

// CatalogoCache guarda la última lectura EXITOSA por marketplace + su timestamp (AG-D8
// decisión 3, BR-3). Leer devuelve ok=false tanto si no hay caché como si el archivo está
// corrupto — con `motivo` poblado en el segundo caso, para que el degradado sea visible.
type CatalogoCache interface {
	Leer(nombre string) (cat domain.Catalogo, motivo string, ok bool)
	Guardar(nombre string, cat domain.Catalogo) error
	Olvidar(nombre string) error
}
```

> **No confundir con lo existente:** `ports.ArnesRegistry` (arnés→cwd de sesión),
> `ports.PortafolioStore` (identidades `(home,id)`), `ports.RepoConfigStore` (repo del
> self-update). Ningún nombre colisiona.

### 3.5 · `internal/usecase/marketplace.go` (nuevo)

```go
// MarketplaceService orquesta el plano Marketplaces y el catálogo a través de puertos —
// cero I/O acá. `remoto` puede ser nil (sin `gh` ni PAT): el servicio degrada honesto en vez
// de fallar, y la fila dice `sin-acceso` con motivo.
type MarketplaceService struct {
	store      ports.MarketplaceStore
	detector   ports.MarketplaceDetector
	local      ports.CatalogoReader
	remoto     ports.CatalogoReader
	validador  ports.CatalogoValidador
	cache      ports.CatalogoCache
	portafolio ports.PortafolioStore
	ahora      func() time.Time // inyectable: los tests no dependen del reloj.
}

func NewMarketplaceService(
	store ports.MarketplaceStore,
	detector ports.MarketplaceDetector,
	local, remoto ports.CatalogoReader,
	validador ports.CatalogoValidador,
	cache ports.CatalogoCache,
	portafolio ports.PortafolioStore,
) *MarketplaceService
```

Errores centinela (mismo patrón que `ErrObservarClaveNoEncontrada`, para que el handler mapee
el status por `errors.Is` y no por parsing de strings):

```go
var (
	// ErrMarketplaceNoConocido: el nombre no está ni detectado ni declarado → 404.
	ErrMarketplaceNoConocido = errors.New("marketplace: nombre no conocido")
	// ErrMarketplaceYaRegistrado: BR-7 — no se duplica ni se pisa → 409.
	ErrMarketplaceYaRegistrado = errors.New("marketplace: ya registrado (no se duplica ni se pisa)")
	// ErrURLNoCanonicalizable: la url no parsea a host/owner/repo → 400.
	ErrURLNoCanonicalizable = errors.New("marketplace: url no resuelve a host/owner/repo")
	// ErrNoEsMarketplace: el repo existe pero no tiene `.claude-plugin/marketplace.json`
	// legible → 400 (E-17: mensaje explícito, no un 500 genérico).
	ErrNoEsMarketplace = errors.New("marketplace: el repo no expone .claude-plugin/marketplace.json legible")
	// ErrSinViaDeLectura: no hay checkout local ni vía remota (gh ausente/no autenticado y sin
	// PAT) → 503. Distinto de ErrNoEsMarketplace: no es que la url esté mal, es que NO PUEDO MIRAR.
	ErrSinViaDeLectura = errors.New("marketplace: sin vía de lectura (ni checkout local ni gh/PAT)")
	// ErrClaseInvalida: clase distinta de propio|referencia → 400.
	ErrClaseInvalida = errors.New("marketplace: clase inválida (propio|referencia)")
)
```

Métodos:

```go
// Listar arma el plano Marketplaces (S2): collect-all detector+store → MergeMarketplaces →
// hidrata Lectura de cada fila desde el caché (sin tocar red ni parsear catálogos: el caché
// existe para que este GET sea O(1) lecturas por fila, AG-D16). Cuenta también las entradas
// del Portafolio sin origen resuelto (el contador cruzado de AG-D8 decisión 7).
// Un detector que falla NO aborta: la lista sale con las declaradas + una discrepancia global.
func (s *MarketplaceService) Listar(ctx context.Context) (ListadoMarketplaces, error)

// ListadoMarketplaces es el wire del plano. Corruptas y AvisoDetector son VISIBLES: un registro
// dañado o una metadata de CC ilegible nunca se oculta ni impide listar el resto (BR-11).
type ListadoMarketplaces struct {
	Marketplaces      []domain.MarketplaceConocido `json:"marketplaces"`
	Corruptas         []corruptaWire               `json:"corruptas,omitempty"`
	AvisoDetector     string                       `json:"aviso_detector,omitempty"`
	SinOrigenResuelto int                          `json:"sin_origen_resuelto"`
}

// Catalogo devuelve el catálogo de `nombre` con la situación ya calculada por fila.
// refrescar=false → caché primero; si no hay caché (o está corrupto), hace UNA lectura y la
// persiste. refrescar=true → lectura fresca obligatoria (AG-D8 decisión 3: refresco explícito).
// Si la lectura falla y HAY caché: devuelve el caché + Lectura degradada con motivo + el
// `Cuando` viejo (BR-3+BR-4). Si falla y NO hay caché: Entradas=nil (null en el wire) +
// Lectura con motivo. JAMÁS un `[]` fabricado, JAMÁS un 500.
func (s *MarketplaceService) Catalogo(ctx context.Context, nombre string, refrescar bool) (domain.Catalogo, error)

// Validar prueba una url SIN persistir nada (S6 paso 1, BR-5): canonicaliza → si algún
// marketplace conocido ya apunta a ese repo usa su checkout local (cero red) → si no, remoto.
// Devuelve el Catalogo leído (con Entradas reales) + si el nombre ya está registrado.
func (s *MarketplaceService) Validar(ctx context.Context, url string) (Validacion, error)

// Validacion es lo que S6 pinta en el ✓: datos LEÍDOS del archivo real, nunca inferidos.
type Validacion struct {
	URLCanonica  string           `json:"url_canonica"`
	Nombre       string           `json:"nombre"`
	OwnerNombre  string           `json:"owner_nombre,omitempty"`
	OwnerEmail   string           `json:"owner_email,omitempty"`
	OwnerURL     string           `json:"owner_url,omitempty"`
	Descripcion  string           `json:"descripcion,omitempty"`
	Entradas     int              `json:"entradas"`
	Fuente       string           `json:"fuente"` // local | remoto
	YaRegistrado bool             `json:"ya_registrado"`
	ClaseActual  ClaseMarketplace `json:"clase_actual,omitempty"` // solo si YaRegistrado.
}

// Registrar persiste el lado declarado (S6 paso 4). Re-valida (TOCTOU-safe, mismo criterio que
// AgregarProyecto) y devuelve la fila mergeada lista para aterrizar en el catálogo.
// Nombre ya registrado → ErrMarketplaceYaRegistrado (BR-7): no duplica, NO PISA.
// NO clona, NO instala, NO escribe fuera de `~/.arnesia/` (boundary §11.3).
func (s *MarketplaceService) Registrar(ctx context.Context, url string, clase domain.ClaseMarketplace) (domain.MarketplaceConocido, error)

// Olvidar quita el lado DECLARADO. Si CC igual lo conoce, la fila sigue apareciendo con
// eslabón `cc-known-marketplaces` solo (y clase degradada a `referencia`, fail-safe) — es lo
// honesto: no podemos hacer que Claude Code deje de conocerlo.
func (s *MarketplaceService) Olvidar(ctx context.Context, nombre string) (bool, error)

// CandidatosDeOrigen alimenta el selector de S7: los marketplaces conocidos ordenados por
// señales BLANDAS respecto de la entrada `clave` (PENDIENTE-01 §3). Orden:
//  1. el repo ya está en RegistriesDe(entrada) — la señal más fuerte que existe sin manifiesto;
//  2. coincidencia de owner del repo con el `author`/`empresas` de la entrada;
//  3. clase `propio` antes que `referencia`;
//  4. resto por nombre.
// NADA premarcado: el orden es sugerencia, la elección es del operador (BR-11).
func (s *MarketplaceService) CandidatosDeOrigen(ctx context.Context, clave string) ([]CandidatoOrigen, error)

type CandidatoOrigen struct {
	Nombre string                  `json:"nombre"`
	Repo   string                  `json:"repo,omitempty"`
	Clase  domain.ClaseMarketplace `json:"clase"`
	// Senal es el texto que la fila muestra debajo del nombre ("autor coincide · propio",
	// "propio · no leído aún") — se ARMA acá, no en el widget, para que sea testeable.
	Senal string `json:"senal,omitempty"`
}
```

### 3.6 · Extensión de `internal/usecase/portafolio.go`

```go
// ErrAsignarOrigenColisiona: re-keyear la entrada a (home,id) daría una clave que YA existe.
// Fusionar dos identidades es una acción EXPLÍCITA fuera de alcance (C-ID-2, boundary
// portafolio-identidad-y-deriva-honesta §1) — el sistema no las junta solo.
var ErrAsignarOrigenColisiona = errors.New("portafolio: ya existe una entrada con esa identidad (fusionar es otra operación)")

// AsignarOrigen escribe el `home` DECLARADO de la entrada `clave` (S7, BR-11) y la re-keyea.
// NO clona, NO instala, NO escribe ningún archivo fuera de `~/.arnesia/` (C3 de design.md: el
// dir de una instalación `referenciada-cc` vive bajo `~/.claude`, que validarRootPortafolio
// PROHÍBE escribir — por eso el home va al store, no al sello).
//
// Con home=="" ⇒ el operador eligió «ninguno — dejarlo sin origen»: NO se toca la identidad,
// solo se estampa OrigenSinResolverDesde (deja de contar en «N sin origen resuelto») — jamás
// se inventa un home (E-26).
//
// Con home≠"": canonicaliza (400 si no resuelve) → arma IdentidadArnes{Home:canon, ID:id} →
// si la clave nueva existe y NO es la misma entrada ⇒ ErrAsignarOrigenColisiona → Upsert bajo
// la clave nueva + Desvincular la vieja → anota en CADA instalación el eslabón
// {Fuente:"declarado-por-operador", Campo:"home", Valor:canon} y re-corre ResolverOrigen para
// que la discrepancia home≠registry (si la hay) quede visible (BR-3/C-OR-6) →
// re-evalúa deriva con el home nuevo (ahora RutaReferencia SÍ puede resolver: es el efecto
// útil de reconciliar, y aparece solo, sin un botón nuevo).
func (s *PortafolioService) AsignarOrigen(ctx context.Context, clave, home string) (domain.EntradaPortafolio, error)

// SinOrigenResuelto cuenta las entradas con identidad provisional (Home=="") y
// OrigenSinResolverDesde=="" — el contador cruzado de AG-D8 decisión 7.
func (s *PortafolioService) SinOrigenResuelto(ctx context.Context) (int, error)
```

### 3.7 · Adaptadores — `internal/adapters/marketplace/` (paquete nuevo)

| archivo | tipo exportado | satisface | mecanismo |
|---|---|---|---|
| `store.go` | `Store` | `ports.MarketplaceStore` | JSON atómico `~/.arnesia/marketplaces.json`, envelope versionado entry-wise (calco de `portafolio.Store`) |
| `cache.go` | `Cache` | `ports.CatalogoCache` | JSON atómico `~/.arnesia/catalogos/<slug>.json`, uno por marketplace |
| `detector.go` | `DetectorCC` | `ports.MarketplaceDetector` | lee `~/.claude/plugins/known_marketplaces.json` READ-ONLY |
| `catalogo_local.go` | `LectorLocal` | `ports.CatalogoReader` (`Fuente()=="local"`) | `<InstallLocation>/.claude-plugin/marketplace.json` + `catalogo.json` opcional |
| `catalogo_remoto.go` | `LectorRemoto` | `ports.CatalogoReader` + `ports.CatalogoValidador` (`Fuente()=="remoto"`) | `gh api` (shell-out), PAT como fallback |
| `parse.go` | — (privado) | — | traducción del shape ajeno → `domain.*` (anti-corruption layer) |

Campos inyectables para tests (mismo patrón que `portafolio.Scanner.CCPluginsDir`):

```go
type DetectorCC struct{ CCPluginsDir string } // default ~/.claude/plugins
type Store struct{ /* path string */ }        // NewStore(path string) — "" ⇒ ~/.arnesia/marketplaces.json
type Cache struct{ /* dir string */ }         // NewCache(dir string) — "" ⇒ ~/.arnesia/catalogos
type LectorLocal struct{}                     // sin estado: la ruta viene en el MarketplaceConocido
type LectorRemoto struct {
	GHBin      string        // default "gh" resuelto por PATH; inyectable
	Token      string        // PAT fallback (ARNESIA_GH_TOKEN); "" = no usar
	Timeout    time.Duration // default 10 s
	MaxBytes   int64         // default 8 MiB (§5.3)
	HTTPClient *http.Client  // solo la rama PAT; nil ⇒ default con Timeout
}
```

### 3.8 · Transporte — `internal/adapters/transport/http/marketplace.go` (nuevo)

Handlers, uno por endpoint (§7), con la misma disciplina que `portafolio.go`: `decodeJSON` para
el body, `writeJSON` para todo, `errorBody{Error}` para los errores, `errors.Is` para el status.
Un tipo de body nuevo (para que el FE pueda ofrecer «ir a él» en el 409 sin parsear texto):

```go
// conflictoMarketplaceBody — 409 de POST /api/marketplaces (BR-7): además del motivo, el
// nombre existente, para que la UI ofrezca navegar en vez de solo mostrar un error.
type conflictoMarketplaceBody struct {
	Error  string `json:"error"`
	Nombre string `json:"nombre"`
}
```

### 3.9 · Composition root — `cmd/arnesia/main.go`

Nueva factoría al lado de `newPortafolioService` (línea ~507), y cableado en `runServe`
**después** de `portafolioSvc` (necesita el mismo store del Portafolio):

```go
// newMarketplaceService cablea el plano Marketplaces a sus 7 puertos. `remoto` se cablea
// SIEMPRE (el lector decide en runtime si tiene gh/PAT y degrada honesto) — un nil acá
// convertiría «no puedo leer» en «no existe la función», que es peor.
func newMarketplaceService(pfStore ports.PortafolioStore) (*usecase.MarketplaceService, error) {
	st, err := marketplace.NewStore("")
	if err != nil {
		return nil, err
	}
	cache, err := marketplace.NewCache("")
	if err != nil {
		return nil, err
	}
	remoto := &marketplace.LectorRemoto{Token: os.Getenv("ARNESIA_GH_TOKEN")}
	return usecase.NewMarketplaceService(
		st, &marketplace.DetectorCC{}, &marketplace.LectorLocal{}, remoto, remoto, cache, pfStore,
	), nil
}
```

`newPortafolioService` debe **devolver también el store** (o exponerlo) para poder compartirlo:
cambiar su firma a `(*usecase.PortafolioService, ports.PortafolioStore, error)` y ajustar los dos
callers (`runServe` línea ~266 y el subcomando CLI línea ~543). `NewHandler` gana un parámetro
`marketplaces *usecase.MarketplaceService` **al lado de `portafolio`** (no al final: agrupado con
su hermano, la firma ya es larga y el orden semántico importa más que el diff).

`docs/architecture/fitness/.go-arch-lint.yml` gana:

```yaml
components:
  marketplace:
    in: internal/adapters/marketplace/**
deps:
  marketplace:
    mayDependOn: [domain, ports]   # registro + caché + lector de catálogo (este paquete)
  cmd:
    mayDependOn: [..., portafolio, marketplace, history, dogfood, doctrina]
```

---

## 4 · Persistencia

### 4.1 · Registro de marketplaces declarados → `~/.arnesia/marketplaces.json`

**JSON atómico (temp+rename), envelope versionado, decodificación entry-wise.** Calco exacto de
`internal/adapters/portafolio/store.go`.

```json
{
  "version": 1,
  "marketplaces": [
    {
      "nombre": "prenter-marketplace",
      "repo": "github.com/alpacapurpura/prenter-marketplace",
      "clase": "propio",
      "eslabones": ["declarado-por-operador"],
      "registrado": "2026-07-25T14:02:11Z"
    }
  ]
}
```

**Por qué acá y no en SQLite:**

1. **El `.db` es desechable por doctrina.** `indice-desechable-jsonl-es-verdad` v1.3: en mismatch
   de `schema_meta.version` el índice hace `wipeFile` (borra `.db`, `-wal`, `-shm`) y re-indexa
   desde `ports.ArnesRegistry`. Un registro de marketplaces **no** se puede reconstruir de
   `ArnesRegistry`: la clase `propio`/`referencia` es una **declaración del operador que no vive
   en ningún otro lado**. Ponerlo ahí sería garantizar su pérdida silenciosa en el próximo bump
   de esquema. La doctrina no está mal: el dato no calza en ese contenedor.
2. **Precedente exacto y probado.** El registro del Portafolio —el dato hermano, misma criticidad—
   ya vive así y su degradado honesto está `enforced`
   (`portafolio-identidad-y-deriva-honesta` check `store-degrada-honesto`).
3. **Es chico y auditable a mano.** 5-20 filas. Que el operador pueda abrirlo con un editor y ver
   qué declaró es parte de la honestidad del repo.
4. **Cero superficie de esquema nueva.** Ninguna tabla, ninguna migración, ningún `ALTER`.

**Degradado ante corrupción** (idéntico a `portafolio.Store`, y por la misma razón):

| daño | comportamiento |
|---|---|
| archivo ausente | store vacío, sin error — el plano nace igual poblado por el detector CC |
| envelope ilegible (JSON roto) | 1 `EntradaCorrupta{Raw: todo el archivo, Motivo:"envelope ilegible: …"}`; el store abre; el plano lista **solo lo detectado** + banner de corrupta |
| una fila ilegible | esa fila → `EntradaCorrupta`; las demás se cargan |
| fila sin `nombre` | `EntradaCorrupta{Motivo:"fila sin nombre: no hay clave de merge"}` — nunca se le inventa un nombre |
| error de I/O real (permisos) | `NewStore` devuelve error → el daemon **no arranca el servicio de marketplaces** pero sigue vivo; el endpoint responde 500 con el motivo real (`ErrSinRegistro`), jamás una lista vacía |
| corruptas al reescribir | se re-serializan crudas junto a las sanas si son JSON válido (mismo criterio y misma limitación honesta que `portafolio.Store.saveLocked`) |

### 4.2 · Caché de catálogo → `~/.arnesia/catalogos/<slug(nombre)>.json`

**Un archivo por marketplace**, JSON atómico, `slug` vía `domain.Slug` (el mismo del Portafolio:
`[a-z0-9-_]`, runs colapsados) — así un nombre con `@`, espacios o unicode no puede escapar del
directorio ni colisionar de forma sorpresiva. Colisión de slug entre dos nombres distintos
(`mi mkt` y `mi-mkt`) se resuelve con sufijo de huella: `<slug>~<HuellaPath(nombre)[:8]>.json`
cuando el archivo existente declara otro `nombre` adentro. **Nunca se pisa el caché de otro.**

```json
{
  "version": 1,
  "leido_en": "2026-07-25T14:07:33Z",
  "fuente": "local",
  "catalogo": { "…": "domain.Catalogo serializado" }
}
```

**Por qué existe (y no es opcional):** el plano Marketplaces necesita `N entradas` y `leído hace
<t>` **por fila**. Sin caché, cada `GET /api/marketplaces` tendría que parsear los 5
`marketplace.json` — uno de ellos **159 KB / 273 entradas** (AG-D16). El caché convierte el listado
en O(1) lecturas chicas por fila. El valor «offline» es real pero secundario: para los
marketplaces que CC ya clonó, el checkout local **es** la fuente offline.

**Por qué un archivo por marketplace y no uno solo:** (a) el radio de daño de una corrupción es un
marketplace, no todos; (b) refrescar uno no reescribe ~200 KB; (c) `Olvidar` es un `os.Remove`.

**Degradado:**

| daño | comportamiento |
|---|---|
| archivo ausente | `Leer` → `ok=false, motivo=""` ⇒ la fila dice `no-leido` (nunca «0 entradas») |
| JSON corrupto | `ok=false, motivo="caché de catálogo ilegible: …"` ⇒ fila `no-leido` **con el motivo visible**; el próximo refresco lo sobreescribe |
| `version` desconocida | `ok=false, motivo="caché de versión N desconocida (se re-leerá)"` — el caché SÍ se descarta y se reconstruye: acá el dato **es** derivable, a diferencia del registro |
| `catalogo.entradas == null` guardado | nunca se guarda: `Guardar` **rechaza** un `Catalogo` con `Entradas==nil` (`ErrCacheSinEntradas`) — cachear un «no sé» convertiría un fallo transitorio en un estado persistente |
| dir sin permiso de escritura | `Guardar` devuelve error; el servicio **igual responde** el catálogo leído + `Lectura.Motivo` gana el sufijo «(no se pudo cachear: …)» — la lectura no se pierde por no poder guardarla |

### 4.3 · Concurrencia

- `Store` y `Cache`: `sync.Mutex` por instancia, escritura temp+rename (atómica a nivel POSIX).
  Dos lecturas simultáneas del mismo catálogo son seguras (el rename es atómico: se ve el archivo
  viejo o el nuevo, nunca uno partido).
- **Dos refrescos simultáneos del mismo marketplace** (dos clicks, dos pestañas): el último gana
  y ambos responden un catálogo consistente. Se acepta a propósito — un `singleflight` agrega
  maquinaria para ahorrar una lectura de 159 KB o un `gh api`. Queda anotado como deuda con razón
  en el `PARIDAD.md`, y `plan-pruebas.md` E-40 lo verifica (no corrompe, no mezcla filas).

---

## 5 · Lectura de catálogo

### 5.1 · Política (vive en el usecase, no en el adapter)

```
Catalogo(nombre, refrescar):
  m ← fila mergeada de `nombre`            (404 si no está)
  si !refrescar:
      cat, motivo, ok ← cache.Leer(nombre)
      si ok:  return cat con Lectura del caché            ← camino normal, cero I/O grande
      (si !ok y motivo≠"" ⇒ se arrastra al Lectura.Motivo del intento que sigue)
  cat, err ← leerFresco(ctx, m)
  si err == nil:
      cat.Entradas ← AgruparCanales(...)  ; cruzar con Portafolio (§6) ; cache.Guardar
      return cat con Lectura{leido, ahora, len(Entradas), fuente}
  // falló la lectura fresca
  viejo, _, hay ← cache.Leer(nombre)
  si hay:  return viejo con Lectura{tipoDe(err), Cuando: viejo.leido_en, Entradas: len(viejo), Motivo: err, Fuente: viejo.fuente}
  return Catalogo{Entradas: nil, Lectura{tipoDe(err), Motivo: err}}        ← null, jamás []

leerFresco(ctx, m):
  si m.InstallLocation ≠ "" ∧ existe <InstallLocation>/.claude-plugin/marketplace.json:
      return local.Leer(ctx, m)                      ← barato, offline, cero red (AG-D9)
  si remoto ≠ nil:  return remoto.Leer(ctx, m)
  return ErrSinViaDeLectura
```

`tipoDe(err)`: `ErrNoEsMarketplace` / JSON inválido → `url-no-resuelve`; `ErrSinViaDeLectura`,
`gh` no autenticado, 401/403/404 → `sin-acceso`; timeout/red → `sin-acceso` con el motivo real.
**Cualquiera de los dos exige `Motivo≠""`.**

### 5.2 · `LectorLocal`

1. `<InstallLocation>/.claude-plugin/marketplace.json` — la ruta **exacta** de AG-D10 punto 1.
2. Parseo tolerante (§5.4).
3. Enriquecimiento opcional: `<InstallLocation>/catalogo.json`. Si existe y parsea → `Canales` +
   `Versiones` + `EstadoCanal` por fila (cruzando `EntradaCatalogo.Version` con
   `versiones[].version`). Si **no** existe, o no parsea, o `marketplace` de adentro no coincide
   con el nombre → **degrada sin ruido** (AG-D11 sub-hallazgo: es convención de prenter, no
   estándar; un marketplace ajeno no tiene por qué tenerlo). Un `catalogo.json` corrupto **sí**
   deja un aviso en `Lectura.Motivo` con el sufijo «(catálogo.json ignorado: …)» — degradar sin
   ruido ≠ ocultar un archivo roto que el operador puso ahí a propósito.
4. **Nunca escribe** en el checkout ajeno (enforcer del boundary §11.3).
5. `EvalSymlinks` sobre `InstallLocation` antes de leer, y si el resultado no existe o no es dir:
   `ErrNoEsMarketplace` con motivo «installLocation ya no existe en disco: `<ruta>`» (E-35).
6. **Solo en la lectura local** (ya estamos en disco, el `stat` es barato): para cada fila con
   `Source.Tipo == SourceRutaRelativa`, un `os.Stat(<InstallLocation>/<ruta>)`; si falta, la fila
   gana `Aviso: ["el catálogo apunta a <ruta> pero ese dir no existe en el checkout"]` — la fila
   **se muestra igual** (el catálogo declara lo que declara; no somos nosotros los que la
   borramos), con el problema del dato a la vista (E-34). En la lectura **remota no se verifica**:
   exigiría N requests para recorrer el árbol ajeno, y eso sí sería la app haciendo de proxy.

### 5.3 · `LectorRemoto` — `gh api`, no clone

**Decisión: `gh api`.** Mecanismo:

```
gh api "repos/<owner>/<repo>/contents/.claude-plugin/marketplace.json" \
   --jq '.content'            # base64 → decode → parse
```

y, best-effort para el enriquecimiento, `…/contents/catalogo.json` (404 = ausente, normal).
`owner/repo` sale de `domain.CanonicalizarRepo` (que devuelve `host/owner/repo`): si el host **no**
es `github.com` ⇒ `ErrSinViaDeLectura` con motivo «solo se sabe leer catálogos de github.com por
ahora» — honesto y explícito, no un fallo genérico.

**Por qué `gh api` y no un clone shallow:**

| criterio | `gh api` | clone shallow |
|---|---|---|
| red | 1-2 requests, ~160 KB peor caso | árbol completo (`claude-plugins-official` trae `plugins/` con cientos de dirs) |
| disco | cero escrituras | temp dir, cleanup, ventana de fallo a mitad |
| privados | funciona con la auth propia del operador | idem, pero además necesita credential helper para git |
| duplicación | ninguna | re-clona lo que CC ya clona en `installLocation` |
| auth-terms | **la app es conductor**: invoca el CLI YA autenticado del operador, nunca ve ni guarda un token | idem |

Coherencia con boundaries: `superficie-local-confinada` gobierna **quién le habla al daemon** y
**qué config ve una sesión spawneada** — no prohíbe lecturas salientes. Lo que sí impone es que
el crédito nunca lo tenga la app: shell-out a `gh` cumple eso literalmente. Precedente en el
árbol: `internal/adapters/selfupdate/updater.go` y `internal/adapters/publish/publisher.go` ya
hacen shell-out a `git`. `vision.md` §Ecosistema: «la app es conductor, no proxy».

**Fallback PAT** (auth-terms firmado, `spec.md §8`): si `gh` no está en PATH o
`gh auth status` falla, y `ARNESIA_GH_TOKEN` está poblado → `GET
https://api.github.com/repos/<owner>/<repo>/contents/.claude-plugin/marketplace.json` con
`Authorization: token <PAT>` + `Accept: application/vnd.github+json`. El **selector de
credencial** es una función pura y se testea sin ningún login real:

```go
// TipoCredencial es la vía de acceso elegida. El orden es doctrina (auth-terms): el CLI del
// operador primero — la app JAMÁS proxya un login ni guarda una credencial propia.
type TipoCredencial string
const (CredencialGH TipoCredencial = "gh"; CredencialPAT TipoCredencial = "pat"; CredencialNinguna TipoCredencial = "")

// elegirCredencial es pura: decide con los hechos que el caller ya averiguó.
func elegirCredencial(ghDisponible, ghAutenticado bool, pat string) (TipoCredencial, string /*motivo si Ninguna*/)
```

**Guardas duras del lector remoto (todas obligatorias):**

| guarda | valor | por qué |
|---|---|---|
| `context` con timeout | 10 s default, inyectable | un `gh` colgado no cuelga el daemon (E-41) |
| techo de bytes | 8 MiB (`io.LimitReader`) | el oficial son 159 KB; un archivo hostil no puede OOMear el daemon |
| techo de entradas | 5 000 (`Truncado` = descartadas) | `Truncado>0` viaja al wire y **se muestra** — recorte VISIBLE, nunca silencioso |
| host | solo `github.com` | no se inventa una API para un host desconocido |
| salida de `gh` | stderr capturado y **usado como `Motivo`** | el motivo real llega al operador, no un «error genérico» |
| exit code ≠ 0 | mapea a `ErrNoEsMarketplace` (404 del path) o `sin-acceso` (401/403) según stderr | E-16/E-17 exigen mensajes distinguibles |

### 5.4 · Parseo tolerante (`parse.go`) — el anti-corruption layer

El shape ajeno se traduce; el dominio **no** lo adopta (§11.3). Reglas:

| campo ajeno | tratamiento |
|---|---|
| `name` (raíz) | requerido. Ausente ⇒ `ErrNoEsMarketplace` con motivo «marketplace.json sin `name`» |
| `owner` | `{name?, email?, url?}` — los tres opcionales, se copian tal cual (AG-D15) |
| `description` (raíz) | se prefiere; si falta, `metadata.description`; si faltan las dos, `""` |
| `plugins` | **ausente** ⇒ `Entradas=nil` + `Lectura.Motivo="marketplace.json sin la clave plugins"` (E-31). **`[]`** ⇒ `Entradas=[]` + `leido` con `Entradas:0` (E-32) — leí y no declara ninguno |
| `plugins[].name` | requerido por fila. Fila sin `name` ⇒ **se descarta la fila** y se suma a `Truncado`, con el motivo «N fila(s) sin `name`» en `Lectura.Motivo` — visible, no silencioso |
| `plugins[].source` | `json.RawMessage` → intenta `string`, luego objeto; ninguno ⇒ `SourceDesconocido` con `Crudo` = el JSON crudo recortado a 120 chars (visible, nunca descartado) |
| `plugins[].version` | `*string` (para distinguir ausente de `null` de `""`) |
| `renames` (raíz) | `map[string]string` viejo→nuevo; se invierte y se puebla `NombreAnterior` de la fila destino |
| resto (`strict`, `skills`, `lspServers`, `homepage`, `author`, `displayName`, `keywords`, `category`, `tags`, `$schema`) | **se ignoran**. Que el dominio no crezca por cada campo ajeno es el punto del boundary |
| duplicados de `name` | dos filas con el mismo `name` en un catálogo: **las dos se conservan**, las dos ganan `Aviso: ["nombre duplicado en el catálogo: X"]` y las dos fuerzan `no-comparable` (fila 4 de §6.1) — el catálogo ajeno está mal, y eso se muestra en vez de elegir una (E-37) |

---

## 6 · Cálculo de situación

Vive en `internal/domain/marketplace_situacion.go`. **Puro**: sin I/O, sin reloj, sin transporte.
Es el entregable central del paquete (`spec.md §0`: «lo que este paquete SÍ entrega es el cálculo
honesto de la situación»).

### 6.1 · Tabla de verdad de `CalcularSituacion` — el orden de las filas ES la precedencia

Notación: `E` = la coincidencia elegida (§6.2); `vMia` = `E.Canonico.Version`;
`vEstante` = `fila.Version`; `insts` = `E.Instalaciones`.

| # | condición evaluada | resultado | motivo (literal) |
|---|---|---|---|
| 1 | `len(coincidencias) == 0` | `no-lo-tengo` | — |
| 2 | `len(coincidencias) > 1` | `no-comparable` | `"N entradas de tu portafolio coinciden con esta fila (<claves>): resolvé el origen para desambiguar"` |
| 3 | `fila.Source.Tipo == SourceDesconocido` | `no-comparable` | `"el catálogo declara un `source` que no reconozco: <crudo>"` |
| 4 | nombre duplicado en el catálogo | `no-comparable` | `"nombre duplicado en el catálogo: <nombre>"` |
| 5 | `∃ i ∈ insts : i.Deriva == en-deriva` | `instalaciones-en-deriva`, `Cuantas = n` | — |
| 6 | `vEstante == ""` | `no-comparable` | `"el catálogo no declara versión de esta entrada ni se puede derivar de su source"` |
| 7 | `E.Canonico == nil` | `no-comparable` | `"no tenés canónico de este arnés (solo instalaciones read-only): no hay copia editable que comparar contra el estante"` |
| 8 | `vMia == ""` | `no-comparable` | `"tu canónico no declara versión"` |
| 9 | `!EsSemver(vMia) ∨ !EsSemver(vEstante)` | `no-comparable` | `"versiones no comparables (no-semver): «<vMia>» vs «<vEstante>»"` |
| 10 | `CompararSemver(vMia, vEstante) > 0` | `mi-copia-adelantada`, `Mia`, `Estante` | — |
| 11 | `CompararSemver(vMia, vEstante) < 0` | `estante-adelantado`, `Mia`, `Estante` | — |
| 12 | iguales ∧ `∃ i : i.Deriva == deriva-no-evaluable` | `no-comparable` | `"versión igual al estante, pero N instalación(es) con deriva no evaluable: <detalles>"` |
| 13 | iguales ∧ `∃ i : i.Aviso ≠ ""` | `no-comparable` | `"versión igual al estante, pero N instalación(es) con aviso: <avisos>"` |
| 14 | iguales ∧ (`len(insts)==0` ∨ todas `al-hilo`) | `al-hilo` | — |

Todas las ramas pueblan `ClavePortafolio` y `Via` cuando hay exactamente una coincidencia.

**Lectura de la precedencia, para que nadie la re-derive mal:**

- `instalaciones-en-deriva` (5) va **antes** que cualquier comparación de versión porque es un
  hecho **duro** de hash (no le falta ningún insumo) y es lo más accionable — así lo ordena
  `spec.md §4.3`.
- `no-comparable` (6-9, 12, 13) va **antes** de toda afirmación positiva. `al-hilo` (14) es la
  única afirmación positiva del set y es **la última fila**: solo se llega ahí cuando *ningún*
  insumo falta y *ninguna* instalación tiene señal. Eso es literalmente «no sé» ≠ «sano».
- Las filas 2-4 son degradados de **integridad del insumo** y ganan a todo lo demás: si no
  sabemos de qué entrada hablamos, cualquier veredicto sería inventado.

### 6.2 · Cruce de identidad — por qué `(home,id)` a secas NO alcanza

`spec.md §4.3` dice cruzar «por identidad `(home, id)` donde `home` = nombre del marketplace».
**Contra los datos reales eso falla en el caso dominante** y hay que decirlo:

`harness@prenter-marketplace` está instalado en `luana-vitalia`, pero el dir de instalación
(`~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2`) **no tiene `arnes.l0.json`**, así que
`ResolverIdentidad` deja la identidad **provisional**: `Home==""`, clave `sin-home~harness~…`.
Un cruce estricto por `identidad.Home == repo` **no encontraría nada** y toda la columna diría
`no-lo-tengo` — que es exactamente el pass fabricado al revés.

Pero `origen.Registry` **sí** resolvió el repo (vía `cc-plugins` × `known_marketplaces.json`), y
`entrada.Registries` lo persiste. De ahí las **tres vías, en orden de fuerza**:

| vía | condición | fuerza | qué muestra la UI |
|---|---|---|---|
| `home-declarado` | `identidad.Home == repoMarketplace ∧ identidad.ID == fila.Nombre` | fuerte: el autor lo declaró en el manifiesto | nada extra |
| `faceta-registry` | `identidad.Home == "" ∧ identidad.ID == fila.Nombre ∧ repoMarketplace ∈ RegistriesDe(entrada)` | media: la adquisición lo dice, el manifiesto no | chip «origen no declarado — cruzado por el registry de la copia» + la fila entra en el contador «sin origen resuelto» |
| `rename` | igual que 1 o 2 pero con `identidad.ID ∈ fila.NombreAnterior` | media | chip «el catálogo renombró `<viejo>` → `<nuevo>`» |

**Nunca se elige entre N coincidencias** (fila 2 de la tabla): se dice cuántas y se manda a
reconciliar. Ese es el circuito completo — el cruce débil es justamente lo que S7 existe para
convertir en `home-declarado`.

Comparación de repos: `CanonicalizarRepo` de los dos lados **antes** de comparar (RN-IDENT-1: un
`owner/repo` corto y una url `https://…​.git` del mismo repo colapsan).

### 6.3 · `AccionDeSituacion` — tooltips literales, un solo lugar

| clase | situación | `Verbo` | `Habilitada` | `Motivo` (tooltip LITERAL) |
|---|---|---|---|---|
| `referencia` | `no-lo-tengo` | `traer-canonico` | `false` | `no aplica: solo arneses propios` |
| `referencia` | cualquier otra | `` (ninguna) | `false` | `no aplica: solo arneses propios` |
| `propio` | `no-lo-tengo` | `traer-canonico` | **`true`** | `` ← **única celda habilitada** (AG-D17, §13) |
| `propio` | `al-hilo` | `` | `false` | `` (la UI pinta «nada que hacer · ver ficha») |
| `propio` | `mi-copia-adelantada` | `publicar` | `false` | `Publicar se construye en su propio paquete (ítem 3 del outcome)` |
| `propio` | `estante-adelantado` | `actualizar-mi-copia` | `false` | `Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)` |
| `propio` | `instalaciones-en-deriva` | `reparar` | `false` | `Reparar se construye en su propio paquete (ítem 5 del outcome)` |
| `propio` | `no-comparable` | `` | `false` | `= Situacion.Motivo` (se propaga, no se duplica) |

El literal de la fila `referencia` es **el del mockup firmado** (`disabled title="no aplica: solo
arneses propios"`, la única fila del dibujo que trae el par correcto) — se conserva tal cual.

> **Habilitada es `false` en 7 de las 8 filas.** Eso no es un placeholder: es el alcance de
> `spec.md §0` hecho contrato de tipo. La excepción es `propio × no-lo-tengo`, que **AG-D17 habilita**
> (mecanismo en §13) — y la asimetría es la que el operador firmó: `Traer` escribe solo en
> `~/.arnesia/checkouts/`, mientras Publicar/Actualizar/Reparar escriben en cosas del cliente o en el
> estante remoto. Cuando llegue el paquete del ítem 3, cambia **una celda más** de esta tabla y su
> caso de test — ningún widget.

---

## 7 · Endpoints HTTP

Base `/api` (igual que el resto). Todos bajo `withAuth` (Host+Origin+token). Se montan en
`router.go` **debajo del bloque del Portafolio**, con su propio comentario de bloque.

```go
// Plano Marketplaces + catálogo (este paquete): el estante de lo que vendemos y el espejo de
// si el cliente coincide (AG-D8). Lectura de catálogo CACHEADA con refresco explícito (BR-3).
mux.HandleFunc("GET /api/marketplaces", listMarketplaces(marketplaces))
mux.HandleFunc("POST /api/marketplaces", postRegistrarMarketplace(marketplaces))
mux.HandleFunc("POST /api/marketplaces/validaciones", postValidarMarketplace(marketplaces))
mux.HandleFunc("DELETE /api/marketplaces/{nombre}", deleteOlvidarMarketplace(marketplaces))
mux.HandleFunc("GET /api/marketplaces/{nombre}/catalogo", getCatalogoMarketplace(marketplaces))
mux.HandleFunc("POST /api/marketplaces/{nombre}/lecturas", postLeerCatalogo(marketplaces))
// Reconciliación de origen (S7): actúa sobre un ARNÉS, por eso vive bajo /portafolio.
mux.HandleFunc("GET /api/portafolio/arneses/{clave}/origen/candidatos", getCandidatosOrigen(marketplaces))
mux.HandleFunc("POST /api/portafolio/arneses/{clave}/origen", postAsignarOrigen(portafolio))
```

### 7.1 · `GET /api/marketplaces`

Respuesta 200:

```json
{
  "marketplaces": [
    {
      "nombre": "prenter-marketplace",
      "repo": "github.com/alpacapurpura/prenter-marketplace",
      "clase": "propio",
      "eslabones": ["cc-known-marketplaces", "declarado-por-operador"],
      "install_location": "/home/chalreme/.claude/plugins/marketplaces/prenter-marketplace",
      "cc_actualizado": "2026-07-10T00:36:43.459Z",
      "lectura": { "tipo": "leido", "cuando": "2026-07-25T14:07:33Z", "entradas": 2, "fuente": "local" },
      "registrado": "2026-07-25T14:02:11Z"
    },
    {
      "nombre": "caveman",
      "repo": "github.com/juliusbrussee/caveman",
      "clase": "referencia",
      "eslabones": ["cc-known-marketplaces"],
      "install_location": "/home/chalreme/.claude/plugins/marketplaces/caveman",
      "cc_actualizado": "2026-06-02T21:24:11.366Z",
      "lectura": { "tipo": "no-leido" }
    }
  ],
  "corruptas": [{ "motivo": "fila sin nombre: no hay clave de merge" }],
  "aviso_detector": "",
  "sin_origen_resuelto": 3
}
```

| status | cuándo | body |
|---|---|---|
| 200 | siempre que el servicio esté cableado — **incluso con detector ilegible o registro corrupto** (van en `aviso_detector` / `corruptas`) | arriba |
| 500 | el servicio no está cableado (`NewStore` falló al arrancar por I/O real) | `{"error":"…"}` |

`marketplaces` **nunca** es `null`: si no hay ninguno, `[]` — porque «no conozco ningún
marketplace» es una afirmación verdadera y verificable (a diferencia de un catálogo no leído).

### 7.2 · `POST /api/marketplaces/validaciones`

Request: `{"url":"https://github.com/alpacapurpura/prenter-marketplace"}`

| status | cuándo | body |
|---|---|---|
| 200 | se leyó un `marketplace.json` REAL | `usecase.Validacion` (§3.5) |
| 400 | `ErrURLNoCanonicalizable` · `ErrNoEsMarketplace` · JSON del catálogo inválido | `{"error":"<motivo textual>"}` |
| 503 | `ErrSinViaDeLectura` (ni checkout local ni `gh`/PAT) | `{"error":"<motivo>"}` |

**El 400 y el 503 son semánticamente distintos y el FE los usa distinto:** 400 = «tu url no sirve»
(E-16/E-17), 503 = «no puedo mirar» (nunca insinúa que la url esté mal). **Solo un 200 pinta ✓**
— eso es BR-5/G3 hecho contrato de transporte: no hay forma de fabricar un ✓ sin un archivo leído.

### 7.3 · `POST /api/marketplaces`

Request: `{"url":"…","clase":"propio"}`

| status | cuándo | body |
|---|---|---|
| 200 | registrado | `domain.MarketplaceConocido` ya mergeado |
| 400 | url no canonicalizable · `ErrClaseInvalida` · `ErrNoEsMarketplace` | `{"error":"…"}` |
| 409 | `ErrMarketplaceYaRegistrado` (BR-7: **no duplica ni pisa**) | `{"error":"…","nombre":"prenter-marketplace"}` |
| 503 | `ErrSinViaDeLectura` | `{"error":"…"}` |

El 409 lleva `nombre` para que S6 ofrezca «ya está registrado — ir a él» (E-18) sin parsear texto.

### 7.4 · `DELETE /api/marketplaces/{nombre}`

| status | cuándo | body |
|---|---|---|
| 200 | se quitó el lado declarado | `{"olvidado":true,"sigue_detectado":true}` |
| 404 | no había lado declarado | `{"error":"nombre no registrado por el operador"}` |

`sigue_detectado:true` es honestidad: **no podemos hacer que Claude Code deje de conocerlo**; la
fila va a seguir apareciendo con eslabón `cc-known-marketplaces` y clase `referencia`.

### 7.5 · `GET /api/marketplaces/{nombre}/catalogo` · `POST …/lecturas`

- `GET` = caché primero (primera visita hace UNA lectura y la persiste: idempotente).
- `POST …/lecturas` = **refresco explícito** (AG-D8 decisión 3). Es lo que pegan `Refrescar`,
  `Reintentar` y `Leer catálogo` del plano. Body vacío. Misma respuesta que el `GET`.

Respuesta 200 (los dos), `domain.Catalogo`:

```json
{
  "marketplace": "prenter-marketplace",
  "clase": "propio",
  "repo": "github.com/alpacapurpura/prenter-marketplace",
  "owner_nombre": "Prenter",
  "owner_email": "hola@alpacapurpura.lat",
  "lectura": { "tipo": "leido", "cuando": "2026-07-25T14:07:33Z", "entradas": 2, "fuente": "local" },
  "entradas": [
    {
      "nombre": "harness",
      "source": { "tipo": "ruta-relativa", "crudo": "./plugins/harness/0.5.3", "ruta": "./plugins/harness/0.5.3" },
      "version": "0.5.3",
      "version_de": "derivada-de-source",
      "descripcion": "Canal ESTABLE — …",
      "comparte_source_con": ["harness-beta"],
      "estado_canal": "habilitada",
      "situacion": { "tipo": "estante-adelantado", "mia": "0.5.2", "estante": "0.5.3", "clave_portafolio": "sin-home~harness~", "via": "faceta-registry" },
      "accion": { "verbo": "actualizar-mi-copia", "habilitada": false, "motivo": "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)" }
    }
  ],
  "canales": { "estable": "0.5.3", "beta": "0.5.3" },
  "versiones": [{ "version": "0.5.1", "estado": "deprecada", "fuente": "prenter-harness@v0.5.1", "fecha": "2026-07-02" }]
}
```

Respuesta 200 **degradada** (BR-4 en el cable — nótese `entradas: null`):

```json
{
  "marketplace": "vitalia-arneses",
  "clase": "propio",
  "lectura": { "tipo": "sin-acceso", "motivo": "gh: HTTP 404 — repo inexistente o sin acceso con la credencial actual" },
  "entradas": null
}
```

| status | cuándo |
|---|---|
| 200 | el marketplace es conocido — **incluso si no se pudo leer** (`entradas: null` + `lectura.motivo`) |
| 404 | `ErrMarketplaceNoConocido` |
| 500 | servicio no cableado |

**No hay 5xx por «no pude leer el catálogo».** Un 500 no le da al FE nada que renderizar y
empujaría a la UI a mostrar una lista vacía — el pass fabricado que BR-4 mata.

### 7.6 · `GET /api/portafolio/arneses/{clave}/origen/candidatos`

200: `{"candidatos":[{"nombre":"vitalia-arneses","repo":"github.com/vitalia/arneses","clase":"propio","senal":"el registry de tu copia ya apunta acá"}], "actual":""}`
· 404 clave desconocida. `actual` = el home ya declarado (`""` si provisional) para que el diálogo
pueda decir «ya tiene origen».

### 7.7 · `POST /api/portafolio/arneses/{clave}/origen`

Request: `{"home":"github.com/vitalia/arneses"}` **o** `{"sin_origen":true}`.

| status | cuándo | body |
|---|---|---|
| 200 | asignado (o «ninguno» confirmado) | la `EntradaPortafolio` re-keyed (con su `identidad` nueva → el FE re-apunta la fila/drawer, igual que `identificar`) |
| 400 | `home` no canonicalizable · ni `home` ni `sin_origen` en el body | `{"error":"…"}` |
| 404 | clave desconocida | `{"error":"…"}` |
| 409 | `ErrAsignarOrigenColisiona` | `{"error":"…"}` |

**No hay 2xx que clone o instale nada** (BR-11). El único efecto secundario es la re-evaluación de
deriva, que es una lectura.

### 7.8 · Contrato OpenAPI

`docs/architecture/contracts/api/openapi.yaml`: bump a `0.7.0-marketplaces` + un comentario de
versión en la cabecera (el archivo ya lleva ese changelog inline) + los 7 paths de arriba con el
mismo estilo denso de `summary` que usan los de `/portafolio`. **`entradas: null` se documenta
explícitamente** (`nullable: true` + la razón), porque es la parte del contrato que un
implementador «prolijo» convertiría en `[]` y rompería BR-4 sin darse cuenta.

---

## 8 · Frontend

### 8.1 · Árbol exacto de archivos

```
web/src/entities/marketplace/                        ← NUEVO (organismo de dominio)
  index.ts                       barrel público (todo import externo pasa por acá)
  model/types.ts                 espejo EXACTO del wire (§3.1) — hand-authored, como portafolio
  model/selectors.ts             PRESENTADORES puros (no calculan situación — C1)
  model/selectors.test.ts        tabla de casos (proyecto vitest `unit`, environment node)
  ui/chips.tsx                   ClaseChip · EstadoLecturaChip · SituacionChip · CanalChip · ViaChip
  ui/chips.stories.tsx           una story por estado, con play()
  testing/marketplaces.ts        fixtures con shape REAL (copiadas de la máquina, §8.5)

web/src/widgets/marketplace/                         ← NUEVO (chrome)
  index.ts
  ui/marketplace-list.tsx        S2 — filas + puerta cruzada + estados degradados
  ui/marketplace-list.stories.tsx
  ui/marketplace-catalogo.tsx    S3/S4 — buscador+filtro+lazy + acción por situación
  ui/marketplace-catalogo.stories.tsx

web/src/widgets/portafolio/ui/
  resolver-origen-dialog.tsx     S7 — NUEVO acá (actúa sobre un ARNÉS, no sobre un marketplace)
  resolver-origen-dialog.stories.tsx
  portafolio-wizard.tsx          MODIFICADO — Atrás/Cancelar/contador/ruta visible/rama Marketplace
  portafolio-wizard.stories.tsx  MODIFICADO — stories nuevas (§9)
  portafolio-list.tsx            MODIFICADO — GrupoControl (rótulos VER POR/FILTROS) + FiltroDisclosure importado
  portafolio-list.stories.tsx    MODIFICADO

web/src/shared/ui/                                   ← moléculas extraídas
  grupo-control.tsx (+ .stories.tsx)      rótulo visible + grupo de botones (AG-D4)
  lista-lazy.tsx (+ .stories.tsx)         centinela IntersectionObserver (AG-D3)
  buscador-filtro.tsx (+ .stories.tsx)    input search + N FiltroDisclosure
  filtro-disclosure.tsx (+ .stories.tsx)  PROMOVIDO de portafolio-list.tsx, sin cambio de conducta

web/src/pages/shell/ui/
  portafolio-view.tsx            MODIFICADO — conmutador de planos + TODO el transporte nuevo

web/src/shared/api/client.ts     MODIFICADO — 7 métodos genéricos <T> nuevos
web/src/app/styles/portafolio.css  MODIFICADO — clases nuevas, cero color literal
```

### 8.2 · Las moléculas de `shared/ui/` — la trampa que hay que evitar

**`shared/ui/**` NO PUEDE importar `entities/*`.** No es solo `ui-not-domain` (que prohíbe
`entities|features/*/model/`): es `shared-no-upward`, en `error`, que prohíbe
`^src/shared/` → `^src/(entities|features|widgets|pages|app)/` **entero**.

Consecuencia dura: las 4 moléculas son **genéricas y tontas**. El caller (widget) inyecta el
vocabulario de dominio como props.

```ts
// shared/ui/filtro-disclosure.tsx — PROMOCIÓN LITERAL de lo que hoy es privado en
// portafolio-list.tsx: mismo patrón sin-portal, aria-expanded en el botón, cero click-outside.
// Genérico <T extends string> + `labelDe` inyectado ⇒ no necesita conocer SaludPortafolio.
export function FiltroDisclosure<T extends string>(props: {
  etiqueta: string; abierto: boolean; onToggleAbierto: () => void; panelId: string
  valores: readonly T[]; labelDe: (v: T) => string
  seleccion: ReadonlySet<T>; onCambiar: (s: ReadonlySet<T>) => void
  vacio?: string | undefined
}): JSX.Element

// shared/ui/grupo-control.tsx — AG-D4: cierra L1. El rótulo pasa de `aria-label` invisible a
// TEXTO visible, sin perder el `role="group"`+`aria-label` que un lector de pantalla ya usaba.
export function GrupoControl(props: {
  rotulo: string                  // "Ver por" | "Filtros" — mayúsculas por CSS, no por el dato
  ariaLabel: string               // el aria-label EXISTENTE se conserva (no se degrada a11y)
  children: React.ReactNode
  hint?: string | undefined       // L3 de la auditoría: el hint del mockup, opcional
}): JSX.Element

// shared/ui/lista-lazy.tsx — AG-D3 + AG-D16 (273 filas es el caso normal). Render incremental
// por centinela IntersectionObserver; NUNCA trunca en silencio (E-28): cuando quedan filas sin
// montar lo DICE y ofrece el botón de fallback (IntersectionObserver puede no disparar en un
// contenedor sin scroll, y en jsdom no existe — de ahí el botón, que también hace la story
// testeable sin polyfill).
export function ListaLazy<T>(props: {
  items: readonly T[]
  render: (item: T, i: number) => React.ReactNode
  claveDe: (item: T) => string
  paso?: number                   // default 20
  etiquetaRestantes?: (n: number) => string   // default: "… N más (se cargan al bajar)"
}): JSX.Element

// shared/ui/buscador-filtro.tsx — composición: un input search + N FiltroDisclosure ya armados
// por el caller. No sabe filtrar: solo compone y rotula (los selectores son del dominio).
export function BuscadorFiltro(props: {
  busqueda: string; onBusqueda: (q: string) => void
  placeholder: string; ariaLabel: string
  filtros?: React.ReactNode | undefined
}): JSX.Element
```

**Refactor de movimiento, no de conducta:** `portafolio-list.tsx` borra su `FiltroDisclosure`
privado y su `toggleEnSet` (que también sube a `shared/lib/toggle-en-set.ts` — `shared/lib` ya
está en el allowlist de R2) e importa desde `@/shared/ui/filtro-disclosure`. Las 2 stories que hoy
ejercitan el disclosure (`FiltroEstadoAcota`/`FiltroMarketplaceAcota`) **deben seguir pasando sin
tocarlas** — ése es el criterio de éxito del refactor.

### 8.3 · `entities/marketplace/model/selectors.ts` — presentadores, no calculadoras

```ts
// La situación LA CALCULA EL DOMINIO GO y viaja en el wire (C1). Acá solo se presenta.
export function etiquetaDeSituacion(s: SituacionCatalogo): string
export function tonoDeSituacion(s: SituacionCatalogo): SaludVisual   // "ok" | "atencion" | "sin-senal"
export function textoDeLectura(l: EstadoLectura, ahora: Date): string  // "leído hace 4 min" | "sin acceso — <motivo>"
export function esNavegable(m: MarketplaceConocido): boolean          // false si sin-acceso/url-no-resuelve/no-leido (spec §4.1.4)
export function accionDeFila(m: MarketplaceConocido): "ver-catalogo" | "leer-catalogo" | "reintentar"
export function filtrarEntradasCatalogo(es: EntradaCatalogo[], q: string): EntradaCatalogo[]
export function filtrarPorSituacion(es: EntradaCatalogo[], tipos: ReadonlySet<TipoSituacion>): EntradaCatalogo[]
export function situacionesDisponibles(es: EntradaCatalogo[]): TipoSituacion[]
export function ordenarMarketplaces(ms: MarketplaceConocido[]): MarketplaceConocido[]  // propios legibles → propios no legibles → referencia
```

`SaludVisual` es un **alias local** de `entities/marketplace` con los mismos 3 valores que
`SaludPortafolio` (no se importa: sería cross-import de entidades). Se documenta en el
doc-comment: *«mismos 3 valores que `SaludPortafolio` a propósito — el átomo `DotSaludPortafolio`
los consume; NO se importa el tipo de la otra entidad (steiger `fsd/no-cross-imports`)»*.
`tonoDeSituacion`: `al-hilo`→`ok`; `mi-copia-adelantada`/`estante-adelantado`/`instalaciones-en-deriva`→`atencion`;
`no-lo-tengo`/`no-comparable`→`sin-senal`. **`no-comparable` jamás cae en `ok`** (§1 del spec).

### 8.4 · Chips de `entities/marketplace/ui/chips.tsx`

| chip | ramas | tono |
|---|---|---|
| `ClaseChip` | `propio` · `de referencia` | descriptivo (`--secondary`), sin tono de salud |
| `EstadoLecturaChip` | `leído hace <t>` · `no leído aún` · `sin acceso` · `url no resuelve` | `no-leido` = `sin-senal` (transparente + dashed) · `sin-acceso`/`url-no-resuelve` = `--warn` (recuperable: autenticar), **jamás `--crit`** |
| `SituacionChip` | las 6 | vía `tonoDeSituacion` |
| `CanalChip` | `mismo contenido que <X>` | descriptivo (AG-D11) |
| `ViaChip` | `origen no declarado — cruzado por el registry de la copia` · `renombrado: <viejo> → <nuevo>` | `--warn-soft` con texto `--foreground` (§1 del spec: no agravar la deuda de `.text-warn`) |

`no-comparable` se pinta con **`DotSaludPortafolio salud="sin-senal"`** (átomo ya firmado, importado
del barrel de `entities/portafolio` por el **widget**, no por la entidad) + el motivo textual
completo, **nunca truncado** (mismo criterio que `AvisoChip`/BR-8).

### 8.5 · Fixtures con shape real

`entities/marketplace/testing/marketplaces.ts` — copiadas de la máquina, **no inventadas**:

| export | de dónde sale |
|---|---|
| `mkPrenterPropio` | `prenter-marketplace` real (repo, installLocation, lastUpdated, 2 entradas) |
| `mkOficialReferencia` | `claude-plugins-official` real — `entradas: 273`, `catalogo.json` ausente |
| `mkSinAcceso` | fila con `lectura:{tipo:"sin-acceso",motivo:"gh: HTTP 404 …"}` y `entradas: null` |
| `mkNoLeido` | `caveman` real, `lectura:{tipo:"no-leido"}` |
| `mkConDiscrepancia` | mismo `nombre`, `repo` distinto entre eslabones (BR-8) |
| `catPrenter` | las 2 entradas reales (`harness` + `harness-beta`, **mismo `source`** → `comparte_source_con`) + `catalogo.json` real (`0.5.1` deprecada) |
| `catOficialMuestra` | 25 filas reales del oficial, cubriendo las 4 formas de `source` y `version: null` |
| `catVacio` | `entradas: []` + `lectura.tipo:"leido"`, `entradas:0` (≠ `null`) |
| `catNull` | `entradas: null` + motivo — el fixture que prueba que la UI no dice «no tiene arneses» |
| `situacionesLas6` | una `EntradaCatalogo` por rama, con su `accion` (todas `habilitada:false`) |

### 8.6 · Orquestación — `pages/shell/ui/portafolio-view.tsx`

**La página es la única que hace transporte** (`fe-transporte-independiente`). Estado nuevo:

```ts
const [plano, setPlano] = useState<"arneses" | "marketplaces">("arneses")
const [marketplaces, setMarketplaces] = useState<MarketplaceConocido[]>([])
const [mkEstado, setMkEstado] = useState<"cargando" | "error" | "datos">("cargando")
const [mkError, setMkError] = useState<string>()
const [mkCorruptas, setMkCorruptas] = useState<EntradaCorrupta[]>([])
const [avisoDetector, setAvisoDetector] = useState<string>()
const [sinOrigenResuelto, setSinOrigenResuelto] = useState(0)
const [catalogoAbierto, setCatalogoAbierto] = useState<string>()   // nombre del marketplace
const [catalogo, setCatalogo] = useState<Catalogo>()
const [catEstado, setCatEstado] = useState<"cargando" | "error" | "datos">("cargando")
const [catError, setCatError] = useState<string>()
const [resolverPara, setResolverPara] = useState<string>()          // clave del arnés en S7
const [candidatosOrigen, setCandidatosOrigen] = useState<CandidatoOrigen[]>()
```

Reglas de orquestación:

1. **Lazy por plano.** `GET /api/marketplaces` se pide al entrar a `plano === "marketplaces"` la
   primera vez, no al montar la vista. `hash-state` guarda el plano (`#/portafolio?plano=marketplaces`)
   para que S6 pueda aterrizar ahí (AG-D8 decisión 8) y para que un reload no pierda el lugar.
2. **Aterrizaje de S6:** `Registrar y ver catálogo` → cierra el wizard → `setPlano("marketplaces")`
   → `setCatalogoAbierto(nombre)` → refetch de la lista + del catálogo. Es un solo callback,
   `onRegistrado(nombre: string)`.
3. **Contador cruzado:** `sin_origen_resuelto` del listado → botón en el plano Marketplaces que
   hace `setPlano("arneses")` + `setFiltroSinOrigen(true)`. Ese filtro es un **filtro nuevo de la
   lista** (`filtrarSinOrigen`, selector puro de `entities/portafolio`), no una lente.
4. **Un solo overlay a la vez.** El wizard, el drawer, el catálogo y S7 comparten el z-stack
   existente; S7 se abre **sobre** el drawer (el drawer es el contexto de la fila). Regla:
   abrir S7 no cierra el drawer; cerrar S7 devuelve el foco al botón que lo abrió.
5. **`AbortController` por cada fetch cancelable** (catálogo y validación son los lentos): mismo
   patrón que `onEscanear` hoy. Cerrar el catálogo aborta su fetch — cero efectos.
6. **Errores de escritura** (`registrar`, `asignar origen`): banner propio de la página cuando el
   contrato del widget no tenga slot, exactamente como se resolvió `observarError` (S1-D21).
   `resolver-origen-dialog.tsx` **sí** lleva slot `error?: string` en su contrato: nace nuevo, no
   hay razón para repetir el parche.

**Métodos nuevos en `shared/api/client.ts`** (genéricos `<T>`, domain-free — el mismo patrón que
`listPortafolio`; `shared/api` **no puede** importar `entities/*`):

```ts
listMarketplaces: <T = unknown>() => req<T>("/api/marketplaces"),
validarMarketplace: <T = unknown>(url: string, signal?: AbortSignal) =>
  req<T>("/api/marketplaces/validaciones", { method: "POST", body: JSON.stringify({ url }), ...(signal ? { signal } : {}) }),
registrarMarketplace: <T = unknown>(url: string, clase: string) =>
  req<T>("/api/marketplaces", { method: "POST", body: JSON.stringify({ url, clase }) }),
olvidarMarketplace: <T = unknown>(nombre: string) =>
  req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}`, { method: "DELETE" }),
catalogoDeMarketplace: <T = unknown>(nombre: string, signal?: AbortSignal) =>
  req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}/catalogo`, { ...(signal ? { signal } : {}) }),
leerCatalogo: <T = unknown>(nombre: string, signal?: AbortSignal) =>
  req<T>(`/api/marketplaces/${encodeURIComponent(nombre)}/lecturas`, { method: "POST", ...(signal ? { signal } : {}) }),
candidatosDeOrigen: <T = unknown>(clave: string) =>
  req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/origen/candidatos`),
asignarOrigen: <T = unknown>(clave: string, home: string | null) =>
  req<T>(`/api/portafolio/arneses/${encodeURIComponent(clave)}/origen`, {
    method: "POST", body: JSON.stringify(home === null ? { sin_origen: true } : { home }),
  }),
```

`ApiError.status` ya existe → el 409 de registrar y el 503 de validar se distinguen sin parsear
texto. **El FE debe usar `status`, no `message`**, para decidir el copy.

### 8.7 · Punto de cambio de AG-D11 — **CERRADO** (no se deduplica por `source`)

> **AG-D11 está FIRMADA 🧑‍⚖️** (2026-07-25): el operador eligió **una fila por canal** tras ver las
> dos opciones renderizadas. **La fila del catálogo es la entrada de índice y NO se deduplica por
> `source`.** Esta sección queda como registro del análisis y del costo que HABRÍA tenido la otra
> opción — **no** es una vía abierta. Nada de abajo se implementa.

Si se hubiera preferido **deduplicar por `source`** (una fila por arnés físico, canales como chips):

| capa | cambia? | qué |
|---|---|---|
| Go `domain`/`usecase`/`adapters`/HTTP | **NO** | `ComparteSourceCon` ya viene poblado; `Situacion`/`Accion` se calculan por-canal igual |
| `entities/marketplace/model/selectors.ts` | **SÍ** | +1 selector puro: `dedupPorSource(es): {principal: EntradaCatalogo; canales: string[]}[]` — la primera aparición es la principal, el resto son chips |
| `widgets/marketplace/ui/marketplace-catalogo.tsx` | **SÍ** | render: `dedupPorSource(entradas).map(...)` en vez de `entradas.map(...)`; `CanalChip` pasa de «mismo contenido que X» a «canales: estable, beta» |
| `Situacion` de la fila fusionada | **SÍ (regla nueva)** | con N canales, la fila muestra **la peor** situación por la precedencia de §6.1 y lista las otras en el tooltip — nunca la más optimista |
| stories | **SÍ** | +1 story `CatalogoDedupPorSource` |

Costo total: 1 selector + 1 `map` + 1 regla de agregación + 1 story. **Cero backend.** Por eso la
propuesta puede diseñarse ahora sin bloquear la firma.

### 8.8 · Gates de lint — verificado contra las reglas reales

| gate | riesgo | veredicto |
|---|---|---|
| `layer-direction` (error) | `entities/marketplace` → `widgets/pages/app` | ✅ no ocurre por diseño |
| `widgets-no-upward` (error) | `widgets/marketplace` → `pages/app` | ✅ |
| `shared-no-upward` (error) | las 4 moléculas → `entities/*` | ⚠️ **el riesgo #1**. Mitigación: genéricos + `labelDe`/`render` inyectados (§8.2). Si el implementador importa `SaludPortafolio` en `shared/ui`, `verify` falla. |
| `ui-not-domain` (error) | `shared/ui` → `entities/*/model` | ✅ misma mitigación |
| `no-deep-import` (warn) | `@/entities/marketplace/ui/chips` en vez del barrel | ⚠️ importar **siempre** por `@/entities/marketplace` |
| `domain-not-transport` (error) | `entities/marketplace/model` → `shared/api` | ✅ los tipos son hand-authored, el fetch vive en la página |
| `no-circular` (error) | `widgets/marketplace` ↔ `widgets/portafolio` | ⚠️ **no debe existir ningún import cruzado**: la página compone los dos. Hoy no hay ni uno entre widgets (verificado) → se agrega la regla `no-sibling-widget-imports` (§11.4) para que no aparezca |
| `steiger fsd/no-cross-imports` | `entities/marketplace` → `entities/portafolio` | ⚠️ **prohibido** (C1). El chip `DotSaludPortafolio` lo importa el **widget**, no la entidad |
| `steiger fsd/public-api` | `entities/marketplace` y `widgets/marketplace` sin `index.ts` | ⚠️ los dos `index.ts` son obligatorios |
| `stylelint strict-value` (error) | color literal en `portafolio.css` | ⚠️ cero hex/rgb en propiedades de color. Nota: el **mockup** sí tiene dos (`#fff` en `.pf-emblema` línea 145, `rgba(0,0,0,.55)` en `.scrim` línea 214) — **no se copian**; `portafolio.css` ya resuelve los dos casos con tokens |
| `capability_trace` R2 (error) | archivo nuevo sin capability | ⚠️ cada `.tsx`/`.ts` nuevo de producto debe estar en `pointers:` de algún capability **en el mismo commit**. `*.stories.tsx`, `index.ts`, `shared/ui/**`, `shared/lib/**` y `*_test.go`/`*.test.ts` ya están en el allowlist |

---

## 9 · Máquina de estados del wizard

`spec.md §4.4` (AG-D6): se **conserva** la máquina de 4 estados; se agregan `Atrás` y `Cancelar`.
`Atrás` retrocede un paso; `Cancelar` aborta el flujo entero. El mockup dibuja pasos simultáneos:
**se descarta a propósito** (AG-D6 lo dice explícitamente).

### 9.1 · Rama Proyecto

Estados (los 4 de hoy, sin agregar ninguno):

```
              ┌──────── Atrás ────────┐   ┌──── Atrás / «cambiar ruta» ────┐
              │                       │   │                                │
              ▼                       │   ▼                                │
        ┌──────────┐  Escanear  ┌───────────┐  ok   ┌────────────┐ Agregar N ┌───────────┐
  ──▶   │  fuente  │───────────▶│ escaneando│──────▶│ candidatos │──────────▶│ agregando │──▶ cierra+refetch
        └──────────┘            └───────────┘       └────────────┘           └───────────┘
             ▲                        │  abort/err        │  error 400            │ error 400
             └────────────────────────┘                   └───────────────────────┘
                                                              (vuelve a `candidatos` con `error`)
```

| estado | `Atrás` | `Cancelar` | ✕ / Esc | qué se preserva al salir del estado |
|---|---|---|---|---|
| `fuente` | **`disabled` VISIBLE** + `title="ya estás en el primer paso"` (literal del mockup; nunca oculto) | habilitado → cierra, cero efectos | permitidos | — |
| `escaneando` | habilitado → **aborta el fetch** y vuelve a `fuente` | habilitado → aborta + cierra | permitidos | `path` y `modo` intactos |
| `candidatos` | habilitado → vuelve a `fuente` **preservando `path`, `modo` y `elegidos`** | habilitado → cierra, cero efectos | permitidos | ídem |
| `agregando` | **`disabled`** + `title="agregando en curso — esperá a que termine"` | **`disabled`**, mismo title | **bloqueados** (S1-D19, ya implementado) | — |

Detalles que cierran W3/W4 de la auditoría:

- **`Atrás` desde `candidatos` NO borra la selección.** Volver a paso 1, corregir un typo de la
  ruta y re-escanear no debe castigar al operador con re-tildar 30 filas. Si el nuevo escaneo
  devuelve claves distintas, las que ya no existen se descartan del set al recibir los candidatos
  (`elegidos ∩ claves(candidatos)`) — sin aviso: no hay nada que reportar, la fila ya no está.
- **`cambiar ruta`** (spec §4.4) es **la misma transición** que `Atrás` desde `candidatos`, con
  otra puerta. Un solo handler; dos botones. (E-11 y E-12 asertan el mismo resultado a propósito.)
- **Ruta escaneada visible** en `candidatos`: `<span class="mono">{pathEscaneado}</span>` + el
  botón `cambiar ruta`. Hoy `PasoFuente` desaparece en la rama feliz — eso era el agujero W3.
- **Rótulo de paso + contador (W4):** `Paso 2 de 2 · escaneo — encontrados {candidatos.length} arneses`,
  literal del mockup. El botón sigue contando **elegidos** (`Agregar {elegidos.size} al portafolio`),
  que es otra cifra y a propósito.
- **Nada premarcado** (AG-D7): `elegidos` nace vacío, el botón nace `disabled`.
- **Buscador + filtro + lazy** sobre los hallazgos (AG-D3), reusando `BuscadorFiltro` + `ListaLazy`
  + el agrupamiento por subcarpeta que ya existe (`gruposCandidatosDe`, S1-D26). Con **un solo**
  grupo la lista queda plana, como hoy.

**Contrato del widget** — 2 props nuevas y ninguna quitada (BR-12, superset estricto):

```ts
export interface PortafolioWizardProps {
  // …todo lo de hoy, intacto…
  /** vuelve al paso anterior; la página aborta el fetch en vuelo si hay uno (AG-D6). */
  onAtras: () => void
  /** ruta ya escaneada, para mostrarla en `candidatos` (W3/W4). "" en `fuente`. */
  pathEscaneado?: string | undefined
  /** rama activa del wizard (AG-D8 decisión 8). Default "proyecto" ⇒ las stories viejas no cambian. */
  rama?: "proyecto" | "marketplace" | undefined
  onRama?: ((r: "proyecto" | "marketplace") => void) | undefined
  /** === rama Marketplace (§9.2) === */
  mkEstado?: "url" | "validando" | "validado" | "registrando" | undefined
  mkValidacion?: Validacion | undefined
  mkError?: string | undefined
  mkYaRegistrado?: { nombre: string } | undefined
  onValidarMarketplace?: ((url: string) => void) | undefined
  onRegistrarMarketplace?: ((url: string, clase: "propio" | "referencia") => void) | undefined
  onIrAMarketplace?: ((nombre: string) => void) | undefined
}
```

`Cancelar` reusa `onClose` (ya existe y ya hace «cero efectos» + abort). **No se agrega
`onCancelar`**: serían dos nombres para un acto (`mockups/INDEX.md` regla dura 4).

### 9.2 · Rama Marketplace — solo registrar

```
        ┌──────┐  Validar   ┌───────────┐  200   ┌──────────┐  Registrar  ┌─────────────┐
  ──▶   │ url  │───────────▶│ validando │───────▶│ validado │────────────▶│ registrando │──▶ cierra + aterriza en S3/S4
        └──────┘            └───────────┘        └──────────┘             └─────────────┘
           ▲   ▲                   │ 400/503          │ Atrás                    │ 400/503/409
           │   └───────────────────┘                  │                          │
           └──── Atrás ────────────────────────────────┘                         ▼
                                                                    vuelve a `validado` con mkError
                                                                    (409 ⇒ + botón «ir a él», E-18)
```

| estado | qué se ve | `Atrás` | `Cancelar` | primario |
|---|---|---|---|---|
| `url` | input git url + nota «se valida que exista `.claude-plugin/marketplace.json` legible; sin respuesta real no se pinta ningún ✓ (criterio G3)» | `disabled` + `title="un solo paso"` → **cambia**: ahora `Atrás` vuelve a la **tab Proyecto**, así que queda habilitado. El literal `"un solo paso"` del mockup se descarta (había un solo paso, ahora hay 4) | cierra | `Validar` (`disabled` con url vacía) |
| `validando` | spinner + `Validando…` | aborta → `url` | aborta + cierra | `disabled` |
| `validado` | ✅ **`name` + `owner.name` (+ email/url si vienen) + «N arneses en el catálogo» + `fuente: local\|remoto`** (C6/E-15) · radios `Propio`(default) / `De referencia` + la nota que explica la diferencia (literal del mockup) | → `url`, **preservando la url** | cierra | `Registrar y ver catálogo` |
| `registrando` | `disabled` | `disabled` | **bloqueado** (POST en vuelo, mismo criterio S1-D19) | `Registrando…` |

- **Sin `validado` no hay `Registrar`.** El botón no existe en `url`: no hay forma de registrar
  algo no validado. Eso es BR-5 hecho máquina de estados, no un `disabled` que se pueda saltear.
- **La clase arranca en `propio`** porque es el caso dominante, y la etiqueta explica la diferencia
  a la vista (spec §4.5.3) — no es un default silencioso.
- **409:** se muestra «ya está registrado» + botón `ir a él` → `onIrAMarketplace(nombre)` → la
  página cierra el modal, va al plano y abre ese catálogo. **No se pisa nada** (BR-7).
- **La tab `Marketplace` deja de estar `disabled`.** El `TOOLTIP_S2` de esa tab se borra; el de
  «Repositorio GitHub» dentro de la rama Proyecto **se queda** (sigue fuera de alcance, spec §0).

### 9.3 · S7 — `ResolverOrigenDialog`

No es un wizard: un paso, `role="dialog"` + `aria-modal`, focus trap con `trapTabKeyDown`
(patrón S1-D18, sin portal, para que `within(canvasElement)` funcione en las stories).

```ts
export interface ResolverOrigenDialogProps {
  abierto: boolean
  arnesId: string                       // para el título «Resolver origen — <id>»
  candidatos: CandidatoOrigen[]         // ya ordenados por el backend (§3.5) — el widget NO ordena
  actual?: string | undefined           // home ya declarado ⇒ el diálogo lo dice
  /** null = «ninguno — dejarlo sin origen». undefined = nada elegido ⇒ Confirmar disabled. */
  eleccion: string | null | undefined
  onElegir: (home: string | null) => void
  confirmando: boolean
  error?: string | undefined
  onConfirmar: () => void
  onClose: () => void
}
```

- **Nada premarcado**; `Confirmar origen` nace `disabled` con
  `title="elegí un destino primero"` (literal del mockup, que ya trae el fix de ese defecto).
- La opción **«ninguno — dejarlo sin origen»** es la **última** de la lista, con la señal
  `honesto, no se inventa un home` (literal del mockup).
- `eleccion === null` es un valor **elegido** (≠ `undefined`) — la distinción de tres estados es la
  que permite que «ninguno» habilite `Confirmar` sin premarcar nada.
- **`Cancelar` = cero efectos.** Confirmar escribe el `home` (o el sello de «sin origen») y **no
  clona ni instala nada** (BR-11).

---

## 10 · Tokens y CSS

**Cero token nuevo** (`spec.md §1`). Clases nuevas en `web/src/app/styles/portafolio.css`, todas
calificadas con el ancestro `.arnesia-portafolio` (como todas las existentes, para que el
decorador `Frame` de las stories rinda igual que la app — hallazgo §1 de la auditoría):

```
.pf-planos, .pf-plano-tab                      conmutador Arneses|Marketplaces
.mk-lista, .mk-fila, .mk-fila.inerte           filas del plano (calcan las clases del mockup)
.mk-id, .mk-meta, .mk-pendientes, .mk-alcance
.cat-lista, .cat-fila, .cat-situacion, .cat-accion
.cat-lazy                                      pie del centinela de ListaLazy
.pf-grupo-control, .pf-grupo-control-rotulo     GrupoControl (rótulo VISIBLE, AG-D4)
.pf-btn-mini                                   botón de acción de fila (del mockup)
.pf-chip-clase, .pf-chip-lectura, .pf-chip-canal, .pf-chip-via
.pf-wizard-mk-*                                rama Marketplace del wizard
.pf-resolver-*                                 S7
```

Reglas duras que el implementador debe respetar (§1 del spec):

1. **Cero color literal.** `stylelint strict-value` está en `error` sobre `/color$/`,
   `background-color`, `border-color`, `fill`, `stroke`. **No copiar** los dos literales del
   mockup (`#fff` en `.pf-emblema`, `rgba(0,0,0,.55)` en `.scrim`): `portafolio.css` y el overlay
   de la página ya los resuelven con tokens.
2. **`sin-senal` nunca reusa el color de `ok`.** Reusar la clase existente `.pf-dot-salud.sin-senal`
   (transparente + borde dashed `--muted-foreground`), no una clase nueva.
3. **`--warn` = «necesita acción», `--crit` reservado a fallo.** `sin-acceso` es `--warn`.
4. **No agravar la deuda de `.text-warn`** (contraste 4.5:1, abierta en `BACKLOG.md`): todo texto
   nuevo sobre `--warn-soft` usa `--foreground`, nunca `--warn`.
5. El rótulo de `GrupoControl` va en mayúsculas **por CSS** (`text-transform: uppercase`), no en el
   dato — igual que el mockup (cuyo DOM dice `"Ver por"`).

---

## 11 · Arquitectura as-code

### 11.1 · Capabilities creadas (`status: stub`, sin `valida:` ⇒ R4 consistente)

`cap_num` máximo actual = **CAP-101** (101 hojas). Las nuevas arrancan en **CAP-102**.

| archivo | cap_num | qué reclama |
|---|---|---|
| `docs/product/capabilities/portafolio/registrar-marketplace.yaml` | CAP-102 | registro collect-all (CC + declarado), merge por nombre, discrepancias visibles, clase fail-safe |
| `docs/product/capabilities/portafolio/leer-catalogo-marketplace.yaml` | CAP-103 | lectura local-primero / remoto por `gh`, caché con `leído hace`, degradado honesto (`entradas: null`), enriquecimiento opcional |
| `docs/product/capabilities/portafolio/situacion-de-catalogo.yaml` | CAP-104 | las 6 ramas + precedencia + cruce por 3 vías + `AccionDeSituacion` (BR-1/BR-10) |
| `docs/product/capabilities/portafolio/asignar-origen.yaml` | CAP-105 | S7 en el backend: re-key por home declarado, «ninguno» confirmado, colisión explícita |
| `docs/product/capabilities/http-sse/superficie-marketplaces.yaml` | CAP-106 | los 7 endpoints y su semántica de status (400 vs 503 vs 409) |
| `docs/product/capabilities/fe-portafolio/plano-marketplaces.yaml` | CAP-107 | S2: conmutador de planos, filas, estados degradados, contador cruzado |
| `docs/product/capabilities/fe-portafolio/catalogo-de-marketplace.yaml` | CAP-108 | S3/S4: buscador+filtro+lazy, situación por fila, read-only de `referencia` |
| `docs/product/capabilities/fe-portafolio/wizard-registrar-marketplace.yaml` | CAP-109 | S6: máquina `url→validando→validado→registrando`, ✓ solo con lectura real |
| `docs/product/capabilities/fe-portafolio/resolver-origen-dialogo.yaml` | CAP-110 | S7 en el FE: nada premarcado, «ninguno» como elección |

**Por qué `status: stub` y `pointers:` solo a archivos que YA existen:** R1
(`TestCapabilityPointersResolve` + `TestCapabilityPointerSymbolsResolve`) hace `os.Stat` de cada
ruta del bloque `pointers:` y resuelve el `#Símbolo` con `go/parser`. Un puntero a
`internal/domain/marketplace.go` **hoy rompería CI**. Cada hoja lleva, **después** de los punteros
vivos, un bloque comentado `# ── punteros a activar al construir ──` con la lista exacta (el
parser corta el bloque en la primera línea que no empieza con `- `, así que los comentados no se
recolectan). **El commit que crea cada archivo Go debe descomentar sus punteros en el mismo
commit** — eso es R2 (`cap-sin-huerfano`) y R3 (gate de `pre-commit`).

**Capabilities existentes a EXTENDER durante la implementación** (no se tocan ahora: cambiarlas
antes de que el código exista sería reclamar trabajo no hecho):

| hoja | qué agregar | entrada de `change_log` |
|---|---|---|
| `fe-portafolio/wizard-agregar-proyecto.yaml` | `Atrás`/`Cancelar` (AG-D6), contador de hallazgos + ruta visible (W3/W4), buscador+filtro+lazy (AG-D3), tab Marketplace habilitada | `type: extend`, `story_id: 2026-07-23-portafolio-agregar-marketplace` |
| `fe-portafolio/lista-del-portafolio.yaml` | `GrupoControl` con rótulos visibles (AG-D4/L1), `FiltroDisclosure` promovido a `shared/ui`, conmutador de planos, filtro «sin origen» | ídem |
| `portafolio/registrar-identidad.yaml` | campo `OrigenSinResolverDesde` + `AsignarOrigen` re-keyea por home declarado | ídem |

### 11.2 · Boundary EXTENDIDO — `portafolio-identidad-y-deriva-honesta.md` → v1.2

Mismo L1 («identidad y estado de integridad no admiten respuesta inventada cuando falta el dato»),
4 invariantes nuevos. **Los 4 enforcers todavía no existen** → sus celdas van con
`(pendiente — enforcer colocado, llega con la implementación)` **sin ningún token
`archivo_test.go:TestX`**: el parser del ruleset (`reGoTest`/`reTestName`) los levantaría e
intentaría correrlos, y un paquete inexistente daría FAIL en vez de `deferred`. `enforced_by:` del
frontmatter **no se toca** hasta que los tests existan. Nombres exactos que la implementación debe
usar (y recién entonces cablear):

| check nuevo | enforcer objetivo |
|---|---|
| `registro-marketplaces-collect-all` | `internal/domain/marketplace_test.go:TestMergeMarketplacesAnotaDiscrepancia` |
| `catalogo-nunca-vacio-fabricado` | `internal/usecase/marketplace_test.go:TestCatalogoIlegibleNoFabricaVacio` |
| `version-catalogo-nunca-inventada` | `internal/domain/marketplace_test.go:TestVersionCatalogoNuncaInventada` |
| `situacion-no-comparable-de-primera-clase` | `internal/domain/marketplace_situacion_test.go:TestSituacionNoComparableGana` |

4 → **8 checks**.

### 11.3 · Boundary NUEVO — `marketplace-referencia-es-solo-procedencia.md`

**Por qué uno nuevo y no otra extensión:** los 4 invariantes de arriba comparten el L1 de la
honestidad (*no inventar cuando falta el dato*). Éste es un L1 **distinto**: *un dato de un sistema
que no controlo entra traducido y read-only; adoptar su modelo corrompe el mío*. Es
bounded-context + anti-corruption layer, no honestidad de dato. Meterlo en el otro nodo
diluiría los dos.

- **L1** — Bounded Context (Fowler: «total unification of the domain model for a large system will
  not be feasible or cost-effective») + Anti-Corruption Layer (Evans, vía Azure Architecture
  Center: «the core purpose of an anti-corruption layer is to protect the domain model»). Fuentes
  **verificadas en vivo el 2026-07-25** (las dos respondieron y se leyó su contenido).
- **L2** — `vision.md` §Qué mutó («muere operar/cargar arneses de terceros») + `CLAUDE.md` («solo
  arneses propios») aterrizado en 4 invariantes:
  1. `domain.AccionDeSituacion` devuelve `Habilitada:false` para `ClaseReferencia` en **toda** rama.
  2. Una clase desconocida degrada a `referencia`, nunca a `propio` (`ClaseSegura`).
  3. `Registrar` no clona, no instala, no escribe fuera de `~/.arnesia/`; el lector de catálogo
     nunca escribe en `~/.claude/**` ni en el checkout del marketplace.
  4. El shape ajeno se traduce en el adapter: `internal/domain/marketplace.go` no menciona
     `git-subdir`, `$schema`, `installLocation`, `renames` ni ningún campo del formato de CC sin
     normalizar.
- **4 checks**, `status: proposed` (el código no existe; nace honesto, no `enforced`), enforcers
  objetivo nombrados en el archivo con la misma precaución del §11.2.

### 11.4 · Fitness

| archivo | cambio |
|---|---|
| `docs/architecture/fitness/.go-arch-lint.yml` | componente `marketplace` + `deps` + `cmd.mayDependOn` (§3.9) |
| `web/.dependency-cruiser.js` | **regla nueva** `no-sibling-widget-imports` (`error`), espejo de `no-sibling-feature-imports`. Verificado: hoy **cero** imports cruzados entre widgets → la regla nace verde |
| `docs/architecture/boundaries/fe-taxonomia-componentes.md` | v1.1 → v1.2, +1 check `no-sibling-widget-imports` (4 → 5) |
| `docs/architecture/INDEX.md` | 21 → **22 boundaries**; 92 → **101 checks** (+4 portafolio-identidad, +4 boundary nuevo, +1 fe-taxonomía), con la nota honesta de que los 9 nacen **declarados**, no corriendo |

### 11.5 · Contratos y convenciones

| archivo | cambio |
|---|---|
| `docs/architecture/contracts/api/openapi.yaml` | `0.6.0-portafolio-slice1` → `0.7.0-marketplaces`, +7 paths, `entradas` documentado `nullable: true` con su razón |
| `docs/architecture/conventions/*` | **sin cambios**. Se revisó: no hay convención nueva (naming, estilo Go/TS, commits, hooks) que este diseño introduzca |
| `mockups/INDEX.md` | fila del mockup del paquete: anotar los 3 supersets que el spec agrega sobre el dibujo (C5/C6/C7) para que el próximo lector no los lea como drift |

---

## 12 · Orden de construcción y riesgos

### 12.1 · Orden sugerido (cada paso deja el árbol verde)

| # | qué | por qué en ese lugar |
|---|---|---|
| T1 | `domain/marketplace.go` + `marketplace_situacion.go` + tests de tabla | el corazón; puro, sin puertos, testeable de una |
| T2 | `ports/marketplace.go` + extensión de `domain.EntradaPortafolio` | contratos antes de adaptadores |
| T3 | `adapters/marketplace/{parse,detector,catalogo_local}.go` + tests con fixtures reales | la vía barata y offline primero: E-01..E-06 ya pasan |
| T4 | `adapters/marketplace/{store,cache}.go` + tests de degradado | persistencia y sus 6 ramas de corrupción |
| T5 | `adapters/marketplace/catalogo_remoto.go` + test del selector de credencial | lo que depende de red va último del backend |
| T6 | `usecase/marketplace.go` + `PortafolioService.AsignarOrigen` + tests con fakes | la política, con los mecanismos ya disponibles |
| T7 | `transport/http/marketplace.go` + router + openapi | superficie observable; a partir de acá el E2E con `curl` es posible |
| T8 | `shared/ui/{filtro-disclosure,grupo-control,lista-lazy,buscador-filtro}` + stories | el refactor de movimiento, con las 2 stories viejas como red |
| T9 | `entities/marketplace` completo + `selectors.test.ts` + `chips.stories.tsx` | dominio FE |
| T10 | `widgets/marketplace/{marketplace-list,marketplace-catalogo}` + stories | S2/S3/S4 |
| T11 | `portafolio-wizard.tsx` (Atrás/Cancelar/contador/rama Marketplace) + stories | S5/S6 |
| T12 | `resolver-origen-dialog.tsx` + stories | S7 |
| T13 | `portafolio-list.tsx` (GrupoControl) + `portafolio-view.tsx` (planos + transporte) | orquestación final |
| T14 | `domain/traer.go` (`PlanificarTraer` · `RutaCanonico` · `DentroDeCheckouts`) + tests de tabla | **puro y sin I/O**: BR-13 y las precondiciones de BR-1/BR-14 se prueban antes de que exista un solo `os.` |
| T15 | `ports/traer.go` + `adapters/traer/copia.go` (set de exclusión · symlinks · permisos) + tests | la copia es lo que comparten los dos caminos; se construye una vez |
| T16 | `adapters/traer/local.go` (camino A) + tests con el fixture de prenter | sin red: E-76/E-78/E-79 pasan acá |
| T17 | `adapters/traer/externo.go` (camino B) + tests con shim de `git` + 1 test de red opt-in | lo que depende de red, último |
| T18 | `usecase/traer.go` (`Traer`: las 6 ramas de falla + registro + deriva) + tests con fakes | **la atomicidad de BR-15 se prueba acá, en un solo lugar** |
| T19 | `transport/http/marketplace.go` (`postTraerCanonico`) + router + openapi | superficie observable: E2E con `curl` posible |
| T20 | FE: `accion.habilitada` real + estado `trayendo`/`traído`/`falló` en `portafolio-view.tsx` + botón del drawer + stories | las dos puertas al mismo acto |
| T21 | E2E vivo de los escenarios de `plan-pruebas.md` + `PARIDAD.md` | gate humano |

### 12.2 · Riesgos concretos para quien implementa

1. **`shared/ui` no puede importar `entities/*`** (`shared-no-upward`, `error`). Si aparece un
   `import { SALUD_LABEL } from "@/entities/portafolio"` en `shared/ui/buscador-filtro.tsx`,
   `pnpm --dir web run verify` falla. Las 4 moléculas son genéricas; el vocabulario entra por prop.
2. **`entities/marketplace` no puede importar `entities/portafolio`** (`steiger
   fsd/no-cross-imports`, y `fsd` está dentro de `verify`). Por eso la situación se calcula en Go
   y el chip `DotSaludPortafolio` lo importa el **widget**.
3. **Capabilities y código shippean en el MISMO commit.** Un `.go`/`.tsx` nuevo sin `pointers:`
   rompe R2; un `pointers:` a un archivo que todavía no existe rompe R1. Los bloques comentados de
   §11.1 son la única forma de tener el diseño escrito hoy sin romper CI.
4. **`vitest-browser` NO corre en background** (Chromium no headless en este entorno). Para
   verificar: `pnpm --dir web run verify` (typecheck+biome+depcruise+steiger+stylelint) en
   background, y `pnpm --dir web test --project=storybook` **en sesión interactiva**.
5. **`bash scripts/estado.sh --check` rompe CI si el checkpoint queda stale**, y hay hook
   `pre-commit` que regenera cifras. No teclear cifras a mano en ningún `.md` — se generan.
6. **`entradas: []` vs `null`** es el corazón de BR-4 y el error más fácil de cometer: un
   `make([]EntradaCatalogo, 0)` «defensivo» en el lugar equivocado convierte «no pude leer» en
   «este marketplace no tiene arneses». En Go: `Entradas` **sin** `omitempty`, y nunca se
   inicializa el slice antes de saber que hubo lectura.
7. **`~/.claude` es un árbol AJENO y protegido.** Se lee, nunca se escribe (y
   `validarRootPortafolio` ya lo prohíbe explícitamente). Todo lo que ArnesIA persiste va a
   `~/.arnesia/`.
8. **`CanonicalizarRepo` devuelve `host/owner/repo`** (3 segmentos, minúsculas), no `owner/repo`.
   Comparar contra `known_marketplaces.json` (que trae `owner/repo` corto) exige canonicalizar
   **los dos lados** — es el bug clásico de este cruce.
9. **El catálogo oficial son 273 filas / 159 KB.** Cualquier `map()` sin lazy sobre eso se nota en
   una máquina real, y un `<option>` por fila en un `<select>` es inusable. `ListaLazy` no es
   opcional.
10. **`arnesia conformance --todo` va a crecer** en total y en `deferred` (10 checks nuevos
    declarados). Eso **no** es regresión; una regresión es un `fail` nuevo. Verificar por FAIL, no
    por el total.

**Riesgos específicos de `Traer` (§13):**

11. **El temporal DEBE vivir bajo `~/.arnesia/`, no en `/tmp`.** `os.Rename` entre filesystems falla
    con `EXDEV` y `/tmp` es `tmpfs` en muchas instalaciones. Es el error más fácil de cometer en
    toda la sección y rompe la atomicidad de BR-15 en la máquina del operador, no en el CI.
12. **El pin de camino B es el `sha`, NO el `ref`.** Verificado en vivo que en una entrada sana
    apuntan a commits distintos: implementar la letra del spec (clonar por `ref`, verificar `sha`)
    **aborta siempre**. Y `ref` está en 77/220 entradas; `sha`, en 220/220.
13. **`sparse-checkout --cone` arrastra los archivos de la raíz del repo.** El contenido canónico es
    `<tmp>/<path>`, jamás `<tmp>`: sin el staging, el canónico se lleva el `README.md` del repo
    ajeno y `.git`.
14. **El set de exclusión de la copia tiene que ser EXACTAMENTE el de `HashFormaPlugin`**
    (`.git/`, `.in_use`, `.orphaned_at`). Excluir menos ⇒ el canónico hereda el remote del
    marketplace (un `Publicar` futuro podría empujar al lugar equivocado). Excluir más ⇒ el hash no
    coincide y BR-17 reporta `en-deriva` sobre algo recién traído.
15. **`ports.DerivaEvaluator` tiene firma FIJA: no se toca.** Se reusa **la misma instancia** que
    `newPortafolioService` ya construye (`&portafolio.Referencias{}`); crear una segunda duplicaría
    la resolución de `RutaReferencia` y podrían divergir.
16. **La única rama con estado parcial posible es el fallo del `Upsert` (paso 9).** El dir queda y
    es VISIBLE (el próximo Traer da 409). **No** se borra: destruir contenido que costó red por un
    fallo de escritura del registro es peor. Cualquier otra rama deja cero residuo.
17. **El token del PAT nunca en `argv`.** `GIT_ASKPASS` a un script en el temporal (modo `0700`) que
    lee de su propio env. Un `-c http.extraHeader=...` o una URL con token quedan visibles en `ps`
    y en el reflog respectivamente.

---

## 13 · `↧ Traer canónico` (S8) — mecanismo

> Habilitado por **AG-D17 FIRMADA 🧑‍⚖️** (resuelve C2, que era un hueco del `spec.md`, no del
> diseño). Contrato funcional: `spec.md` §4.7 · BR-13..BR-18 · E-76..E-90.
> **Todo lo de esta sección se verificó contra la máquina real y contra GitHub en vivo** el
> 2026-07-25; los tres hallazgos que cambian el mecanismo respecto de la letra del spec están en
> §13.1 y documentados como contradicciones C16/C17/C18 en §2.

### 13.1 · Evidencia que cambia el mecanismo (verificada en vivo)

`spec.md §4.7` dice: «camino B = clone shallow con `ref` → si viene `sha`, se **verifica**». Contra
los 220 `source` objeto reales del catálogo oficial eso **no funciona**, y hay que decirlo antes de
escribir una línea:

| # | hecho verificado | cómo se verificó | consecuencia |
|---|---|---|---|
| 1 | **`sha` está en 220/220 `source` objeto** (100 %); **`ref` solo en 77/220** | conteo sobre `claude-plugins-official/.claude-plugin/marketplace.json` | el caso DOMINANTE de camino B es **`sha` sin `ref`** (143/220), no «clone por ref». Un mecanismo centrado en `ref` falla en 2 de cada 3 entradas |
| 2 | **`ref` y `sha` APUNTAN A COMMITS DISTINTOS** en una entrada sana | `42Crunch-AI/claude-plugins`: `ref: v1.5.5` → `faf5305385de8afe…` · `sha: 30287f5e3f122a64…` (los dos resueltos con `gh api …/commits/{x}`) | implementar la letra del spec (clonar por `ref`, verificar `sha`) **abortaría en la primera entrada real**. **El `sha` es la autoridad; el `ref` es una pista.** |
| 3 | **`sha` ES un commit git real** en las 3 formas de objeto | `gh api repos/{owner}/{repo}/commits/{sha}` devolvió el mismo sha para `git-subdir`, `url` y `github` | BR-16 («verificar el `sha`») **sí es implementable**, como identidad de commit — no hace falta inventar un digest de contenido |
| 4 | **`source: "github"` trae DOS hashes distintos** (`commit` **y** `sha`), los dos commits válidos del mismo repo, sin semántica documentada | `fullstorydev/fullstory-skills`: `commit 1ec5865e…` y `sha b20614e2…`, ambos resuelven | no se adivina cuál materializar. **Regla: gana `sha`** (el único campo presente en las 220 formas objeto; `commit` es un extra de la forma `github`) y la divergencia queda como `Aviso` VISIBLE |
| 5 | **137/220 `source` objeto NO traen `path`** | conteo | «extraer solo `path`» aplica cuando hay `path`; sin él, la raíz del repo **es** el arnés |
| 6 | **el mecanismo real funciona**: `git init` + `remote add` + `sparse-checkout --cone` + `fetch --depth 1 --filter=blob:none origin <sha>` + `checkout FETCH_HEAD` | corrido en vivo contra `42Crunch-AI/claude-plugins`: `HEAD == 30287f5e…` exacto, solo el subdir pedido materializado, **664 KB** totales | GitHub **sí** permite fetch por sha alcanzable. Es el mecanismo elegido (§13.6) |
| 7 | **cone mode arrastra los archivos de la RAÍZ del repo** además del subdir | el mismo test dejó un `README.md` de raíz junto a `plugins/` | el contenido canónico es `<tmp>/<path>`, **nunca** `<tmp>`: el staging (§13.5) no es opcional |
| 8 | **los checkouts de marketplace tienen `.git`** (caveman 424 K · ponytail 2,2 M · prenter 672 K) y **`source: "./"` existe de verdad** (caveman, ponytail) | `du -sh */.git` + lectura de sus `marketplace.json` | copiar «la subcarpeta» con `source: "./"` copiaría el repo del marketplace **con su `.git`**, dejando el canónico apuntando al remote del MARKETPLACE. Regla de exclusión obligatoria (§13.5) |

### 13.2 · Contratos Go — dominio

`internal/domain/traer.go` (nuevo). Puro: decide **qué** se materializa y **a dónde**; no toca disco.

```go
// PlanTraer es la decisión pura de CÓMO materializar una fila del catálogo: qué camino, de
// dónde, qué se verifica y a qué destino. Se calcula ANTES de tocar el disco para que la
// precondición, el destino y el rechazo sean testeables sin I/O (BR-13/BR-14 son decisiones,
// no efectos).
type PlanTraer struct {
	Camino  CaminoTraer `json:"camino"`
	Destino string      `json:"destino"` // absoluto, YA validado dentro de checkouts (BR-13).
	// Identidad con la que se registrará el canónico. Home = repo CANONICALIZADO (RN-IDENT-1),
	// NO el nombre del marketplace — ver C17 de este design.md.
	Identidad IdentidadArnes `json:"identidad"`

	// ── camino A ──
	OrigenLocal string `json:"origen_local,omitempty"` // <installLocation>/<ruta del source>.
	// RaizDeMarketplace marca el caso `source: "./"`: el arnés declarado ES la raíz del repo del
	// marketplace. Se materializa igual (el catálogo declara lo que declara), con Aviso visible.
	RaizDeMarketplace bool `json:"raiz_de_marketplace,omitempty"`

	// ── camino B ──
	URL string `json:"url,omitempty"` // https://host/owner/repo.git ya armada.
	// SHAEsperado es el pin del catálogo y la AUTORIDAD (§13.1 hecho 2): se hace fetch de ESTE
	// commit y se verifica que HEAD lo iguale (BR-16). "" ⇒ no hay pin: se usa Ref.
	SHAEsperado string `json:"sha_esperado,omitempty"`
	Ref         string `json:"ref,omitempty"`  // pista/fallback, NUNCA la autoridad.
	Subruta     string `json:"subruta,omitempty"` // `path` del source; "" ⇒ la raíz del repo.

	// Avisos son problemas del DATO que no impiden traer y que se muestran (BR-8): `commit`≠`sha`
	// en un source `github`, `ref` que no coincide con `sha`, `source: "./"`.
	Avisos []string `json:"avisos,omitempty"`
}

// CaminoTraer son las dos vías de AG-D17 más el rechazo explícito.
type CaminoTraer string

const (
	CaminoLocal   CaminoTraer = "local"   // copia de subcarpeta del checkout de CC. Sin red.
	CaminoExterno CaminoTraer = "externo" // fetch shallow por sha + extracción de subruta.
	CaminoNinguno CaminoTraer = ""        // no materializable: el motivo dice por qué.
)

// Errores de PLANIFICACIÓN (puros, sin I/O). El handler los mapea por errors.Is (§13.8).
var (
	// ErrTraerClaseReferencia: BR-1 — en `referencia` la acción NO EXISTE. Es el mismo
	// invariante que AccionDeSituacion ya enforça; acá se re-chequea porque un POST puede
	// llegar sin haber pasado por la UI (el botón deshabilitado no es un control de acceso).
	ErrTraerClaseReferencia = errors.New("traer: no aplica: solo arneses propios")
	// ErrTraerSourceNoMaterializable: `SourceDesconocido`, o un objeto sin `sha` NI `ref`
	// (nada que pinear), o `github` sin `sha`. El motivo lleva el crudo visible (E-90).
	ErrTraerSourceNoMaterializable = errors.New("traer: el catálogo declara un source que no sé materializar")
	// ErrTraerDestinoEscapa: BR-13 — el destino calculado no cae dentro de checkouts.
	ErrTraerDestinoEscapa = errors.New("traer: el destino calculado escapa de ~/.arnesia/checkouts")
)

// PlanificarTraer decide el plan SIN tocar disco (`installLocationExiste` lo averigua el
// usecase y lo pasa como hecho). Reglas, en orden:
//  1. clase != propio                            → ErrTraerClaseReferencia (BR-1)
//  2. Source.Tipo == SourceDesconocido            → ErrTraerSourceNoMaterializable (E-90)
//  3. Source.Tipo == SourceRutaRelativa ∧ installLocationExiste → CaminoLocal
//  4. Source.Tipo == SourceRutaRelativa ∧ ¬existe  → ErrTraerSourceNoMaterializable
//     («el catálogo declara una ruta relativa pero el marketplace no está clonado en disco»:
//      una ruta relativa no dice a qué repo pertenece — inventarlo sería fabricar el origen)
//  5. Source objeto sin SHA ni Ref                → ErrTraerSourceNoMaterializable
//  6. Source objeto                               → CaminoExterno
// El destino sale de RutaCanonico (§13.4) y se valida ahí mismo.
func PlanificarTraer(
	mkt MarketplaceConocido, fila EntradaCatalogo, installLocationExiste bool, raizCheckouts string,
) (PlanTraer, error)
```

**`Avisos` que `PlanificarTraer` puebla** (visibles, nunca bloqueantes):

| condición | aviso literal |
|---|---|
| source `github` con `commit` ≠ `sha` | `el catálogo declara dos hashes distintos (commit <c> y sha <s>) sin semántica documentada: se materializa el sha, que es el único campo presente en todas las formas` |
| objeto con `ref` **y** `sha` | `el ref declarado (<ref>) puede no apuntar al sha pineado (<sha>): manda el sha` |
| `source: "./"` | `el catálogo declara la raíz del marketplace como el arnés (source "./"): el canónico es una copia de ese árbol sin su .git` |
| sin `path` en un objeto | `el catálogo no declara subruta: el canónico es la raíz del repo` |

### 13.3 · Contratos Go — puerto y caso de uso

```go
// internal/ports/traer.go (nuevo)

// Materializador materializa un PlanTraer en un directorio de STAGING que el usecase le da.
// NO mueve al destino, NO registra nada, NO limpia: el usecase es dueño del ciclo de vida del
// temporal (BR-15) — así la atomicidad se testea en UN lugar y los dos adapters quedan tontos.
// Devuelve el sha efectivo que materializó ("" en camino local: no hay commit que reportar).
type Materializador interface {
	// Materializar deja el CONTENIDO FINAL del canónico en `staging` (ya con la subruta
	// extraída y con el set de exclusión aplicado). ctx cancelable: un fetch en vuelo se corta.
	Materializar(ctx context.Context, plan domain.PlanTraer, staging string) (shaEfectivo string, err error)
	Camino() domain.CaminoTraer
}
```

> **Por qué el staging lo dicta el usecase y no el adapter:** BR-15 (atomicidad + cero residuo) es
> una regla de negocio, no un detalle de transporte. Si cada adapter creara su propio temporal
> habría dos implementaciones de la limpieza y dos formas de dejar basura. Con esta firma, el
> `defer os.RemoveAll(tmp)` vive en **una** función y los dos caminos lo heredan.

```go
// internal/usecase/traer.go (nuevo) — métodos de MarketplaceService (mismo servicio: la acción
// nace de una fila del catálogo). Campos nuevos del struct:
//   local, externo ports.Materializador   // externo puede ser nil-safe: degrada a 503
//   pfStore        ports.PortafolioStore  // ya lo tiene, para registrar el canónico
//   deriva         ports.DerivaEvaluator  // NUEVO: firma FIJA, no se toca (BR-17)
//   raizArnesia    string                 // "" ⇒ ~/.arnesia — inyectable para tests

// Errores de EJECUCIÓN.
var (
	// ErrTraerDestinoPoblado: BR-14 — el destino ya tiene contenido. Se aborta SIN TOCAR NADA.
	ErrTraerDestinoPoblado = errors.New("traer: el destino ya tiene contenido: no se pisa el canónico")
	// ErrTraerSHANoCoincide: BR-16 — lo que vino no es lo que el catálogo declara.
	ErrTraerSHANoCoincide = errors.New("traer: el contenido traído no coincide con el sha declarado")
	// ErrTraerRemotoNoTiene: el repo/sha/ref que el catálogo declara no existe en el remoto (404).
	ErrTraerRemotoNoTiene = errors.New("traer: el remoto no tiene lo que el catálogo declara")
	// ErrTraerSinAuth: BR-18 — sin gh autenticado ni PAT para un repo que los exige (401/403).
	ErrTraerSinAuth = errors.New("traer: sin credencial para acceder al repo (gh no autenticado y sin PAT)")
	// ErrTraerLocal: fallo local (disco, permisos, rename) — 500 con el motivo real.
	ErrTraerLocal = errors.New("traer: fallo local al materializar")
)

// Traer materializa la fila `entrada` del catálogo de `nombre` como canónico editable.
// Secuencia EXACTA (spec §4.7 + BR-13..BR-18) en §13.5. Devuelve la entrada del Portafolio ya
// registrada + el veredicto de deriva recién evaluado (BR-17: se muestra tal cual salga).
func (s *MarketplaceService) Traer(ctx context.Context, nombre, entrada string) (ResultadoTraer, error)

// ResultadoTraer es lo que la UI pinta tras traer. `Deriva` viaja SIEMPRE, incluso
// `deriva-no-evaluable` con motivo: BR-17 dice que se muestra lo que salga, no lo que
// conviene.
type ResultadoTraer struct {
	Entrada       domain.EntradaPortafolio `json:"entrada"`
	Destino       string                   `json:"destino"`
	Camino        domain.CaminoTraer       `json:"camino"`
	SHAEfectivo   string                   `json:"sha_efectivo,omitempty"`
	Deriva        domain.EstadoDeriva      `json:"deriva"`
	DerivaDetalle string                   `json:"deriva_detalle,omitempty"`
	Avisos        []string                 `json:"avisos,omitempty"`
}
```

### 13.4 · BR-13 — el destino es una función pura y testeable

```go
// internal/domain/traer.go

// RaizCheckouts es el único territorio donde Traer escribe: `<raizArnesia>/checkouts`.
func RaizCheckouts(raizArnesia string) string { /* filepath.Join(raizArnesia, "checkouts") */ }

// RutaCanonico calcula el destino de un Traer y GARANTIZA que cae dentro de RaizCheckouts
// (BR-13). Es la función que E-87/E-88 atacan. Reglas:
//  1. Los dos segmentos pasan por Slug() — la MISMA regla del Portafolio: minúsculas, runs de
//     [^a-z0-9-_] colapsados a un '-', trim de '-'. Un `..` colapsa a "" y un "../../etc" a
//     "etc": estructuralmente NO puede haber un segmento de escape.
//  2. Un segmento que queda vacío tras Slug (nombre íntegramente no-ASCII, p.ej. "…") NO se
//     acepta como "": se reemplaza por HuellaPath(<crudo>) — mismo desempate que el caché de
//     catálogo, así dos nombres distintos nunca colapsan al mismo destino.
//  3. Se re-verifica el resultado con dentroDeCheckouts(): defensa en profundidad, porque la
//     garantía de (1) es un argumento sobre Slug y este chequeo es una aserción sobre el path.
// slugHome sale del NOMBRE del marketplace (legible, y es la clave única del registro);
// `id` es EntradaCatalogo.Nombre. ok=false ⇒ ErrTraerDestinoEscapa.
func RutaCanonico(raizArnesia, nombreMarketplace, id string) (destino string, ok bool)

// DentroDeCheckouts reporta si `p` está contenido en RaizCheckouts(raizArnesia) — comparación
// sobre paths Clean-eados y con el símbolo `..` ya resuelto. NO llama a EvalSymlinks: el caller
// decide si el path ya existe (destino) o todavía no (a crear). La variante que sí resuelve
// symlinks es DentroDeCheckoutsResuelto, para el chequeo POST-creación (E-88 §13.9).
func DentroDeCheckouts(raizArnesia, p string) bool
```

**El destino de E-76 sale exacto:** `~/.arnesia/checkouts/prenter-marketplace/harness/`.

> ⚠ **La trampa que hay que no caer** (C17): el **path** usa `Slug(nombreMarketplace)` (legible,
> único porque el nombre es la clave del registro) pero la **identidad** usa
> `Home = CanonicalizarRepo(mkt.Repo)` = `github.com/alpacapurpura/prenter-marketplace`. Son dos
> cosas distintas y colapsarlas rompe el cruce de §6.2: un escaneo de proyecto produce
> `Registries: ["github.com/alpacapurpura/prenter-marketplace"]`, así que si el canónico se
> registrara con `Home = "prenter-marketplace"` **no cruzaría con la fila del catálogo de la que
> vino**. El path es almacenamiento; la identidad es el contrato (RN-IDENT-1).

### 13.5 · Atomicidad (BR-15) — dónde vive el temporal y cómo se limpia

```
Traer(ctx, nombre, entrada):
  1. mkt ← fila mergeada de `nombre`                        → 404 si no está
     fila ← entrada del catálogo (del caché o una lectura)  → 400 si no está en el catálogo
  2. plan, err ← domain.PlanificarTraer(mkt, fila, installLocationExiste, raizArnesia)
     err ⇒ 400 (clase referencia · source no materializable · destino que escapa)   ← CERO I/O hasta acá
  3. BR-14 · destino: si existe Y tiene ≥1 entrada de dir  ⇒ ErrTraerDestinoPoblado (409)
     — se chequea con os.ReadDir, NO con os.Stat: un dir VACÍO no es "poblado" y se puede usar.
     Un archivo (no dir) en el destino también es "poblado". NADA se tocó todavía.
  4. tmp ← os.MkdirTemp(<raizArnesia>/tmp, "traer-*")        ← ⚠ BAJO ~/.arnesia, ver nota
     defer os.RemoveAll(tmp)                                 ← la ÚNICA limpieza, cubre las 6 ramas
     staging ← filepath.Join(tmp, "staging")   (MkdirAll 0o750)
  5. shaEfectivo, err ← materializador.Materializar(ctx, plan, staging)
     err ⇒ se mapea (502/503/500) y el defer limpia. NADA registrado, NADA en el destino.
  6. BR-16 · si plan.SHAEsperado != "" ∧ shaEfectivo != plan.SHAEsperado
        ⇒ ErrTraerSHANoCoincide (502). El defer limpia. NADA registrado.
  7. MkdirAll(filepath.Dir(destino), 0o750)
     os.Rename(staging, destino)                             ← el ÚNICO efecto observable
     err ⇒ ErrTraerLocal (500). El defer limpia. NADA registrado.
  8. re-verificación BR-13/BR-88: DentroDeCheckoutsResuelto(raizArnesia, destino)
     ¬ok ⇒ os.RemoveAll(destino) + ErrTraerDestinoEscapa (400). Ver §13.9.
  9. REGISTRO: pfStore.Upsert(EntradaPortafolio{
        Identidad: plan.Identidad,                    // Home = repo canonicalizado (C17)
        Nombre/Descripcion: de la fila del catálogo,
        Registries: [Home],                           // faceta N:M, misma regla que entradaDeCandidato
        Canonico: &Canonico{Path: destino, Version: fila.Version},
        Agregado: ahora(),
     })
     err ⇒ 500. ⚠ El dir YA existe: se devuelve el error con el destino en el motivo, sin
     borrarlo — borrar contenido ya materializado por un fallo de escritura del registro sería
     destruir lo único que costó red. Es el ÚNICO estado parcial posible y es VISIBLE
     (el próximo Traer da 409 destino-poblado, no un silencio).
 10. BR-17 · deriva ← s.deriva.Evaluar(destino, Home, id, fila.Version)  ← firma FIJA, no se toca
     Se persiste en la entrada y se devuelve TAL CUAL salga (incluso `no-evaluable` con motivo).
```

**El temporal vive en `<raizArnesia>/tmp/traer-*`, NO en `/tmp`.** Razón dura: `os.Rename` entre
filesystems distintos falla con `EXDEV`, y `/tmp` es `tmpfs` en muchas instalaciones. Poner el
staging bajo `~/.arnesia/` garantiza **mismo filesystem que el destino** ⇒ el `Rename` es atómico
de verdad. Es el error más fácil de cometer en toda esta sección.

**Las 6 ramas de falla y qué queda en disco:**

| falla en | destino | temporal | registro | código |
|---|---|---|---|---|
| 2 · planificación | intacto | no se creó | nada | 400 |
| 3 · destino poblado | **intacto** (no se abre, no se lee su contenido) | no se creó | nada | 409 |
| 5 · materialización (red/auth/404/disco) | intacto | **borrado por el `defer`** | nada | 502/503/500 |
| 6 · `sha` no coincide | intacto | **borrado por el `defer`** | nada | 502 |
| 7 · `Rename` | intacto | **borrado por el `defer`** | nada | 500 |
| 9 · `Upsert` | **existe** (único parcial posible, VISIBLE) | borrado | nada | 500 con el destino en el motivo |

**Huérfanos de una interrupción dura** (SIGKILL del daemon entre 4 y 7, E-86 sin proceso vivo que
corra el `defer`): quedan como `<raizArnesia>/tmp/traer-*`. **No se barren en caliente ni con un
GC** — se barren en `NewMarketplaceService`, al arrancar: `os.RemoveAll` de cada `traer-*` de más de
1 hora. Barrer al boot es seguro (no puede haber un Traer en vuelo de un proceso anterior que ya
murió); barrer en caliente correría el riesgo de matar un temporal de otra instancia del daemon.

### 13.6 · Camino A · `internal/adapters/traer/local.go`

`type CopiadorLocal struct{}` — satisface `ports.Materializador`, `Camino() == CaminoLocal`.

1. `origen ← EvalSymlinks(plan.OrigenLocal)`; debe existir y ser dir, y **debe seguir estando
   dentro de `installLocation`** (si el `source` relativo escapara con `..`, se rechaza: el
   catálogo ajeno no dicta dónde leemos).
2. Copia recursiva `origen → staging` con **el set de exclusión de `HashFormaPlugin`**:
   `.git/` (skip dir), `.in_use`, `.orphaned_at`.
   **Por qué exactamente ese set y no otro:** `portafolio.HashFormaPlugin` —el evaluador de deriva
   ya construido y firmado— excluye esos tres. Copiar excluyendo **el mismo** set garantiza que el
   hash del canónico recién traído iguale el de la referencia ⇒ BR-17 da `al-hilo`, que es lo que
   el spec espera. Excluir menos (llevarse `.git`) haría que el canónico heredara el remote del
   **marketplace** —peligroso: un futuro `Publicar` podría empujar al lugar equivocado—; excluir más
   alteraría contenido en silencio.
3. **Symlinks dentro del árbol copiado:** un symlink cuyo `EvalSymlinks` cae **dentro** de `origen`
   se copia **como symlink relativo** (se preserva la estructura). Uno que **escapa** de `origen`
   **no se sigue y no se copia**: se anota `Aviso` con la ruta y el destino que apuntaba. Copiar el
   contenido de un symlink que escapa metería en el canónico un archivo de otro árbol —
   exactamente lo que `dentroDe` ya protege en el walker del Portafolio (C-P-12), mismo criterio.
   Un symlink **roto** tampoco se copia; se anota.
4. **Permisos:** dirs `0o750`, archivos `0o640` **más el bit de ejecución del origen** si lo tenía
   (`perm & 0o111`) — un hook o un script del arnés tiene que seguir siendo ejecutable. No se
   copian setuid/setgid/sticky (nunca los necesita un árbol de arnés y son superficie de riesgo).
5. **`source: "./"`** (`plan.RaizDeMarketplace`): se copia igual, con el mismo set de exclusión —
   así el `.git` del marketplace **no** entra— y con el `Aviso` que ya puso `PlanificarTraer`.
   Verificado que el caso existe (caveman, ponytail) aunque los dos sean `referencia` y por BR-1
   nunca lleguen acá; un marketplace **propio** de un solo plugin sí podría usarlo.
6. **Cero red, cero auth, cero `git`.** `Camino() == CaminoLocal` y el adapter no importa nada de
   red: es verificable por inspección de imports.

### 13.7 · Camino B · `internal/adapters/traer/externo.go`

`type ClonadorExterno struct { GitBin, GHBin, Token string; Timeout time.Duration; MaxBytes int64 }`
— satisface `ports.Materializador`, `Camino() == CaminoExterno`.

**`git`, no `gh`, para traer — y `gh` como credential helper.** Los dos son shell-out a un CLI
del operador (criterio ya fijado en §5.3, precedentes `selfupdate/updater.go` y
`publish/publisher.go`), pero para **esta** operación `git` gana por tres razones medidas:

| criterio | `git` sparse+fetch-por-sha | `gh api …/tarball/{ref}` |
|---|---|---|
| verificación de `sha` (BR-16) | `git rev-parse HEAD` = **40 hex exactos** | el tar trae el sha en el nombre del dir raíz, **7 chars** → o una 2ª llamada a la API |
| datos bajados | `--filter=blob:none` + cone: **solo los blobs de la subruta** (medido: 664 KB en un repo con 78 plugins) | el tarball **completo** del repo |
| pin por `sha` sin `ref` | soportado (**143/220 entradas lo necesitan**, §13.1) | el endpoint quiere un `ref`; con un sha funciona pero sin verificación fuerte |
| credencial | `-c credential.helper='!gh auth git-credential'` ⇒ **ArnesIA nunca ve el token** (BR-18 literal) | idem vía `gh` |

**Secuencia exacta** (probada en vivo, §13.1 hecho 6):

```
git -C <staging-padre> init -q <work>
git -C <work> remote add origin <plan.URL>
git -C <work> sparse-checkout init --cone
git -C <work> sparse-checkout set <plan.Subruta>          # solo si Subruta != ""
git -C <work> -c credential.helper='!<gh> auth git-credential' \
    fetch --depth 1 --filter=blob:none origin <pin>
git -C <work> checkout -q FETCH_HEAD
sha ← git -C <work> rev-parse HEAD
```

- **`pin` = `plan.SHAEsperado`** cuando existe (220/220 hoy). Sin `sha`, `pin = plan.Ref`; sin
  ninguno de los dos, `PlanificarTraer` ya rechazó (regla 5).
- **`ref` NUNCA es el pin cuando hay `sha`** — §13.1 hecho 2: en una entrada sana apuntan a
  commits distintos, y el spec, tomado literal, abortaría siempre.
- **Si el remoto no permite fetch por sha** (`allowReachableSHA1InWant` apagado — GitHub sí lo
  permite, verificado; otro host puede no): se **reintenta una sola vez** con
  `fetch --depth 1 --branch <Ref>` **si hay `Ref`**, y en ese caso el chequeo de BR-16 sigue
  aplicando (si `HEAD != sha` ⇒ 502, honesto). Sin `Ref`, se aborta con el stderr real como motivo.
- **Extracción de la subruta:** el contenido final es `<work>/<Subruta>` (o `<work>` si no hay
  subruta). Se copia a `staging` con **el mismo set de exclusión y las mismas reglas de symlink y
  permisos del camino A** (§13.6 pasos 2-4) — eso descarta `.git`, y también los archivos de raíz
  que el cone mode arrastra (§13.1 hecho 7). Una sola función de copia compartida por los dos
  caminos, en `internal/adapters/traer/copia.go`.
- **Auth (BR-18):** `credential.helper='!gh auth git-credential'` cuando `gh` está autenticado.
  Fallback PAT: `GIT_ASKPASS` a un script de 2 líneas escrito **dentro del temporal** en modo
  `0o700` que hace `echo "$ARNESIA_GH_TOKEN"`, más `GIT_TERMINAL_PROMPT=0`. El token viaja por
  **env del hijo**, jamás en `argv` (visible en `ps`) ni en la URL (queda en el reflog). Reusa
  `elegirCredencial` (§5.3), que ya es puro y testeable sin login.
- **Guardas:** `context` con `Timeout` (default 60 s, más generoso que la lectura de catálogo
  porque acá se baja un árbol) · `MaxBytes` (default 64 MiB) chequeado **después** del fetch
  midiendo `<work>` con un `WalkDir` acumulador → excedido ⇒ abortar con motivo, el `defer` limpia
  · `GIT_TERMINAL_PROMPT=0` y `GIT_CONFIG_NOSYSTEM=1` (que la config global del operador no cambie
  el comportamiento) · stderr capturado y **usado como motivo**.
- **Clasificación del error por stderr** (E-84/E-85 exigen motivos distintos, ninguno genérico):

  | señal en stderr / exit | error centinela | HTTP |
  |---|---|---|
  | `could not read Username`, `Authentication failed`, `403` | `ErrTraerSinAuth` | 503 |
  | `Repository not found`, `404`, `couldn't find remote ref`, `not our ref` | `ErrTraerRemotoNoTiene` | 502 |
  | `context deadline exceeded` / kill por timeout | `ErrTraerRemotoNoTiene` con motivo de timeout | 502 |
  | `No space left`, `Permission denied` en el staging | `ErrTraerLocal` | 500 |

### 13.8 · Endpoint

```go
mux.HandleFunc("POST /api/marketplaces/{nombre}/traidos", postTraerCanonico(marketplaces))
```

Nombre en plural-acción-como-recurso, igual que `/escaneos`, `/validaciones`, `/lecturas`.

Request: `{"entrada":"harness"}` — la fila del catálogo, no el `source` (el cliente **no** elige el
mecanismo: el dominio lo decide con `PlanificarTraer`).

200:

```json
{
  "entrada": { "clave": "github-com-alpacapurpura-prenter-marketplace~harness~", "identidad": { "home": "github.com/alpacapurpura/prenter-marketplace", "id": "harness" },
               "canonico": { "path": "/home/chalreme/.arnesia/checkouts/prenter-marketplace/harness", "version": "0.5.3" },
               "registries": ["github.com/alpacapurpura/prenter-marketplace"], "agregado": "2026-07-25T15:02:44Z" },
  "destino": "/home/chalreme/.arnesia/checkouts/prenter-marketplace/harness",
  "camino": "local",
  "deriva": "al-hilo",
  "avisos": []
}
```

| status | cuándo | por qué ese y no otro |
|---|---|---|
| **200** | traído y registrado — **incluso si `deriva` salió `en-deriva` o `no-evaluable`** | BR-17: la deriva se muestra tal cual salga; un veredicto incómodo no es un fallo de la operación |
| **400** | clase `referencia` (BR-1) · `entrada` que no está en el catálogo · `source` no materializable (E-90) · destino que escapa (BR-13, E-87) | son precondiciones del pedido: nada se intentó |
| **404** | marketplace desconocido | coherente con §7.5 |
| **409** | **destino ya poblado** (BR-14, E-79) | conflicto con el estado existente, y el cliente **puede** resolverlo (mirar/mover su canónico). El 409 lleva `{"error":…,"destino":…}` para que la UI pueda ofrecer «abrir el canónico que ya tenés» |
| **502** | el remoto no tiene lo que el catálogo declara (`sha`/`ref`/repo inexistente, E-83/E-85-404) · **`sha` que no coincide** (BR-16, E-82) | «miré, y lo que vino no es lo declarado» — es un tercer estado, distinto de «tu pedido está mal» (400) y de «no puedo mirar» (503). El estante mintió, no el cliente |
| **503** | sin `gh` ni PAT (BR-18, E-84) · 401/403 del remoto (E-85-403) | mismo criterio que §7.2: **«no puedo mirar» nunca se pinta como «tu pedido está mal»** |
| **500** | fallo local: disco lleno, permisos, `Rename`, `Upsert` (E-89) | es nuestro, no del cliente ni del remoto |

**E-85 queda con dos códigos distintos** (`403 → 503`, `404 → 502`) **más dos motivos distintos** —
la exigencia de «motivos distintos, ninguno genérico» se cumple en el status *y* en el texto.

`docs/architecture/contracts/api/openapi.yaml`: el bump ya previsto a `0.7.0-marketplaces` incluye
este path; **documentar que un 200 puede traer `deriva: "en-deriva"`** — si no, el próximo lector
asume que 200 ⟹ al-hilo y agrega una validación que rompe BR-17.

### 13.9 · BR-13 / E-88 — «jamás escribe en `~/.claude`», por aserción de ruta

Tres capas, ninguna de confianza:

1. **Estructural (dominio, puro):** `RutaCanonico` construye el destino **solo** por
   `filepath.Join(RaizCheckouts(raiz), Slug(a), Slug(b))`. `Slug` colapsa todo lo que no es
   `[a-z0-9-_]`, así que **ningún input puede producir un separador ni un `..`**: `"../../.claude"`
   → `"-claude"`. No hay concatenación de input crudo en ninguna parte.
2. **Aserción de path (dominio, puro):** `DentroDeCheckouts` sobre el destino Clean-eado, **antes**
   de crear nada, y `DentroDeCheckoutsResuelto` (con `EvalSymlinks`) **después** del `Rename` —
   porque un atacante con acceso al FS podría haber puesto un symlink en
   `~/.arnesia/checkouts/<algo>` apuntando afuera. Si la re-verificación falla, se borra el destino
   y se devuelve 400. Es defensa en profundidad real, no ceremonia.
3. **Aserción de test (E-88), no de confianza:** el test corre el Traer con `raizArnesia` = un
   `t.TempDir()` y un `HOME` falso que contiene un `.claude/` **poblado con archivos y su mtime
   registrado**; tras el Traer (y tras cada rama de falla) asserta que **el árbol `.claude/`
   completo es byte-por-byte y mtime-por-mtime idéntico**, y que **todo path escrito** cae bajo la
   raíz inyectada. Además un `Materializador` fake que **intenta** escribir en `<HOME>/.claude` y
   verifica que la operación falle sin dejar rastro. Se afirma sobre rutas y bytes, no sobre buena
   voluntad.

### 13.10 · Máquina de estados en el FE

La acción vive en **dos puertas, un acto** (AG-D8 decisión 5): la fila del catálogo (S3) y el botón
del drawer (`portafolio-drawer.tsx`, hoy `disabled` con `TOOLTIP_S2`). **Un solo estado en la
página**, compartido por las dos puertas:

```
        ┌────────┐  click  ┌───────────┐  200  ┌────────┐
  ──▶   │ ofrece │────────▶│ trayendo  │──────▶│ traído │──▶ (refetch portafolio + catálogo)
        └────────┘         └───────────┘       └────────┘
             ▲                   │ 4xx/5xx          │ «ver ficha» / «abrir en Mapa»
             │                   ▼                  ▼
             │             ┌───────────┐        el drawer re-apunta a la clave nueva
             └─────────────│  falló    │
               «reintentar»└───────────┘
                                 │ 409
                                 ▼  «abrir el canónico que ya tenés»
```

| estado | qué se ve | qué NO se puede hacer |
|---|---|---|
| `ofrece` | `↧ Traer canónico` habilitado **solo** si `accion.verbo == "traer-canonico" && accion.habilitada`; si no, `disabled` con `accion.motivo` (literal del dominio) | — |
| `trayendo` | spinner + `Trayendo…`, botón `disabled`; **la fila entera queda no-interactiva** (no se puede disparar dos veces la misma) | cerrar el catálogo NO aborta el POST (a diferencia de la lectura de catálogo): abortar a mitad de una materialización es peor que esperar — mismo criterio que S1-D19 |
| `traído` | el veredicto de deriva **tal cual salga** (`al-hilo` verde · `en-deriva` atención · `no-evaluable` `sin-senal` + motivo) + los `avisos` + link a la ficha | — |
| `falló` | el motivo textual del backend + `Reintentar`; en 409, además `abrir el canónico que ya tenés` con el `destino` que vino en el body | — |

Reglas de UI:

- **La habilitación NO se decide en el FE.** `accion.habilitada` viene del dominio
  (`AccionDeSituacion`) — y como AG-D17 la habilita para `propio`/`no-lo-tengo`, es **la única
  celda de la tabla de §6.3 que cambia** (`{traer-canonico, true, ""}`); las otras 7 siguen en
  `false`. Ver §13.11.
- **Un Traer en vuelo por identidad**, no global: se puede traer `harness` y `harness-beta` a la
  vez (destinos distintos, E-78). La clave del estado es `nombre@entrada`.
- **Tras el 200 se refetchean las dos superficies**: el Portafolio (aparece el canónico) y el
  catálogo (la fila pasa de `no-lo-tengo` a `al-hilo`/lo que la deriva diga). Sin refetch, la
  columna de situación quedaría mintiendo con el estado viejo.
- **El botón del drawer** (`portafolio-drawer.tsx:287`) deja de ser `disabled`+`TOOLTIP_S2` y pasa
  a llamar el mismo callback. Prop nueva del contrato: `onTraerCanonico?: (() => void) | undefined`
  + `trayendo?: boolean` + `traerError?: string | undefined` — opcionales, así las stories firmadas
  del Slice 1 siguen pasando (BR-12, superset estricto).

### 13.11 · Lo que cambia en lo ya diseñado

| § | cambio |
|---|---|
| §6.3 (`AccionDeSituacion`) | **una sola celda**: `propio` × `no-lo-tengo` pasa de `{traer-canonico, false, "…ítem 2…"}` a **`{traer-canonico, true, ""}`**. Las otras 7 filas quedan idénticas. La frase «`Habilitada` es `false` en las 8 filas» de §6.3 pasa a «en 7 de 8» |
| §2 C2 | **RESUELTA** por AG-D17: se construye completo |
| §2 C12 | **RETIRADA**: era lectura stale — AG-D11 está FIRMADA |
| §8.7 | **CERRADO**: no se deduplica por `source`; la fila es el canal |
| §3.5 | `MarketplaceService` gana 3 campos (`local`, `externo`, `deriva`) y 1 método (`Traer`) |
| §3.9 | `cmd`: `newMarketplaceService` recibe también el `ports.DerivaEvaluator` que `newPortafolioService` ya construye (`&portafolio.Referencias{}`) — **se reusa la misma instancia**, no se crea una segunda |
| §11.4 | `.go-arch-lint.yml`: componente `traer` (`internal/adapters/traer/**`) → `mayDependOn: [domain, ports]`, y `cmd` lo suma. Mismo diff diferido que el componente `marketplace` (§3.9): no se aplica hasta que el paquete exista |
