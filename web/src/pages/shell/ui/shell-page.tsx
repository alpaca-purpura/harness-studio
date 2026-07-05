import { useAppStore, useSessions } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { ChatDock } from "@/widgets/chat-dock"
import { SessionRail } from "@/widgets/session-rail"
import { Topbar } from "@/widgets/topbar"
import { ViewStrip } from "@/widgets/view-strip"
import { GlobalView } from "./global-view"
import { WorkspaceStage } from "./workspace-stage"

const GLOBAL_ROUTES = new Set(["portafolio", "estandar", "ajustes"])

// ShellPage is the composition root of the shell (mockup it.14). Left: the session rail
// (multisesión). Center: either a global destination or the active session's canvas
// (view strip + topbar + near-fullscreen stage) with the invoked conversation dock on
// the right. There is no router — the "route" is hash-state (HS-05).
export function ShellPage() {
  const route = useAppStore((s) => s.view)
  const chatOpen = useSessions((s) => s.chatOpen)
  const isGlobal = GLOBAL_ROUTES.has(route)

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
                className={cn(
                  "flex-none overflow-hidden border-l border-border bg-card transition-[width] duration-200",
                  chatOpen ? "w-[360px]" : "w-0",
                )}
              >
                {chatOpen && <ChatDock />}
              </aside>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
