// Domain types del Marketplace (paquete 2026-07-23-portafolio-agregar-marketplace): espejo
// EXACTO del wire Go — `internal/domain/marketplace.go` + `marketplace_situacion.go` +
// `internal/usecase/marketplace.go` (design.md §3.1/§3.2/§3.5) y los endpoints de §7. Los
// nombres de campo son los tags `json:` del Go, tal cual. Hand-authored, mismo patrón que
// `entities/portafolio/model/types.ts` (hasta que el codegen lo reemplace) — ver
// docs/architecture/boundaries/fe-transporte-independiente.md.
//
// Entidad PROPIA y no parte de `entities/portafolio` (spec §2.2): un marketplace no es un
// arnés — tiene identidad (`nombre`), ciclo de lectura y clase propios. Los dos se cruzan solo
// en la página. Esta entidad **NO importa `entities/portafolio`** (`steiger
// fsd/no-cross-imports`, dentro de `pnpm run verify`).

// ClaseMarketplace parte el universo en dos, y NO son simétricas (AG-D8 decisión 1):
// `propio` = publicamos ahí (participa de traer/publicar/actualizar/reparar);
// `referencia` = solo resuelve la procedencia de un arnés ajeno — catálogo read-only, jamás
// operable (boundary marketplace-referencia-es-solo-procedencia).
export type ClaseMarketplace = "propio" | "referencia"

// EslabonMarketplace dice de dónde salió el CONOCIMIENTO de que este marketplace existe
// (AG-D9, collect-all: no exclusivo, ≥1, los dos = detectado + confirmado). Eje DISTINTO de
// `EslabonOrigen.fuente` de entities/portafolio, que dice de dónde salió un dato de UNA COPIA
// instalada; no unificar (design.md §2 C9): son dos preguntas.
export type EslabonMarketplace = "cc-known-marketplaces" | "declarado-por-operador"

// TipoLectura son los 4 estados explícitos de AG-D8 decisión 4. `no-leido` es el CERO del tipo
// a propósito: un `EstadoLectura` sin poblar dice «no leído aún», nunca «leído con 0 entradas».
export type TipoLectura = "no-leido" | "leido" | "sin-acceso" | "url-no-resuelve"

// EstadoLectura es el veredicto de la última lectura de catálogo de un marketplace.
// Invariante (BR-4): `tipo === "leido"` ⟺ `entradas` es una cuenta REAL y `cuando` ≠ "".
// Cualquier otro `tipo` EXIGE `motivo` ≠ "" — un degradado sin motivo es un degradado mudo.
export interface EstadoLectura {
  tipo: TipoLectura
  /** RFC3339 UTC de la última lectura EXITOSA (aunque `tipo` ya no sea `leido`: así la fila
   *  puede decir «leído hace 2 días · ahora sin acceso» — BR-3 + BR-4 juntas). */
  cuando?: string
  /** cuenta de filas de esa última lectura exitosa; solo significativa con `cuando` ≠ "". */
  entradas?: number
  /** el texto que la UI muestra TAL CUAL (gh no autenticado, 404, JSON inválido…). */
  motivo?: string
  /** "local" (checkout de CC) | "remoto" (gh api) | ausente (nunca leído). */
  fuente?: string
}

// TipoSource son las formas reales que `plugins[].source` toma en los marketplaces de esta
// máquina (AG-D13, verificado sobre 273+2+1+1+1 entradas). El adapter traduce; el FE solo ve
// esto ya normalizado.
export type TipoSource = "ruta-relativa" | "git-subdir" | "url" | "github" | "desconocido"

// SourceCatalogo es `plugins[].source` normalizado. `crudo` es la representación textual
// estable con la que se agrupan canales (AG-D11) — jamás se muestra como si fuera una ruta
// local cuando no lo es.
export interface SourceCatalogo {
  tipo: TipoSource
  crudo: string
  ruta?: string
  url?: string
  repo?: string
  ref?: string
  sha?: string
}

// MarketplaceConocido es UNA fila del plano Marketplaces (S2), ya mergeada collect-all
// (AG-D9). La clave de merge es `nombre` — el mismo con el que Claude Code keyea
// `installed_plugins.json` (`"<id>@<nombre>"`), o sea el nombre INSTALABLE, no el repo.
export interface MarketplaceConocido {
  nombre: string
  /** repo canonicalizado `host/owner/repo` (RN-IDENT-1); ausente si ninguna fuente lo declara. */
  repo?: string
  clase: ClaseMarketplace
  /** ordenados y dedupeados (detectado antes que declarado), ≥1 siempre. */
  eslabones: EslabonMarketplace[]
  /** el checkout local que Claude Code ya mantiene — la ruta contra la que `deriva` compara. */
  install_location?: string
  /** el `lastUpdated` que CC declara (RFC3339). Dato de CC, NO nuestra lectura: nunca se
   *  muestra como «leído hace…» (eso es `lectura.cuando`). */
  cc_actualizado?: string
  lectura: EstadoLectura
  /** discrepancias entre eslabones: se MUESTRAN, jamás se eligen en silencio (BR-8). */
  discrepancias?: string[]
  /** RFC3339 de cuándo el operador lo declaró; ausente si solo lo detectó CC. */
  registrado?: string
}

// ProcedenciaVersion dice de DÓNDE salió `version` — BR-2 con AG-D14: nunca se inventa, y
// siempre se puede auditar cuál de las dos fuentes ganó.
export type ProcedenciaVersion = "campo-version" | "derivada-de-source"

// TipoSituacion son las 6 ramas de spec §4.3. La 6ta (`no-comparable`) es de PRIMERA CLASE:
// sin ella el sistema tendría que elegir entre mentir (`al-hilo` por defecto) o callarse (BR-9).
export type TipoSituacion =
  | "no-lo-tengo"
  | "al-hilo"
  | "mi-copia-adelantada"
  | "estante-adelantado"
  | "instalaciones-en-deriva"
  | "no-comparable"

// SituacionCatalogo es el veredicto del cruce catálogo × Portafolio para UNA fila.
// **La calcula el dominio Go y viaja resuelta en el wire** (design.md §2 C1): el FE solo la
// PRESENTA. Invariante: `tipo === "no-comparable"` ⟹ `motivo` ≠ "".
export interface SituacionCatalogo {
  tipo: TipoSituacion
  /** versión del canónico propio. */
  mia?: string
  /** versión de la fila del catálogo. */
  estante?: string
  /** instalaciones en-deriva. */
  cuantas?: number
  motivo?: string
  /** `Identidad.Clave()` de la entrada cruzada — el link «ver ficha» sin re-derivar el slug. */
  clave_portafolio?: string
  /** CÓMO se cruzó (BR-3): `home-declarado` | `faceta-registry` | `rename`. La UI muestra el
   *  cruce débil COMO TAL — no lo presenta como identidad resuelta. */
  via?: string
}

// Accion es el verbo que la fila del catálogo ofrece (AG-D8 decisión 5): `traer-canonico` es el
// `↧ Traer canónico` que YA existe en el drawer del arnés — dos puertas, un acto, cero
// vocabulario nuevo (mockups/INDEX.md regla dura 4).
export type Accion = "" | "traer-canonico" | "publicar" | "actualizar-mi-copia" | "reparar"

// AccionCatalogo es el verbo + si está habilitado + POR QUÉ no. `habilitada` y `motivo` vienen
// del DOMINIO (`AccionDeSituacion`, design.md §6.3): el widget renderiza `disabled` +
// `title={motivo}` con el literal EXACTO que vino del wire. **El FE nunca decide si un botón se
// habilita** — ni puede habilitar uno que el dominio dejó apagado.
export interface AccionCatalogo {
  verbo: Accion
  habilitada: boolean
  motivo?: string
}

// EntradaCatalogo es UNA fila del catálogo (AG-D11 FIRMADA: la ENTRADA DE ÍNDICE, o sea el
// canal — lo que el cliente instala con `/plugin install <nombre>@<marketplace>`; dos entradas
// pueden compartir `source` y NO son dos arneses).
export interface EntradaCatalogo {
  nombre: string
  source: SourceCatalogo
  version?: string
  version_de?: ProcedenciaVersion
  descripcion?: string
  /** los OTROS nombres del mismo catálogo con idéntico `source.crudo` (canales, AG-D11). */
  comparte_source_con?: string[]
  /** solo del enriquecimiento opcional `catalogo.json` (convención de prenter, NO del estándar). */
  estado_canal?: string
  /** si `renames` mapea algún nombre viejo a este (AG-D15): así un arnés keyeado con el nombre
   *  viejo sigue cruzando. */
  nombre_anterior?: string[]
  /** problemas de CALIDAD DEL DATO de esta fila (no de nuestro cálculo). Se MUESTRAN completos
   *  (BR-8) y jamás se convierten en un descarte silencioso de la fila. */
  aviso?: string[]
  situacion: SituacionCatalogo
  accion: AccionCatalogo
}

// VersionCatalogo es una fila de `catalogo.json#versiones[]` (convención de prenter).
export interface VersionCatalogo {
  version: string
  /** `habilitada` | `deprecada` — crudo, sin enum: es formato ajeno. */
  estado: string
  fuente?: string
  fecha?: string
}

// Catalogo es el catálogo leído de UN marketplace.
//
// **`entradas: null` es NORMATIVO** (design.md §2 C4, BR-4): `null` ⟺ NO hay lectura válida
// («no sé»), `[]` ⟺ el marketplace declara `plugins: []` de verdad («leí y no declara ninguno»).
// La UI JAMÁS dice «no tiene arneses» con `null`: dice el motivo. Es el mecanismo
// anti-pass-fabricado a nivel cable.
export interface Catalogo {
  marketplace: string
  clase: ClaseMarketplace
  repo?: string
  owner_nombre?: string
  owner_email?: string
  owner_url?: string
  descripcion?: string
  lectura: EstadoLectura
  entradas: EntradaCatalogo[] | null
  /** enriquecimiento opcional de `catalogo.json` (prenter). Ausente = degradado SIN ruido. */
  canales?: Record<string, string>
  versiones?: VersionCatalogo[]
  /** cuántas entradas se descartaron por el techo de seguridad. > 0 es un dato VISIBLE, jamás
   *  un recorte silencioso (E-28/E-58). */
  truncado?: number
}

// EntradaCorruptaMarketplace — wire de `corruptaWire` del listado (solo el motivo; el blob
// crudo no cruza al FE, mismo criterio que `EntradaCorrupta` del Portafolio).
export interface EntradaCorruptaMarketplace {
  motivo: string
}

// ListadoMarketplaces — `GET /api/marketplaces` (design.md §7.1). `marketplaces` NUNCA es null:
// si no hay ninguno, `[]` — «no conozco ningún marketplace» es una afirmación verdadera y
// verificable (a diferencia de un catálogo no leído).
export interface ListadoMarketplaces {
  marketplaces: MarketplaceConocido[]
  corruptas?: EntradaCorruptaMarketplace[]
  aviso_detector?: string
  sin_origen_resuelto: number
}

// Validacion es lo que S6 pinta en el ✓ (`POST /api/marketplaces/validaciones`): datos LEÍDOS
// del archivo real, nunca inferidos. Solo un 200 pinta ✓ — BR-5/G3 hecho contrato de transporte.
export interface Validacion {
  url_canonica: string
  nombre: string
  owner_nombre?: string
  owner_email?: string
  owner_url?: string
  descripcion?: string
  entradas: number
  /** `local` | `remoto`. */
  fuente: string
  ya_registrado: boolean
  clase_actual?: ClaseMarketplace
}

// CandidatoOrigen alimenta el selector de S7 (`GET …/origen/candidatos`), YA ORDENADO por el
// backend por señales BLANDAS. `senal` se ARMA en el usecase, no en el widget, para que sea
// testeable. **NADA premarcado**: el orden es sugerencia, la elección es del operador (BR-11).
export interface CandidatoOrigen {
  nombre: string
  repo?: string
  clase: ClaseMarketplace
  senal?: string
}

// CandidatosOrigen — el wire completo del endpoint. `actual` = el home ya declarado ("" si la
// identidad es provisional), para que el diálogo pueda decir «ya tiene origen».
export interface CandidatosOrigen {
  candidatos: CandidatoOrigen[]
  actual?: string
}

// CaminoTraer son las dos vías de AG-D17 (§13.2): `local` = copia de una subcarpeta del
// checkout que CC ya mantiene (sin red) · `externo` = fetch shallow por `sha` + extracción.
export type CaminoTraer = "local" | "externo"

// ResultadoTraer — 200 de `POST /api/marketplaces/{nombre}/traidos` (§13.8). `deriva` viaja TAL
// CUAL salga (BR-17): un 200 con `en-deriva` NO es un fallo de la operación, es un dato honesto.
// `entrada` es la `EntradaPortafolio` re-keyed; el FE la usa vía refetch, así que acá viaja
// como `unknown` a propósito — tipar el arnés acá exigiría importar `entities/portafolio`
// (cross-import prohibido, `steiger fsd/no-cross-imports`).
export interface ResultadoTraer {
  entrada?: unknown
  destino: string
  camino: CaminoTraer
  /** `al-hilo` | `en-deriva` | `deriva-no-evaluable`, literal del wire (BR-17). */
  deriva: string
  deriva_detalle?: string
  avisos?: string[]
}

// EstadoTraer — la máquina de estados de `↧ Traer canónico` en el FE (§13.10). NO es wire: es
// estado de vista, y vive en la ENTIDAD (no en un widget) porque lo comparten dos superficies —
// la fila del catálogo (`widgets/marketplace`) y el botón del drawer (`widgets/portafolio`) —
// que NO pueden importarse entre sí (`no-sibling-widget-imports`, severidad `error`).
//
// El estado `ofrece` es la AUSENCIA de entrada en el mapa (nada en vuelo, nada resuelto): un
// cuarto miembro «ofrece» sería un estado que hay que recordar apagar.
export type EstadoTraer =
  | { fase: "trayendo" }
  | {
      fase: "traido"
      destino: string
      camino: CaminoTraer
      /** el veredicto de deriva TAL CUAL salga (BR-17) — un `en-deriva` acá no es un fallo. */
      deriva: string
      derivaDetalle?: string
      avisos?: string[]
      clavePortafolio?: string
    }
  | {
      fase: "fallo"
      /** el motivo textual del backend, literal. */
      motivo: string
      /** solo en el 409 (BR-14): el destino que ya está poblado, para poder ofrecer abrirlo. */
      destinoExistente?: string
    }
