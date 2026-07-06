import { ArnesNode, type Box } from "@/entities/arnes"

// Lane is one process phase (a column in the middle band). Header = fase name + node count; body =
// the boxes that declared that fase, stacked. An empty lane still renders so the process reads
// end-to-end (a declared phase with no box yet is honest, not hidden).

interface LaneProps {
  fase: string
  nodes: Box[]
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function Lane({ fase, nodes, selectedId, onSelect }: LaneProps) {
  return (
    <section className="flex min-w-44 flex-1 flex-col rounded-lg border border-border bg-card">
      <header className="flex items-center justify-between gap-2 border-b border-border px-2.5 py-2">
        <h2 className="text-sm font-semibold text-foreground">{fase}</h2>
        <span className="rounded-full border border-border bg-secondary px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
          {nodes.length}
        </span>
      </header>
      <div className="flex flex-col gap-2 p-2.5">
        {nodes.map((b) => (
          <ArnesNode key={b.id} box={b} selected={b.id === selectedId} onSelect={onSelect} />
        ))}
      </div>
    </section>
  )
}
