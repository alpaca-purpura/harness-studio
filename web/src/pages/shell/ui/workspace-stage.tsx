import { ComingSoon, SALUD_LABEL, STATUS_LABEL, selectActive, useSessions, VIEWS } from "@/shared"
import { HealthDot, Pip } from "@/shared/ui/indicators"

const viewGlyph = (v: string) => VIEWS.find((x) => x[0] === v)?.[1] ?? "◵"

// WorkspaceStage is the near-fullscreen canvas of the active session. v1 scope: the
// session header is real; each view's interior (Mapa/Diag/…) is "próximamente".
export function WorkspaceStage() {
  const s = useSessions(selectActive)
  if (!s) {
    return (
      <ComingSoon
        glyph="⬡"
        title="Sin sesión activa"
        note="Crea una sesión en el rail para empezar."
      />
    )
  }

  return (
    <div className="flex h-full flex-col">
      <header className="flex flex-wrap items-center gap-2.5 border-b border-border bg-card px-4 py-3">
        <span className="font-mono text-base font-bold">{s.arnes}</span>
        <span
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-semibold"
          style={{
            background: s.status === "streaming" ? "var(--c-skill)" : "var(--secondary)",
            color: s.status === "streaming" ? "#fff" : "var(--muted-foreground)",
          }}
        >
          <Pip status={s.status} />
          {STATUS_LABEL[s.status]}
        </span>
        <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
          {s.empresa} · {s.puesto} · <HealthDot salud={s.salud} />{" "}
          {s.salud ? SALUD_LABEL[s.salud] : ""}
        </span>
      </header>

      <div className="flex-1">
        <ComingSoon
          glyph={viewGlyph(s.view)}
          title={`Vista ${s.view}`}
          note={`El interior de «${s.view}» llega después. El shell, la multisesión y la conversación con Claude Code ya están vivos — abre el dock (⌘K) y pídele algo a este arnés.`}
        />
      </div>
    </div>
  )
}
