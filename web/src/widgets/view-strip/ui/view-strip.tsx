import { selectActive, useSessions, VIEWS } from "@/shared"
import { cn } from "@/shared/lib/cn"

// ViewStrip is the slim per-session view switcher (Mapa|Diag|Corridas|Tren|Hist). It
// parks the active session on the chosen view; the interiors are "próximamente".
export function ViewStrip() {
  const active = useSessions(selectActive)
  const parkView = useSessions((s) => s.parkView)

  return (
    <nav className="flex w-[46px] flex-none flex-col items-center gap-0.5 border-r border-border bg-card py-2.5">
      {VIEWS.map(([label, glyph]) => {
        const pressed = active?.view === label
        return (
          <button
            key={label}
            type="button"
            title={label}
            aria-pressed={pressed}
            onClick={() => void parkView(label)}
            className={cn(
              "flex w-[38px] flex-col items-center gap-0.5 rounded-lg py-1.5 text-[7.5px] leading-none",
              pressed
                ? "bg-accent-soft text-primary"
                : "text-muted-foreground hover:bg-secondary hover:text-foreground",
            )}
          >
            <span className="text-sm">{glyph}</span>
            {label}
          </button>
        )
      })}
    </nav>
  )
}
