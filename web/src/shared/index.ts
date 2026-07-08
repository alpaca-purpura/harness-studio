// Public API de la capa `shared` (FSD). Los consumidores importan desde aquí o desde
// segmentos públicos (`@/shared/ui/button`, `@/shared/api`), nunca de rutas internas.

export {
  ApiError,
  api,
  connectDock,
  type DockConnection,
  type DockFrame,
  type GateReport,
  GLOBAL_VIEWS,
  type HarnessSummary,
  type NewSession,
  type PermissionAsk,
  type Rol,
  SALUD_LABEL,
  type Salud,
  type ScopeNode,
  type Session,
  type SessionStatus,
  STATUS_LABEL,
  type Turn,
  VIEWS,
} from "./api"
export { type TokenName, tokens } from "./config/tokens"
export { cn } from "./lib/cn"
export { bindHashState, type Theme, useAppStore } from "./store/app-store"
export {
  selectActive,
  selectAttention,
  selectPendingPerms,
  selectScope,
  useSessions,
} from "./store/sessions-store"
export { Button, type ButtonProps, buttonVariants } from "./ui/button"
export { ComingSoon } from "./ui/coming-soon"
export { ErrorBoundary } from "./ui/error-boundary"
