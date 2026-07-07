import type { Arnes } from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"
import { type Capa, LAYERS } from "../model/layers"

// MapBar is the top bar of the map surface (galaxia `.mapbar`): the arnés picker (RF-72), the
// META strip (empresa · rol · reporta a · marketplace) and the layer tablist. It is CHROME the
// shell mounts ABOVE the pure canvas, so it stays reachable even when the current arnés fails to
// load (the picker is the escape hatch to a different arnés). The layer switcher lives here
// (signed doctrine: "conmutador de capas en la barra del mapa").

interface MapBarProps {
  arnes?: Arnes | undefined
  capa: Capa
  onCapa: (c: Capa) => void
  // Arnés picker (RF-72). Optional so stories that only show the bar can omit it.
  harnesses?: readonly { id: string; label: string }[] | undefined
  activeId?: string | undefined
  onPick?: ((id: string) => void) | undefined
}

function MetaChip({ k, v }: { k: string; v?: string | null | undefined }) {
  if (!v) return null
  return (
    <span className="rounded border border-border bg-secondary px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
      {k} <b className="font-semibold text-foreground">{v}</b>
    </span>
  )
}

export function MapBar({ arnes, capa, onCapa, harnesses, activeId, onPick }: MapBarProps) {
  // Keep the current arnés selectable even when it is not in the index (a session may point at an
  // arnés the index has not seeded) — otherwise the <select> value would not match any option.
  const inList = !activeId || (harnesses?.some((h) => h.id === activeId) ?? false)
  const options =
    harnesses && activeId && !inList
      ? [{ id: activeId, label: `${activeId} · no indexado` }, ...harnesses]
      : harnesses

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2.5">
      {options && options.length > 0 ? (
        <select
          name="arnes-picker"
          aria-label="Elegir arnés"
          value={activeId ?? ""}
          onChange={(e) => onPick?.(e.target.value)}
          className="rounded-md border border-border bg-secondary px-2 py-1 font-mono text-xs text-foreground"
        >
          {options.map((h) => (
            <option key={h.id} value={h.id}>
              {h.label}
            </option>
          ))}
        </select>
      ) : (
        <span className="font-mono text-xs text-muted-foreground">{arnes?.id ?? "—"}</span>
      )}
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
