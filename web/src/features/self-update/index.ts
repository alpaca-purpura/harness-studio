// Public API de la feature self-update (FSD): la tarjeta pura + los tipos del wire.
// La página (composition-root) inyecta datos y callbacks; la feature jamás fetchea.

export type {
  PasoEstado,
  PasoReporte,
  SelfUpdateReport,
  UpdateEstado,
  UpdateResultado,
  VersionInfo,
} from "./model/types"
export { UpdateCard } from "./ui/update-card"
