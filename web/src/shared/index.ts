// Public API de la capa `shared` (FSD). Los consumidores importan desde aquí o desde
// segmentos públicos (`@/shared/ui/button`, `@/shared/api`), nunca de rutas internas.

export {
  ApiError,
  api,
  type Conversacion,
  type ConversacionActiva,
  type ConversacionesListado,
  connectDock,
  type Dictado,
  type DisponibilidadDictado,
  type DockConnection,
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
  type VentanaQuery,
  VIEWS,
} from "./api"
export { type TokenName, tokens } from "./config/tokens"
export { cn } from "./lib/cn"
export { instalarCazadorDeErrores, reportar } from "./lib/diagnostico"
export { isTauri } from "./lib/platform"
export { aWav, concatenar, HZ_STT, MIME_WAV, pico, remuestrear } from "./lib/wav"
export { bindHashState, type Theme, useAppStore } from "./store/app-store"
export {
  AVISO_MS,
  ETAPA_LABEL,
  type Etapa,
  type Fallo,
  MIME as DICTADO_MIME,
  mmss,
  type Resultado,
  TOPE_MS,
  useDictado,
} from "./store/dictado-store"
export { useMapLive } from "./store/map-live-store"
export {
  selectActive,
  selectAttention,
  selectConvActivaId,
  selectCtxCaliente,
  selectCtxPct,
  selectPendingPerms,
  selectScope,
  selectTituloActiva,
  useSessions,
} from "./store/sessions-store"
export { Button, type ButtonProps, buttonVariants } from "./ui/button"
export { ComingSoon } from "./ui/coming-soon"
export { ErrorBoundary } from "./ui/error-boundary"
export {
  ErrorBody,
  type ErrorBodyProps,
  Skeleton,
  type SkeletonProps,
} from "./ui/estado-carga"
