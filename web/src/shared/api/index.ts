// Public API of the `shared/api` segment.
export { ApiError, api, fetchAuthToken } from "./client"
export { connectDock, type DockConnection } from "./sse"
export {
  type Dictado,
  type DisponibilidadDictado,
  type DockFrame,
  type EstadoDictado,
  type EventoDiagnostico,
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
} from "./types"
