import { selectActive, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"

// Topbar — breadcrumb `arnés / vista` + el afford ⌘K del dock (paquete shell-topbar-selector-arnes,
// TS-D1..D4). NO pinta empresa (el modelo del Portafolio es N:M arnés↔empresa — no hay «la» empresa
// de un arnés para una línea fija) ni «quedaste en …» (el campo `parked` no tiene productor vivo, solo
// seed hardcodeado — no se diseña sobre un dato que no existe). El arnés es una ETIQUETA de solo
// lectura (fijo de por vida de la sesión), no un control: sin ▾, sin onClick.
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
    <div className="flex min-h-[44px] flex-none flex-col gap-1.5 border-b border-border bg-card px-4 py-2">
      {/* fila 1 — breadcrumb: arnés (etiqueta plana) / vista */}
      <div className="flex flex-wrap items-center gap-1.5 text-[12.5px]">
        <span className="inline-flex items-center rounded-md border border-dashed border-border px-2 py-0.5 font-mono text-[11.5px]">
          {active.arnes}
        </span>
        {multi && (
          <span className="text-[11px] text-muted-foreground">· frente «{active.frente}»</span>
        )}
        <span className="text-muted-foreground">/</span>
        <b className="font-semibold">{active.view}</b>
      </div>

      {/* fila 2 — Conversar, alineado a la derecha (ya no compite por ancho con el breadcrumb) */}
      <button
        type="button"
        onClick={toggleChat}
        className={cn(
          "inline-flex items-center gap-2 self-end rounded-lg px-3 py-1.5 text-[12.5px] font-semibold",
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
