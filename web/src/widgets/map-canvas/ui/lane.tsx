import { Fragment } from "react"
import { ArnesNode, type Box, isCaja } from "@/entities/arnes"

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
}

export function Lane({ fase, nodes, related, selectedId, onSelect }: LaneProps) {
  // Stable caja-first sort (mockup:424): cajas keep their relative order, then the rest.
  const ordered = [...nodes].sort((a, b) => (isCaja(b) ? 1 : 0) - (isCaja(a) ? 1 : 0))
  const hasCaja = ordered.some(isCaja)
  const firstSupportIdx = ordered.findIndex((n) => !isCaja(n))

  return (
    <section className="lane">
      <div className="lane-hd">
        <h3>{fase}</h3>
        <span className="count">{nodes.length}</span>
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
              />
            </Fragment>
          )
        })}
      </div>
    </section>
  )
}
