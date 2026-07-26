import { cn } from "@/shared/lib/cn"
import { etiquetaDeSituacion, textoDeLectura, tonoDeSituacion } from "../model/selectors"
import type { ClaseMarketplace, EstadoLectura, SituacionCatalogo } from "../model/types"

// Chips de dominio del Marketplace (design.md §8.4): presentacionales, CERO transporte. Cada uno
// pinta SOLO lo que el wire trae — sin dato, no renderiza (mismo criterio que `AvisoChip` del
// Portafolio: el caller no necesita un `if` propio).
//
// Reglas de token que esta capa NO puede violar (spec §1):
//   · `no-leido` usa la clase `sin-senal` (transparente + dashed sobre `--muted-foreground`),
//     NUNCA el color de `ok` — «no sé» ≠ «sano».
//   · `sin-acceso`/`url-no-resuelve` son `--warn` (recuperable: autenticar), JAMÁS `--crit`
//     (reservado a fallo).
//   · todo texto sobre `--warn-soft` va en `--foreground` (no se agrava la deuda a11y de
//     `.text-warn`, contraste 4.5:1 abierto en BACKLOG.md).
//
// El dot de salud (`DotSaludPortafolio`) NO se importa acá: es un átomo de
// `entities/portafolio` y un cross-import entity↔entity rompe `steiger fsd/no-cross-imports`.
// Lo compone el WIDGET, que sí puede importar las dos entidades.

// ── ClaseChip — la partición de AG-D8 decisión 1, visible ANTES de entrar al catálogo.
// Descriptivo (sin tono de salud): la clase no es un estado sano/enfermo, es qué se puede hacer.
interface ClaseChipProps {
  clase: ClaseMarketplace
}

const CLASE_LABEL: Record<ClaseMarketplace, string> = {
  propio: "propio",
  referencia: "de referencia",
}

export function ClaseChip({ clase }: ClaseChipProps) {
  return <span className={cn("pf-chip", "pf-chip-clase", clase)}>{CLASE_LABEL[clase]}</span>
}

// ── EstadoLecturaChip — los 4 estados de AG-D8 decisión 4 con su motivo COMPLETO. `ahora` se
// inyecta (no `new Date()` adentro): el texto tiene que ser determinista para story=test.
interface EstadoLecturaChipProps {
  lectura: EstadoLectura
  ahora: Date
}

export function EstadoLecturaChip({ lectura, ahora }: EstadoLecturaChipProps) {
  const texto = textoDeLectura(lectura, ahora)
  return (
    <span className={cn("pf-chip", "pf-chip-lectura", lectura.tipo)} title={texto}>
      {texto}
    </span>
  )
}

// ── SituacionChip — las 6 ramas con el tono que `tonoDeSituacion` decide. El MOTIVO no va acá:
// lo pinta el widget con `AvisoChip` (texto completo, jamás truncado — BR-8/BR-9), junto al dot.
interface SituacionChipProps {
  situacion: SituacionCatalogo
}

export function SituacionChip({ situacion }: SituacionChipProps) {
  return (
    <span className={cn("pf-chip", "pf-chip-situacion", tonoDeSituacion(situacion))}>
      {etiquetaDeSituacion(situacion)}
    </span>
  )
}

// ── CanalChip — AG-D11 FIRMADA: dos entradas del índice que comparten `source` son CANALES del
// mismo contenido, no dos arneses. La fila sigue siendo la entrada (es lo instalable) y este chip
// evita que alguien crea que son dos. Descriptivo, sin tono.
interface CanalChipProps {
  otros?: string[] | undefined
}

export function CanalChip({ otros }: CanalChipProps) {
  if (!otros || otros.length === 0) return null
  return (
    <span className="pf-chip pf-chip-canal">
      mismo contenido que <span className="mono">{otros.join(", ")}</span>
    </span>
  )
}

// ── ViaChip — BR-3: un cruce DÉBIL de identidad se declara como tal. La UI no presenta
// «cruzado por el registry de la copia» como si el origen estuviera declarado.
interface ViaChipProps {
  via?: string | undefined
  /** el nombre actual de la fila (destino del rename). */
  nombre: string
  /** `nombre_anterior` del wire: de dónde vino el rename. */
  nombreAnterior?: string[] | undefined
}

export function ViaChip({ via, nombre, nombreAnterior }: ViaChipProps) {
  if (via === "faceta-registry") {
    return (
      <span className="pf-chip pf-chip-via">
        origen no declarado — cruzado por el registry de la copia
      </span>
    )
  }
  if (via === "rename") {
    const viejo = nombreAnterior?.[0] ?? "(nombre anterior desconocido)"
    return (
      <span className="pf-chip pf-chip-via">
        renombrado: <span className="mono">{viejo}</span> → <span className="mono">{nombre}</span>
      </span>
    )
  }
  // `home-declarado` o sin dato: no hay nada que declarar (identidad resuelta) — no se pinta un
  // chip vacío, mismo criterio que AvisoChip.
  return null
}
