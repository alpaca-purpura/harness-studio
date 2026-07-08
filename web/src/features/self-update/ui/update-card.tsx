import type { PasoEstado, SelfUpdateReport, UpdateEstado, VersionInfo } from "../model/types"

// UpdateCard — la tarjeta «Versión y actualización» de la vista Ajustes (RF-101..103,
// mockup v2 firmado, casos 01–05). UI PURA (patrón Inspector): recibe la identidad
// (RF-107), el estado de la máquina y el último reporte por props; JAMÁS fetchea —
// el transporte y el polling viven en la página (fe-transporte-independiente).

interface UpdateCardProps {
  /** identidad del daemon (GET /api/version); null = aún cargando. */
  version: VersionInfo | null
  /** GET /api/version falló — estado honesto, la tarjeta lo dice tal cual. */
  versionError?: string | undefined
  estado: UpdateEstado
  /** último reporte del POST (checklist REAL; presente en error/exito/ya-al-dia/reiniciando). */
  reporte?: SelfUpdateReport | undefined
  /** error fuera del reporte: 409 en vuelo, red caída, timeout del polling (RF-105). */
  mensaje?: string | undefined
  onUpdate: () => void
}

// dotClass mapea el veredicto REAL del paso al dot del mockup (done/fail/doing).
// «agendado» pinta doing mientras se espera el reinicio y done una vez confirmado.
function dotClass(estado: PasoEstado, updateEstado: UpdateEstado): string {
  switch (estado) {
    case "ok":
      return "done"
    case "fallo":
      return "fail"
    case "agendado":
      return updateEstado === "exito" ? "done" : "doing"
    default:
      return ""
  }
}

// Checklist — SOLO con los veredictos del reporte (RF-104): sin reporte no hay pasos.
function Checklist({ reporte, estado }: { reporte: SelfUpdateReport; estado: UpdateEstado }) {
  return (
    <ul className="uc-steps">
      {reporte.pasos.map((p) => (
        <li key={p.paso} className={dotClass(p.estado, estado)}>
          <span className="uc-dot" />
          <span>
            <b>{p.paso}</b>{" "}
            <span className="uc-mono">
              — {p.detalle || (p.estado === "no-corrido" ? "no corrido" : p.estado)}
            </span>
          </span>
        </li>
      ))}
    </ul>
  )
}

export function UpdateCard({
  version,
  versionError,
  estado,
  reporte,
  mensaje,
  onUpdate,
}: UpdateCardProps) {
  const ocupado = estado === "actualizando" || estado === "reiniciando"
  const noEscribible = version !== null && !version.escribible
  const sinRepo = version !== null && version.repo === ""
  const deshabilitado = ocupado || version === null || noEscribible || sinRepo

  // El porqué SIEMPRE viaja en el title del botón disabled (design §A11y).
  let motivo: string | undefined
  if (noEscribible) motivo = "Migra a ~/.local/bin para actualizar sin sudo"
  else if (sinRepo) motivo = "Configura --repo / ARNESIA_REPO al daemon"
  else if (version === null) motivo = "Esperando la identidad del daemon (GET /api/version)"

  const pasoFallido = reporte?.pasos.find((p) => p.estado === "fallo")

  return (
    <section className="uc-card" aria-label="Versión y actualización">
      <h3>Versión y actualización</h3>

      {versionError ? (
        <p className="uc-critbox">GET /api/version falló — {versionError}</p>
      ) : version === null ? (
        <div className="uc-kv">
          <span className="uc-k">daemon</span>
          <span className="uc-v">
            <span className="uc-mut">leyendo /api/version…</span>
          </span>
        </div>
      ) : (
        <>
          <div className="uc-kv">
            <span className="uc-k">daemon</span>
            <span className="uc-v">
              {version.huella === "dev" ? (
                <>
                  arnesia · <span className="uc-mut">versión no embebida (dev)</span>
                </>
              ) : (
                <>
                  arnesia · <span className="uc-mut">huella</span> {version.huella}
                  {version.fecha ? ` · ${version.fecha}` : ""}
                </>
              )}
            </span>
          </div>
          <div className="uc-kv">
            <span className="uc-k">instalado en</span>
            <span className="uc-v">
              {version.instalado_en}{" "}
              {version.escribible ? (
                <span className="uc-pill ok">espacio de usuario · sin sudo</span>
              ) : (
                <span className="uc-pill">root · actualizar pide sudo</span>
              )}
            </span>
          </div>
          <div className="uc-kv">
            <span className="uc-k">origen</span>
            <span className="uc-v">
              {sinRepo ? (
                <span className="uc-mut">repo no configurado (arranca sin --repo)</span>
              ) : (
                <>
                  {version.repo}{" "}
                  <span className="uc-mut">
                    (build local — configurado por flag, jamás del request)
                  </span>
                </>
              )}
            </span>
          </div>
        </>
      )}

      {reporte && <Checklist reporte={reporte} estado={estado} />}

      {estado === "exito" && (
        <p className="uc-okbox">
          Actualizado a <span className="uc-mono">{version?.huella}</span> — daemon reiniciado, UI
          reconectada.
        </p>
      )}
      {estado === "ya-al-dia" && (
        <p className="uc-okbox">
          Ya estás al día — la huella del repo (
          <span className="uc-mono">{reporte?.huella_nueva}</span>) es la del binario corriendo.
        </p>
      )}
      {estado === "error" && pasoFallido && (
        <p className="uc-critbox">
          {pasoFallido.paso} falló — {pasoFallido.detalle}
        </p>
      )}
      {estado === "error" && mensaje && <p className="uc-critbox">{mensaje}</p>}
      {estado === "error" && pasoFallido && pasoFallido.paso === "build" && (
        <p className="uc-note">
          El binario instalado NO se tocó (el rename solo ocurre tras verificar el build). Corrige
          el repo y reintenta.
        </p>
      )}

      {noEscribible && (
        <p className="uc-warnbox">
          Este binario vive en <b>{version?.instalado_en}</b> (sin permiso de escritura): la app NO
          puede reemplazarlo sin privilegios. Migración (una sola vez):{" "}
          <span className="uc-mono">install -m755 &lt;repo&gt;/bin/arnesia ~/.local/bin/</span> —
          ~/.local/bin precede en PATH; desde ahí el botón funciona sin sudo.
        </p>
      )}

      <button
        type="button"
        className="uc-btn"
        disabled={deshabilitado}
        title={motivo}
        onClick={onUpdate}
      >
        {estado === "actualizando"
          ? "Actualizando…"
          : estado === "reiniciando"
            ? "Reiniciando…"
            : estado === "error"
              ? "Reintentar"
              : "Actualizar desde el repo"}
      </button>

      {estado === "reiniciando" && (
        <p className="uc-note">
          El daemon se reinicia (re-exec) — la UI reconecta sola por /api/version y confirma la
          huella nueva.
        </p>
      )}
      {(estado === "idle" || estado === "ya-al-dia") && !noEscribible && !sinRepo && (
        <p className="uc-note">
          El self-update compila el árbol local (requiere go + pnpm — feature de operador) y
          reemplaza el binario en espacio de usuario: cero sudo. Releases remotos llegan con el tren
          (KIT-06).
        </p>
      )}
      {estado === "actualizando" && (
        <p className="uc-note">
          Cada paso puede fallar y se muestra al terminar — jamás un spinner mudo.
        </p>
      )}
    </section>
  )
}
