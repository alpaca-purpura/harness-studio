// Tipos del wire de self-update (arch/contracts/api/openapi.yaml 0.3.0-hs11).
// La feature es UI pura: estos tipos describen lo que la PÁGINA le inyecta tras
// hablar con el daemon (fe-transporte-independiente).

// VersionInfo — GET /api/version (RF-107): la identidad honesta del binario corriendo.
export interface VersionInfo {
  /** vcs.revision[:7] (+«+sucio» si el árbol estaba modificado); «dev» sin VCS info. */
  huella: string
  /** fecha del commit (solo día); vacía sin VCS info. */
  fecha: string
  /** ruta REAL del ejecutable (os.Executable del daemon). */
  instalado_en: string
  /** el proceso puede reemplazarse a sí mismo (escritura en su directorio) — RF-102. */
  escribible: boolean
  /** repo configurado al daemon (--repo / ARNESIA_REPO); vacío = RF-103. */
  repo: string
  /** vcs.modified del build corriente (decisión #7). */
  sucio: boolean
  /**
   * Identificador COMPLETO del build: `0.2.21.2607260225` = semver + sello de compilación
   * `AAMMDDHHMM`. «dev» en un binario sin identidad inyectada (RF-231).
   *
   * Es aditivo a `huella`, no su reemplazo: la huella dice QUÉ COMMIT, el sello dice QUÉ
   * COMPILACIÓN — dos builds del mismo árbol comparten huella y no comparten sello, que es
   * justo la pregunta que el operador se hace («¿corro lo último que compilé?»).
   */
  version: string
  /** el sello en formato legible (`2026-07-26 02:25`); vacío en un build de dev. */
  compilado?: string
  /** hay un build MÁS NUEVO en disco que el que corre, en una línea accionable. Vacío = al día. */
  aviso_build?: string
}

export type PasoEstado = "ok" | "fallo" | "no-corrido" | "agendado"

// PasoReporte — el veredicto REAL de un paso (RF-104): la checklist se pinta de aquí,
// jamás de pasos animados inventados (nota de honestidad del spec).
export interface PasoReporte {
  paso: string
  estado: PasoEstado
  detalle?: string
}

export type UpdateResultado = "actualizado" | "ya-al-dia" | "fallo"

// SelfUpdateReport — POST /api/self-update (RF-104): desenlace terminal + 5 pasos.
export interface SelfUpdateReport {
  resultado: UpdateResultado
  huella_nueva?: string
  pasos: PasoReporte[]
}

// UpdateEstado — la máquina de la tarjeta (design.md §Estados): idle → actualizando →
// (error | ya-al-dia | reiniciando → exito | error). Los estados no-actualizables
// (no-escribible · sin-repo) NO son estados de la máquina: se derivan de VersionInfo.
export type UpdateEstado = "idle" | "actualizando" | "reiniciando" | "exito" | "error" | "ya-al-dia"
