import { ArnesNode, type Box } from "@/entities/arnes"

// Band is a transversal region of the map: Guardia (hooks, top) or Base (rules · knowledge · mcp,
// bottom). Nodes flow in a wrapped row. Empty is legit (a newborn arnés with no hooks yet).

interface BandProps {
  label: string
  sublabel: string
  nodes: Box[]
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function Band({ label, sublabel, nodes, selectedId, onSelect }: BandProps) {
  return (
    <section className="rounded-lg border border-border bg-card p-2.5">
      <div className="mb-2 flex items-baseline gap-2">
        <h2 className="text-sm font-semibold text-foreground">{label}</h2>
        <span className="text-xs text-muted-foreground">{sublabel}</span>
      </div>
      {nodes.length === 0 ? (
        <span className="text-xs italic text-muted-foreground">— sin nodos —</span>
      ) : (
        <div className="flex flex-wrap gap-2">
          {nodes.map((b) => (
            <div key={b.id} className="w-44">
              <ArnesNode box={b} selected={b.id === selectedId} onSelect={onSelect} />
            </div>
          ))}
        </div>
      )}
    </section>
  )
}
