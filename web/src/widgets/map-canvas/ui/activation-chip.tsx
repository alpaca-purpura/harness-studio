import type { CSSProperties } from "react"

// ActivationChip is the `.act-chip` pill (mockup:68): a tone-colored capsule naming a band's
// activation (siempre/condicional/bajo demanda/leído/dormida). The tone is a token var passed
// through the --ac custom property (map.css reads it for color/border/background). Empty label
// → no chip (the `mixta` case).

type ChipStyle = CSSProperties & { "--ac"?: string }

export function ActivationChip({ label, tone }: { label: string; tone: string }) {
  if (!label) return null
  const style: ChipStyle = { "--ac": tone }
  return (
    <span className="act-chip" style={style}>
      {label}
    </span>
  )
}
