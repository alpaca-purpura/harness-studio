import { SALUD_LABEL, STATUS_LABEL, selectActive, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { HealthDot, Pip } from "@/shared/ui/indicators"

// Topbar — el ÚNICO header de sesión (paquete shell-topbar-selector-arnes, TS-D1..D4 + TS-D19).
// TS-D19 le absorbe el banner suelto que vivía en WorkspaceStage (arnés grande + status + salud):
// dos headers apilados mostrando el mismo arnés eran el mismo dato partido en dos cajas que nadie
// diseñó juntas — quedan fusionados en una sola fila (arnés · status · salud · vista · Conversar).
// NO pinta empresa (el modelo del Portafolio es N:M arnés↔empresa — no hay «la» empresa de un
// arnés para una línea fija) ni «quedaste en …» (el campo `parked` no tiene productor vivo, solo
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
    <div className="flex min-h-[44px] flex-none items-center justify-between gap-3 border-b border-border bg-card px-4">
      {/* arnés (etiqueta plana) · status · salud · / · vista — min-w-0+truncate para que nunca
          empuje a Conversar fuera de la fila */}
      <div className="flex min-w-0 flex-nowrap items-center gap-2 overflow-hidden text-[12.5px]">
        <span className="flex-none inline-flex items-center rounded-md border border-dashed border-border px-2 py-0.5 font-mono text-[11.5px]">
          {active.arnes}
        </span>
        <span
          className="flex-none inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-[11px] font-semibold"
          style={{
            background: active.status === "streaming" ? "var(--c-skill)" : "var(--secondary)",
            color:
              active.status === "streaming"
                ? "var(--primary-foreground)"
                : "var(--muted-foreground)",
          }}
        >
          <Pip status={active.status} />
          {STATUS_LABEL[active.status]}
        </span>
        {active.salud && (
          <span className="flex-none inline-flex items-center gap-1 text-[11px] text-muted-foreground">
            <HealthDot salud={active.salud} />
            {SALUD_LABEL[active.salud]}
          </span>
        )}
        {multi && (
          <span className="flex-none text-[11px] text-muted-foreground">
            · frente «{active.frente}»
          </span>
        )}
        <span className="flex-none text-muted-foreground">/</span>
        <b className="truncate font-semibold">{active.view}</b>
      </div>

      <button
        type="button"
        onClick={toggleChat}
        className={cn(
          "flex-none inline-flex items-center gap-2 rounded-lg px-3 py-1.5 text-[12.5px] font-semibold",
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
