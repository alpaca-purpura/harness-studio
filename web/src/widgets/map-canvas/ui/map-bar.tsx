import type { Arnes } from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"
import { type Capa, LAYERS } from "../model/layers"

// MapBar is the top bar of the map surface (galaxia `.mapbar`): the arnés id, the META strip
// (empresa · rol · reporta a · marketplace) and the layer tablist. The layer switcher lives here
// (signed doctrine: "conmutador de capas en la barra del mapa").

interface MapBarProps {
  arnes?: Arnes | undefined
  capa: Capa
  onCapa: (c: Capa) => void
}

function MetaChip({ k, v }: { k: string; v?: string | null | undefined }) {
  if (!v) return null
  return (
    <span className="rounded border border-border bg-secondary px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
      {k} <b className="font-semibold text-foreground">{v}</b>
    </span>
  )
}

export function MapBar({ arnes, capa, onCapa }: MapBarProps) {
  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2.5">
      <span className="font-mono text-xs text-muted-foreground">{arnes?.id ?? "—"}</span>
      <div className="flex flex-wrap gap-1.5">
        <MetaChip k="empresa" v={arnes?.empresa} />
        <MetaChip k="rol" v={arnes?.rol} />
        <MetaChip k="reporta a" v={arnes?.reporta_a ?? "—"} />
        <MetaChip k="⬡" v={arnes?.marketplace} />
      </div>
      <div
        role="tablist"
        aria-label="Capas del mapa"
        className="ml-auto flex gap-0.5 rounded-lg border border-border bg-secondary p-0.5"
      >
        {LAYERS.map((l) => (
          <button
            key={l.id}
            type="button"
            role="tab"
            aria-selected={capa === l.id}
            disabled={l.disabled}
            title={l.disabled ? "Necesita telemetría (indexer JSONL)" : undefined}
            onClick={() => onCapa(l.id)}
            className={cn(
              "rounded-md px-2.5 py-1 text-xs text-muted-foreground disabled:opacity-40",
              capa === l.id && "bg-card font-semibold text-foreground shadow-sm",
            )}
          >
            {l.label}
          </button>
        ))}
      </div>
    </div>
  )
}
