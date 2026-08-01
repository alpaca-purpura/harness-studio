import { Fragment } from "react"
import { ArnesNode, type Box, isCaja } from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"

/**
 * Un paso del procedimiento focado SIN caja en este carril (E13, MA-T5): se pinta como nodo
 * fantasma —dashed, muted, NO clickeable— para que el hueco sea visible, jamás saltado en
 * silencio. `numero` es el número de paso dentro del procedimiento (la secuencia lo incluye).
 */
export interface GhostPaso {
  numero: number
  paso: string
  artefacto?: string | undefined
}

// Lane is one process phase — a column in the Proceso region (mockup:84-89,419-434). Header =
// fase name + node count; body = the boxes that declared this fase. Order is CAJA-FIRST
// (RF-25): the caja(s) sit in row 1 so they align across lanes into the horizontal spine.
// Non-caja nodes render as `support` (indented + fainter), preceded by a hairline `lane-sep`
// before the first one (RF-24) — no «apoyo» word.

interface LaneProps {
  fase: string
  nodes: Box[]
  // When set (a node is focused), ids NOT in the set render dimmed (RF-33). undefined = no focus.
  related?: ReadonlySet<string> | undefined
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
  /**
   * Total de la fase, YA formateado (RF-244). `undefined` = la capa Mejora está apagada y el
   * `lane-hd` vuelve a tener dos hijos, exactamente como hoy.
   *
   * Un carril cuyas cajas no tienen costo atribuido muestra `sin dato`, **no `USD 0,00`**: la
   * página pasa `totalUsd={null}` y este componente lo dice. Un cero acá sería afirmar que la
   * fase no gastó nada, que es otra cosa.
   */
  totalUsd?: string | null | undefined
  /** Marcas primitivas por nodo (D18): el widget compone `CifraCaja → props`, el nodo no importa
   *  `entities/telemetria`. */
  mejora?: ReadonlyMap<string, MejoraNodo> | undefined
  /**
   * E5 (MA-T5): la actividad focada no tiene pasos en esta fase → el carril ENTERO se atenúa
   * (se VE que el spike termina antes). `undefined`/false ⇒ DOM idéntico al de hoy (MA-L3).
   */
  dimlane?: boolean | undefined
  /** Ghosts E13 de esta fase (pasos del procedimiento focado sin caja). Ausente ⇒ cero DOM. */
  ghosts?: readonly GhostPaso[] | undefined
}

/** Las props de mejora de UN nodo, ya compuestas por el widget. Espeja `MejoraProps` de
 *  `arnes-node.tsx` — el tipo se duplica (4 literales), el copy NO (D18). */
export interface MejoraNodo {
  cifraUsd?: string | undefined
  participacionPct?: number | undefined
  confianza?: "exacta" | "por-hash" | "por-proceso" | "sin-dato" | undefined
  etiquetaConfianza?: string | undefined
  tituloConfianza?: string | undefined
  marcaFuga?: string | undefined
  marcaFugaGrave?: boolean | undefined
  motivoSinDato?: string | undefined
}

export function Lane({
  fase,
  nodes,
  related,
  selectedId,
  onSelect,
  totalUsd,
  mejora,
  dimlane,
  ghosts,
}: LaneProps) {
  // Stable caja-first sort (mockup:424): cajas keep their relative order, then the rest.
  const ordered = [...nodes].sort((a, b) => (isCaja(b) ? 1 : 0) - (isCaja(a) ? 1 : 0))
  const hasCaja = ordered.some(isCaja)
  const firstSupportIdx = ordered.findIndex((n) => !isCaja(n))

  return (
    <section className={cn("lane", dimlane && "dimlane")}>
      <div className="lane-hd">
        <h3>{fase}</h3>
        {/* J-8: el `.count` se CONSERVA. El total de la fase se suma como tercer hijo, no
            sustituye al conteo de nodos — son dos cosas distintas y las dos se leen. */}
        <span className="count">{nodes.length}</span>
        {totalUsd !== undefined &&
          (totalUsd === null ? (
            <span className="lane-usd lane-usd-sindato">sin dato</span>
          ) : (
            <span className="lane-usd">
              <span className="lane-usd-pfx">USD</span> {totalUsd}
            </span>
          ))}
      </div>
      <div className="lane-body">
        {ordered.map((b, i) => {
          const support = hasCaja && !isCaja(b)
          return (
            <Fragment key={b.id}>
              {support && i === firstSupportIdx && <div className="lane-sep" />}
              <ArnesNode
                box={b}
                support={support}
                dim={related !== undefined && !related.has(b.id)}
                selected={b.id === selectedId}
                onSelect={onSelect}
                {...mejora?.get(b.id)}
              />
            </Fragment>
          )
        })}
        {/* E13 — el paso sin caja del procedimiento focado, EN la secuencia: un div (no
            button) porque no es un nodo real — no se selecciona ni abre inspector. */}
        {ghosts?.map((g) => (
          <div key={g.paso} className="node ghost" data-node-id={`ghost-${g.paso}`}>
            <span className="node-top">
              <span className="ghost-num" aria-hidden="true">
                {g.numero}
              </span>
              <span className="node-nm">{g.paso} — paso sin caja aún</span>
            </span>
            <span className="node-cmd">entrega: {g.artefacto ?? "—"}</span>
          </div>
        ))}
      </div>
    </section>
  )
}
