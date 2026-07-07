import type { ReactNode } from "react"
import { cn } from "@/shared/lib/cn"

// Region is one of the map's three fixed territories (mockup:54-60, RF-10/RF-14): a tinted
// block with an uppercase header + a horizontal rule. The tint per kind (r-guardia/r-proceso/
// r-soporte) lives in map.css. The CSS keeps the historical `r-soporte` class, but the title
// and doctrine are «Base» (canónico VISION A6, design §3.1).

type RegionKind = "guardia" | "proceso" | "soporte"

export function Region({
  kind,
  title,
  children,
}: {
  kind: RegionKind
  title: string
  children: ReactNode
}) {
  return (
    <div className={cn("region", `r-${kind}`)}>
      <div className="region-hd">
        <h2>{title}</h2>
        <span className="rule" />
      </div>
      {children}
    </div>
  )
}
