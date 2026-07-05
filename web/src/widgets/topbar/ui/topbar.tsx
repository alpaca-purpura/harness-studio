import { selectActive, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"

// Topbar shows where the active session is parked (breadcrumb empresa / arnés / frente
// / view + "quedaste en …") and the ⌘K affordance to invoke the conversation dock.
export function Topbar() {
  const active = useSessions(selectActive)
  const chatOpen = useSessions((s) => s.chatOpen)
  const toggleChat = useSessions((s) => s.toggleChat)

  if (!active) {
    return (
      <div className="flex min-h-[44px] flex-none items-center border-b border-border bg-card px-4" />
    )
  }

  const multi = useSessions.getState().sessions.filter((s) => s.arnes === active.arnes).length > 1

  return (
    <div className="flex min-h-[44px] flex-none items-center gap-2.5 border-b border-border bg-card px-4 py-2">
      <div className="flex flex-wrap items-center gap-1.5 text-[12.5px]">
        <span className="text-muted-foreground">{active.empresa ?? "—"}</span>
        <span className="text-muted-foreground">/</span>
        <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-secondary px-2 py-0.5 font-mono text-[11.5px]">
          {active.arnes} ▾
        </span>
        {multi && (
          <span className="text-[11px] text-muted-foreground">· frente «{active.frente}»</span>
        )}
        <span className="text-muted-foreground">/</span>
        <b className="font-semibold">{active.view}</b>
      </div>

      {active.parked && (
        <span className="rounded-md border border-dashed border-input px-2 py-0.5 font-mono text-[10.5px] text-muted-foreground">
          quedaste en: {active.parked}
        </span>
      )}

      <button
        type="button"
        onClick={toggleChat}
        className={cn(
          "ml-auto inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-[12.5px] font-semibold",
          chatOpen
            ? "border border-border bg-secondary text-muted-foreground"
            : "border border-primary bg-accent-soft text-primary",
        )}
      >
        <span>✦</span>
        {chatOpen ? "Conversando" : "Conversar"}
        <kbd className="rounded border border-current px-1 font-mono text-[10px] opacity-80">
          ⌘K
        </kbd>
      </button>
    </div>
  )
}
