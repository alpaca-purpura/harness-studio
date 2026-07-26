import { ArnesNode, type Box } from "@/entities/arnes"
import { ActivationChip } from "./activation-chip"
import type { MejoraNodo } from "./lane"
import { RulesSubband } from "./rules-subband"

// BaseBand is the `base` band of the Base region (mockup:363-394, RF-40/43): it splits its
// nodes into (a) the collapsible «Reglas» subband and (b) a static «Knowledge & servicios»
// subband (mcp / non-rule, «leído por skills»). The rules fold because a real arnés carries
// many; knowledge stays open.

interface BaseBandProps {
  nodes: Box[]
  related?: ReadonlySet<string> | undefined
  selectedId?: string | undefined
  /** Props primitivas de la capa Mejora, ya compuestas por el widget (D18). */
  mejora?: ReadonlyMap<string, MejoraNodo> | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function BaseBand({ nodes, related, selectedId, onSelect, mejora }: BaseBandProps) {
  const rules = nodes.filter((n) => n.clase === "rule")
  const rest = nodes.filter((n) => n.clase !== "rule")

  return (
    <div className="band base-band">
      <RulesSubband
        rules={rules}
        related={related}
        selectedId={selectedId}
        onSelect={onSelect}
        mejora={mejora}
      />
      {rest.length > 0 && (
        <div className="subband">
          <div className="subband-hd static">
            <h4>Knowledge &amp; servicios</h4>
            <span className="count">{rest.length}</span>
            <ActivationChip label="leído por skills" tone="var(--c-knowledge)" />
          </div>
          <div className="row">
            {rest.map((b) => (
              <div key={b.id} className="node-w">
                <ArnesNode
                  box={b}
                  dim={related !== undefined && !related.has(b.id)}
                  selected={b.id === selectedId}
                  {...mejora?.get(b.id)}
                  onSelect={onSelect}
                />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
