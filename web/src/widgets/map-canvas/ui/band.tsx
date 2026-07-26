import { ArnesNode, type Box } from "@/entities/arnes"
import { ACT, type SupportBand } from "../model/bands"
import { ActivationChip } from "./activation-chip"
import type { MejoraNodo } from "./lane"

// Band is a simple support block of the Base region (mockup:355-361, bandBlock): a header
// (label + activation chip + count) and a wrapped row of 232px node cells. Used for
// libreria-expertos · meta-harness · terceros · marcas-dormidas (the `base` band renders
// BaseBand instead). `marcas-dormidas` dims the whole band (RF-44).

interface BandProps {
  band: SupportBand
  nodes: Box[]
  related?: ReadonlySet<string> | undefined
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
  /** Props primitivas de la capa Mejora, ya compuestas por el widget (D18). Con la capa
   *  apagada llega `undefined` y el DOM es el de hoy. */
  mejora?: ReadonlyMap<string, MejoraNodo> | undefined
}

export function Band({ band, nodes, related, selectedId, onSelect, mejora }: BandProps) {
  const act = ACT[band.act]
  return (
    <div className={band.act === "dormida" ? "band dormida" : "band"}>
      <div className="band-hd">
        <h3>{band.label}</h3>
        <ActivationChip label={act.label} tone={act.tone} />
        <span className="count">{nodes.length}</span>
      </div>
      <div className="row">
        {nodes.map((b) => (
          <div key={b.id} className="node-w">
            <ArnesNode
              box={b}
              dim={related !== undefined && !related.has(b.id)}
              selected={b.id === selectedId}
              onSelect={onSelect}
              {...mejora?.get(b.id)}
            />
          </div>
        ))}
      </div>
    </div>
  )
}
