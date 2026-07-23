import { useEffect, useState } from "react"
import { useAppStore, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { ChatDock } from "@/widgets/chat-dock"
import { SessionRail } from "@/widgets/session-rail"
import { Topbar } from "@/widgets/topbar"
import { ViewStrip } from "@/widgets/view-strip"
import { GlobalView } from "./global-view"
import { WorkspaceStage } from "./workspace-stage"

const GLOBAL_ROUTES = new Set(["portafolio", "estandar", "ajustes"])

// CH-D5: el dock se estira desde su borde izquierdo; mínimo legible, máximo 60 % de la
// ventana, y el ancho elegido persiste entre sesiones.
const DOCK_KEY = "arnesia.dock.w"
const DOCK_MIN = 300
const dockMax = () => Math.round(window.innerWidth * 0.6)
const clampDock = (w: number) => Math.min(dockMax(), Math.max(DOCK_MIN, w))

// ShellPage is the composition root of the shell (mockup it.14). Left: the session rail
// (multisesión). Center: either a global destination or the active session's canvas
// (view strip + topbar + near-fullscreen stage) with the invoked conversation dock on
// the right. There is no router — the "route" is hash-state (HS-05).
export function ShellPage() {
  const route = useAppStore((s) => s.view)
  const chatOpen = useSessions((s) => s.chatOpen)
  const isGlobal = GLOBAL_ROUTES.has(route)
  const [dockW, setDockW] = useState(() => clampDock(Number(localStorage.getItem(DOCK_KEY)) || 360))
  const [dragging, setDragging] = useState(false)

  useEffect(() => {
    if (!dragging) localStorage.setItem(DOCK_KEY, String(dockW))
  }, [dragging, dockW])

  return (
    <div className="flex h-full w-full">
      <SessionRail />

      {isGlobal ? (
        <main className="flex-1 overflow-auto bg-background">
          <GlobalView />
        </main>
      ) : (
        <>
          <ViewStrip />
          <div className="flex min-w-0 flex-1 flex-col">
            <Topbar />
            <div className="flex min-h-0 flex-1">
              <div className="flex-1 overflow-auto bg-background">
                <WorkspaceStage />
              </div>
              <aside
                style={{ width: chatOpen ? dockW : 0 }}
                className={cn(
                  "relative flex-none overflow-hidden border-l border-border bg-card",
                  !dragging && "transition-[width] duration-200",
                )}
              >
                {chatOpen && (
                  <>
                    <div
                      role="separator"
                      aria-orientation="vertical"
                      aria-label="Redimensionar la conversación"
                      aria-valuenow={dockW}
                      aria-valuemin={DOCK_MIN}
                      aria-valuemax={dockMax()}
                      tabIndex={0}
                      onKeyDown={(e) => {
                        if (e.key === "ArrowLeft") setDockW((w) => clampDock(w + 16))
                        if (e.key === "ArrowRight") setDockW((w) => clampDock(w - 16))
                      }}
                      onPointerDown={(e) => {
                        e.preventDefault()
                        e.currentTarget.setPointerCapture(e.pointerId)
                        setDragging(true)
                      }}
                      onPointerMove={(e) => {
                        if (dragging) setDockW(clampDock(window.innerWidth - e.clientX))
                      }}
                      onPointerUp={() => setDragging(false)}
                      onLostPointerCapture={() => setDragging(false)}
                      className="absolute inset-y-0 left-0 z-10 w-1 cursor-col-resize hover:bg-primary/40"
                    />
                    <ChatDock />
                  </>
                )}
              </aside>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
