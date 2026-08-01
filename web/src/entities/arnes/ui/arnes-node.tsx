import type { CSSProperties } from "react"
import { Glyph } from "@/shared/canvas"
import { cn } from "@/shared/lib/cn"
import { kindFor } from "../model/kind"
import {
  arquetipoMark,
  gateTone,
  handleFor,
  isCaja,
  isPropuesto,
  transLabel,
} from "../model/node-view"
import { alwFor, isDelPuesto } from "../model/proposals"
import type { Box } from "../model/types"

// ArnesNode is the map's node card (`.node`, mockup:91,336-351). Structure layer (MVP): the
// type is carried by the glyph (shape+color+char) + the derived handle; the name is mono,
// 2-line clamped. It DERIVES its marks: caja/transición/propuesto from REAL data (contract·
// estado·canal via node-view), origen/alw as PROPUESTA from the isolated fixture sets
// (proposals). Positional/interaction state (support·compact·dim·selected) comes from the
// canvas. Styling lives in app/styles/map.css (scoped .arnesia-map) — this emits the mockup DOM.

// NodeStyle extends CSSProperties with the per-node type-color custom property (--tc), which
// map.css reads for the left border, tint, badges and transition tag.
type NodeStyle = CSSProperties & { "--tc"?: string }
type GateStyle = CSSProperties & { "--gate"?: string }

// ── Capa «Mejora» (paquete 2026-07-24, T31) ────────────────────────────────────────────────
//
// 🔴 **`entities/arnes` NO importa `entities/telemetria`. Nunca, por ningún escape** (D18).
// `steiger fsd/no-cross-imports` corre dentro de `verify` y lo rechazaría; el escape `@x/**`
// existe pero `steiger.config.ts` declara su enforcement INCIERTO, y degradar un gate por una
// prop es mal negocio.
//
// Por eso el nodo recibe **primitivos**, y es el widget `map-canvas` —que sí puede importar las
// dos entities, porque widgets→entities es la dirección legal— el que compone `CifraCaja` en
// estas props. Lo que se duplica es el TIPO (una unión de 4 literales), que es lo que `tsc` sí
// puede cuidar; **el copy NO se duplica**: `etiquetaConfianza`/`tituloConfianza` llegan por prop
// desde el widget, que las lee de `entities/telemetria`. La igualdad literal la ata la story
// `CopyConfianzaEsUnaSola` (map-canvas.stories.tsx, candado 1 de D18).
//
// **Superset estricto (BR-M16):** todas son opcionales y sin ellas el DOM es IDÉNTICO al de hoy.
// El guardián es la story `EstructuraIntacta`.
interface MejoraProps {
  /** El monto YA formateado por `CifraUsd`/`usd()`. El nodo no formatea dinero (RF-281). */
  cifraUsd?: string | undefined
  /** Participación en el total del arnés, 0..100 entero. */
  participacionPct?: number | undefined
  confianza?: "exacta" | "por-hash" | "por-proceso" | "sin-dato" | undefined
  /** El literal del chip de confianza, resuelto por el widget. `exacta` ⇒ undefined. */
  etiquetaConfianza?: string | undefined
  tituloConfianza?: string | undefined
  /** El detector NOMBRADO de mayor ahorro. Cero o uno, nunca dos (RF-241). */
  marcaFuga?: string | undefined
  marcaFugaGrave?: boolean | undefined
  /** Por qué este nodo no lleva cifra. El texto ES el portador (RF-280), no la opacidad. */
  motivoSinDato?: string | undefined
}

interface ArnesNodeProps extends MejoraProps {
  box: Box
  // Positional/interaction flags owned by the canvas (not derivable from the node alone).
  support?: boolean | undefined
  compact?: boolean | undefined
  dim?: boolean | undefined
  selected?: boolean | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function ArnesNode({
  box,
  support,
  compact,
  dim,
  selected,
  onSelect,
  cifraUsd,
  participacionPct,
  confianza,
  etiquetaConfianza,
  tituloConfianza,
  marcaFuga,
  marcaFugaGrave,
  motivoSinDato,
}: ArnesNodeProps) {
  const k = kindFor(box.clase)
  const caja = isCaja(box)
  const puesto = isDelPuesto(box.id)
  const propuesto = isPropuesto(box)
  const alw = alwFor(box.id)
  const trans = transLabel(box)
  // Classification marks (§8.1/§8.2) — from REAL contract data; only cajas carry them.
  // Lookups TOTALES (deuda F): un facet fuera del enum degrada SU marca («?» warn que
  // nombra el valor), jamás tumba el lienzo al ErrorBoundary (§4.5).
  const arq = box.contract?.arquetipo
  const marcaArq = arq ? arquetipoMark(arq) : undefined
  const perfil = box.contract?.perfil_harness
  const gate = box.contract?.gate?.tipo
  const tonoGate = gate ? gateTone(gate) : undefined
  const style: NodeStyle = { "--tc": k.color }
  // Capa Mejora: el nodo gana alto solo cuando efectivamente hay algo que decir.
  const conMejora =
    cifraUsd !== undefined || marcaFuga !== undefined || participacionPct !== undefined
  // Compartido visible (MA-L4): la faceta `actividades` viene DERIVADA del loader (MA-T1b);
  // >1 ⇒ badge ×N con las actividades listadas. Sin catálogo la faceta no existe → cero DOM.
  const faceta = box.actividades ?? []

  return (
    <button
      type="button"
      data-node-id={box.id}
      data-proc={box.procedencia}
      aria-pressed={selected}
      onClick={() => onSelect?.(box.id)}
      style={style}
      className={cn(
        "node",
        caja && "caja",
        puesto && "puesto",
        support && "support",
        compact && "compact",
        dim && "dim",
        selected && "selected",
        motivoSinDato !== undefined && "sindato",
        conMejora && "conmejora",
      )}
    >
      {caja && <span className="caja-badge">caja</span>}
      {propuesto && <span className="prop-badge">propuesto</span>}
      {faceta.length > 1 && (
        <span
          className="xn"
          title={`usada por ${faceta.length} actividades: ${faceta.join(" · ")} (MA-L4)`}
        >
          ×{faceta.length}
        </span>
      )}
      <span className="node-top">
        <Glyph color={k.color} char={k.char} shape={k.shape} />
        <span className="node-nm">{box.nombre}</span>
      </span>
      <span className="node-cmd">{handleFor(box, alw)}</span>
      {trans && (
        <span className="node-trans" title="transición del spine que posee esta caja">
          {trans}
        </span>
      )}
      {caja && (arq || perfil || gate) && (
        <span className="node-meta">
          {marcaArq && (
            <span
              className={cn("nm-arq", !marcaArq.reconocido && "nr")}
              role="img"
              title={marcaArq.label}
              aria-label={marcaArq.label}
            >
              {marcaArq.char}
            </span>
          )}
          {perfil && (
            <span
              className={cn("nm-perfil", `p-${perfil.toLowerCase()}`)}
              title={`perfil ${perfil}`}
            >
              {perfil}
            </span>
          )}
          {tonoGate && (
            <span
              className={cn("nm-gate", tonoGate.hollow && "hollow")}
              style={{ "--gate": tonoGate.color } as GateStyle}
              role="img"
              title={tonoGate.label}
              aria-label={tonoGate.label}
            />
          )}
        </span>
      )}

      {/* ── Capa «Mejora» — SIEMPRE al final del flujo, jamás en absoluto ────────────────────
          La esquina superior derecha ya está ocupada dos veces (`.caja-badge` y `.prop-badge`
          comparten `top:8px; right:8px`, map.css:310/:329): un tercer badge ahí se superpondría
          con el de «propuesto» (J-9).

          Y el nodo ES un `<button>`: estas marcas son `<span>`, jamás controles. Anidar un botón
          en un botón es DOM inválido. Abrir la tarjeta de un punto de mejora no se hace desde
          acá — seleccionar la caja lleva el foco a su tarjeta en la lista de abajo. */}
      {cifraUsd !== undefined && (
        <span className="mej-cifra">
          <span className="mej-usd">USD</span>
          <span className="mej-monto">{cifraUsd}</span>
          {participacionPct !== undefined && <span className="mej-pct">{participacionPct} %</span>}
          {etiquetaConfianza !== undefined && (
            <span className="mej-conf" data-confianza={confianza} title={tituloConfianza}>
              {etiquetaConfianza}
            </span>
          )}
        </span>
      )}
      {marcaFuga !== undefined && (
        // D21 · la marca GRAVE va con `color: --crit` sobre `background: --card` (4,75:1 claro /
        // 5,33:1 oscuro), NO sobre `--crit-soft` (4,04:1 en claro: falla el gate). El ⚠ es
        // decorativo y el nombre del detector es el portador (RF-241): un ⚠ sin nombre obliga a
        // adivinar qué te están señalando.
        <span className={cn("mej-fuga", marcaFugaGrave && "grave")}>
          <span aria-hidden="true">⚠</span> {marcaFuga}
        </span>
      )}
      {participacionPct !== undefined && (
        <span
          className="mej-share"
          role="img"
          aria-label={`${participacionPct} % del gasto del arnés en esta ventana.`}
        >
          <span className="mej-share-fill" style={{ width: `${participacionPct}%` }} />
        </span>
      )}
      {motivoSinDato !== undefined && <span className="mej-motivo">{motivoSinDato}</span>}
    </button>
  )
}
