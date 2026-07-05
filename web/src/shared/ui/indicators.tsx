import type { Salud, SessionStatus } from "@/shared/api"
import { cn } from "@/shared/lib/cn"

// Pip is the live conductor-state dot on the rail/dock (streaming pulses, await pulses
// warn, idle dims). Styles live in shell.css.
export function Pip({ status, className }: { status: SessionStatus; className?: string }) {
  return <span className={cn("pip", `pip-${status}`, className)} aria-hidden />
}

// HealthDot is the arnés-health dot (ok|warn|crit|info).
export function HealthDot({
  salud = "info",
  className,
}: {
  salud?: Salud | undefined
  className?: string
}) {
  return <span className={cn("hdot", `hdot-${salud}`, className)} aria-hidden />
}
