// Public API of the `shared/api` segment.
export { ApiError, api, fetchAuthToken } from "./client"
export { connectDock, type DockConnection } from "./sse"
export {
  type DockFrame,
  GLOBAL_VIEWS,
  type HarnessSummary,
  type NewSession,
  type Rol,
  SALUD_LABEL,
  type Salud,
  type Session,
  type SessionStatus,
  STATUS_LABEL,
  type Turn,
  VIEWS,
} from "./types"
