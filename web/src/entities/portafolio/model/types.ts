// Domain types del Portafolio (Slice 0/1, HS-22/HS-23): espejo EXACTO del wire Go —
// internal/domain/portafolio.go (IdentidadArnes/Instalacion/OrigenPortafolio/Canonico/
// EntradaPortafolio) + internal/usecase/portafolio.go (Candidato) + el wrapper HTTP
// (internal/adapters/transport/http/portafolio.go: entradaWire agrega `clave`, corruptaWire
// expone solo `motivo`). Mismo patrón que entities/arnes/model/types.ts: hand-authored hasta
// que el codegen (quicktype) reemplace esto — ver
// docs/architecture/boundaries/fe-transporte-independiente.md.
//
// El tipo se llama `OrigenCopia` (NO `Origen`) para no chocar con el `Origen` L0 de
// entities/arnes (nodo.origen: estandar/del-puesto) — mismo motivo que domain.OrigenPortafolio
// en Go (S0-D12); la UI lo rotula «origen de la copia» (S1-D14, plan §2.3). El campo JSON
// sigue siendo `origen`.

// EstadoDeriva — el veredicto de BR-4 (domain.DerivaAlHilo/EnDeriva/NoEvaluable): nunca
// semver-string ni `git status`, siempre hash de contenido contra la referencia inmutable.
export type EstadoDeriva = "al-hilo" | "en-deriva" | "deriva-no-evaluable"

// TipoInstalacion — las tres formas físicas que el walker distingue (S0-D2).
export type TipoInstalacion = "materializada" | "referenciada-cc" | "proyecto-instalado"

// EslabonOrigen — dato crudo de la cadena de resolución de origen (BR-3): trazabilidad
// honesta, jamás se pierde de dónde salió un dato.
export interface EslabonOrigen {
  fuente: string
  campo: string
  valor: string
}

// OrigenCopia — reconciliación collect-all de una instalación (spec §6, domain.OrigenPortafolio).
// registry/version vacíos son desconocido/`?` honestos, jamás inventados; discrepancias
// nunca se resuelven en silencio (C-OR-6).
export interface OrigenCopia {
  registry?: string
  version?: string
  eslabones?: EslabonOrigen[]
  discrepancias?: string[]
}

// IdentidadArnes — la clave del Portafolio (BR-1): home = marketplace autor-declarado
// canonicalizado (RN-IDENT-1); home=="" ⇒ identidad provisional (RN-IDENT-2) y scope discrimina.
export interface IdentidadArnes {
  home?: string
  id: string
  scope?: string
}

// Instalacion — materialización read-only de un arnés en un proyecto (spec §2, N por identidad).
// `tipo: ""` es wire REAL (S1-D26): un hallazgo declarado sin dir físico resoluble (lock/CC,
// C-P-14) no tiene forma que clasificar — domain.HallazgoInstalacion.Tipo viaja vacío.
export interface Instalacion {
  proyecto_path: string
  install_path: string
  tipo: TipoInstalacion | ""
  origen: OrigenCopia
  deriva: EstadoDeriva
  deriva_detalle?: string
  aviso?: string
}

// Canonico — la única copia editable de una identidad (spec §2, 0 o 1 por identidad). La ley
// anti-drift en la UI: ningún botón «Editar» sobre una instalación, nunca (INV-1/C-BK-3).
export interface Canonico {
  path: string
  version?: string
}

// EntradaPortafolio — 1 card por identidad (BR-1): empresas y registries son facetas N:M,
// nunca dueñas de la identidad. `clave` la agrega el wire HTTP (entradaWire, Identidad.Clave()
// ya resuelto) — no existe en domain.EntradaPortafolio a secas, pero SÍ en todo lo que el FE
// consume (GET /api/portafolio, POST /api/portafolio/proyectos re-fetcheado).
export interface EntradaPortafolio {
  clave: string
  identidad: IdentidadArnes
  nombre?: string
  descripcion?: string
  empresas?: string[]
  registries?: string[]
  canonico?: Canonico
  instalaciones?: Instalacion[]
  agregado?: string
  // origen_sin_resolver_desde — RFC3339 de cuándo el operador CONFIRMÓ dejar esta entrada sin
  // origen (S7, opción «ninguno — dejarlo sin origen»; BR-11/E-26 del paquete
  // 2026-07-23-portafolio-agregar-marketplace). Distinto de ausente, que significa «todavía nadie
  // lo miró»: el contador cruzado «N sin origen resuelto» cuenta las provisionales SIN este
  // campo. Confirmar «ninguno» no inventa un home — solo deja de reclamar atención.
  origen_sin_resolver_desde?: string
}

// EntradaCorrupta — wire de corruptaWire (HTTP): SOLO el motivo, el blob crudo NO viaja
// (a diferencia de domain.EntradaCorrupta, que también lleva `raw` — no cruza al FE).
export interface EntradaCorrupta {
  motivo: string
}

// PortafolioListado — GET /api/portafolio.
export interface PortafolioListado {
  entradas: EntradaPortafolio[]
  corruptas?: EntradaCorrupta[]
}

// Candidato — resultado de un escaneo (POST /api/portafolio/escaneos), NO PERSISTIDO — el
// usuario elige cuáles agregar (spec §7.1). Espejo de usecase.Candidato.
export interface Candidato {
  clave: string
  identidad: IdentidadArnes
  nombre?: string
  descripcion?: string
  empresas?: string[]
  instalacion: Instalacion
  es_canonico?: boolean
}

// SaludPortafolio / LentePortafolio — vocabulario PROPIO del Portafolio (S1-D4): NO se reusa
// el `Salud` de sesión (shared/api/types.ts: ok/warn/crit/info, otro significado — «no pisar
// vocabulario»).
export type SaludPortafolio = "ok" | "atencion" | "sin-senal"
export type LentePortafolio = "empresa" | "plano" | "proyecto" | "marketplace"
