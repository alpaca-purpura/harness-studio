import { Glyph } from "@/shared/canvas"
import { cn } from "@/shared/lib/cn"
import { KIND } from "../model/kind"
import type { Box } from "../model/types"

// ArnesNode is the map's node card (`.node` in the signed v3 mockup): a colored left-border by
// type, a glyph, the mono name and the type sub-label. Structure layer (MVP): type + name only;
// the Tokens/Desempeño/Proceso sub-lines wait for JSONL telemetry. Selectable (click → inspector).
//
// Colors/geometry are inline token references (var(--c-*)) — the accepted pattern in the shell
// (workspace-stage). Zero raw hex → tokens-contrato holds.

interface ArnesNodeProps {
  box: Box
  selected?: boolean | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function ArnesNode({ box, selected, onSelect }: ArnesNodeProps) {
  const k = KIND[box.clase]
  return (
    <button
      type="button"
      data-node-id={box.id}
      aria-pressed={selected}
      onClick={() => onSelect?.(box.id)}
      style={{ borderLeftColor: k.color }}
      className={cn(
        "flex w-full flex-col gap-1 rounded-md border border-l-[3px] border-border bg-secondary px-2.5 py-2 text-left transition-colors hover:border-input",
        selected && "border-ring ring-1 ring-ring",
      )}
    >
      <span className="flex items-center gap-1.5">
        <Glyph color={k.color} char={k.char} shape={k.shape} />
        <span className="truncate font-mono text-xs text-foreground">{box.nombre}</span>
      </span>
      <span className="text-xs uppercase tracking-wide text-muted-foreground">{k.label}</span>
    </button>
  )
}
