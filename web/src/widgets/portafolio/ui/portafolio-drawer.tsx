import type { KeyboardEvent } from "react"
import { useEffect, useId, useRef, useState } from "react"
import {
  AvisoChip,
  DerivaChip,
  EmblemaInicial,
  type EntradaPortafolio,
  identificadorDe,
  registriesDe,
  TipoInstalacionChip,
} from "@/entities/portafolio"
import { trapTabKeyDown } from "@/shared/lib/focus-trap"

// PortafolioDrawer — superficie 3 del Portafolio (plan §2.6/§3 T5,
// G1/G2/G4/G5/G6/G7/G8 en esta superficie). Props puras: CERO transporte
// (fe-transporte-independiente) — el DELETE real lo hace la página (T7), acá solo se llama
// `onDesvincular`. El widget existe montado MIENTRAS está abierto (la página lo desmonta al
// cerrar, mismo patrón que el Inspector del Mapa) — "montar" == "abrir" para el foco inicial.

export interface PortafolioDrawerProps {
  entrada: EntradaPortafolio
  onClose: () => void
  /** undefined ⇒ sin sesión activa: los botones de observar quedan disabled + tooltip (S1-D13). */
  onObservar?: ((installPath: string) => void) | undefined
  observarDisabledMotivo?: string | undefined
  desvinculando?: boolean | undefined
  desvincularError?: string | undefined
  onDesvincular: () => void
  /** Identificar (S1-D28): escribe el sello arnes.l0.json in-situ sobre una presencia sin
   * manifiesto. undefined ⇒ el arnés ya está sellado (no se ofrece). */
  onIdentificar?: ((installPath: string, id: string, nombre: string) => void) | undefined
  identificando?: boolean | undefined
  identificarError?: string | undefined
  /** S7 (AG-D8 decisión 7, paquete 2026-07-23): abre el diálogo «Resolver origen» SOBRE este
   *  drawer — la acción se ejecuta acá, en la ficha del arnés, y el contador cruzado del plano
   *  Marketplaces es solo la segunda forma de llegar. undefined ⇒ no se ofrece. */
  onResolverOrigen?: (() => void) | undefined
  /** S8 (AG-D17, paquete 2026-07-23): `↧ Traer canónico` REAL. undefined ⇒ el botón queda
   *  `disabled` + `TOOLTIP_S2` como en el Slice 1 (superset estricto, BR-12: las stories
   *  firmadas que no pasan este callback siguen viendo exactamente lo de antes). */
  onTraerCanonico?: (() => void) | undefined
  trayendo?: boolean | undefined
  traerError?: string | undefined
}

const TOOLTIP_UPDATE = "update-check llega en Slice 4"
const TOOLTIP_S2 = "próximo · S2"
const TOOLTIP_S3 = "próximo · S3"
const TOOLTIP_S5 = "próximo · S5"

// fechaCorta — S1-D15: la foto del último escaneo, nunca tiempo-real. `agregado` es ISO 8601
// (backend); mostramos solo la fecha (honesta, determinista — nada de Intl/locale que varíe
// entre máquinas de test). Sin dato → dicho, no un placeholder silencioso.
function fechaCorta(iso: string | undefined): string {
  if (!iso) return "fecha desconocida"
  const m = /^(\d{4}-\d{2}-\d{2})/.exec(iso)
  return m?.[1] ?? iso
}

export function PortafolioDrawer({
  entrada,
  onClose,
  onObservar,
  observarDisabledMotivo,
  desvinculando,
  desvincularError,
  onDesvincular,
  onIdentificar,
  identificando,
  identificarError,
  onResolverOrigen,
  onTraerCanonico,
  trayendo,
  traerError,
}: PortafolioDrawerProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const closeBtnRef = useRef<HTMLButtonElement>(null)
  const cancelarBtnRef = useRef<HTMLButtonElement>(null)
  const titleId = useId()
  const [confirmando, setConfirmando] = useState(false)
  const [sellando, setSellando] = useState(false)
  const [idInput, setIdInput] = useState("")
  const [nombreInput, setNombreInput] = useState("")

  // Foco inicial DENTRO del drawer al montar (G8).
  useEffect(() => {
    closeBtnRef.current?.focus()
  }, [])

  // El confirm interno reenfoca su Cancelar al aparecer — el foco nunca "flota" tras el
  // cambio de vista (mismo espíritu que el foco inicial del drawer).
  useEffect(() => {
    if (confirmando) cancelarBtnRef.current?.focus()
  }, [confirmando])

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") {
      e.stopPropagation()
      onClose()
      return
    }
    if (dialogRef.current) trapTabKeyDown(dialogRef.current, e)
  }

  const idMostrado = identificadorDe(entrada.identidad)
  const empresas = entrada.empresas ?? []
  const marketplaces = registriesDe(entrada)
  const instalaciones = entrada.instalaciones ?? []
  const canonico = entrada.canonico
  const observarHabilitado = onObservar !== undefined

  // Sin sello (S1-D28): ningún manifiesto (arnes.l0.json ni plugin.json) dio un id — la
  // identidad es puro scope/huella. Es el caso que se observa en modo degradado y que
  // «Identificar» resuelve. Un arnés ya sellado siempre trae id, así que no se le ofrece.
  const sinSello = (entrada.identidad.id ?? "") === ""
  const selloTarget = instalaciones[0]?.install_path ?? canonico?.path ?? ""
  const puedeIdentificar = sinSello && onIdentificar !== undefined && selloTarget !== ""

  function observar(installPath: string) {
    onObservar?.(installPath)
  }

  function identificar() {
    if (selloTarget) onIdentificar?.(selloTarget, idInput.trim(), nombreInput.trim())
  }

  return (
    <div
      ref={dialogRef}
      className="pf-drawer"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onKeyDown={handleKeyDown}
    >
      <header className="pf-drawer-head">
        <EmblemaInicial texto={entrada.identidad.id} />
        <div className="pf-drawer-id">
          <span className="mono" id={titleId}>
            {idMostrado}
          </span>
          {entrada.nombre && <span className="pf-drawer-nombre">{entrada.nombre}</span>}
        </div>
        <button
          type="button"
          ref={closeBtnRef}
          className="pf-drawer-cerrar"
          aria-label="Cerrar"
          onClick={onClose}
        >
          ✕
        </button>
      </header>

      <p className="pf-drawer-identidad">
        {entrada.identidad.home ? (
          <span className="mono">{entrada.identidad.home}</span>
        ) : (
          <>identidad provisional (sin home) · scope: {entrada.identidad.scope ?? "desconocido"}</>
        )}
      </p>

      {/* S7 (AG-D8 decisión 7) — la reconciliación se ejecuta ACÁ, in-situ sobre la ficha del
          arnés (mismo patrón que «Identificar» del Slice 2). El contador cruzado del plano
          Marketplaces es la SEGUNDA puerta al mismo lugar, no otro lugar. Solo se ofrece cuando
          la identidad es provisional: un home ya declarado no necesita resolverse. */}
      {!entrada.identidad.home && onResolverOrigen && (
        <div className="pf-puerta pf-drawer-puerta">
          <span aria-hidden="true">◇</span>
          <div>
            Este arnés no puede decir de qué marketplace viene — sin eso, Reparar y Actualizar no
            tienen contra qué comparar.
          </div>
          <button type="button" className="pf-btn-mini acento" onClick={onResolverOrigen}>
            Resolver origen
          </button>
        </div>
      )}

      {sinSello && (
        <section className="pf-zona pf-sin-sello">
          <h3 className="pf-zona-titulo">
            Sin sello de la fábrica <span className="pf-chip-rojo">manifiesto-ausente</span>
          </h3>
          <p className="pf-mut">
            Este arnés no tiene <span className="mono">arnes.l0.json</span> — se observa en modo
            degradado. Sellalo para identificarlo y poder mejorarlo en ArnesIA.
          </p>
          {!sellando && (
            <button
              type="button"
              className="pf-btn-primary"
              disabled={!puedeIdentificar}
              title={puedeIdentificar ? undefined : "sin una instalación sellable"}
              onClick={() => setSellando(true)}
            >
              ✦ Identificar
            </button>
          )}
          {sellando && (
            <div className="pf-sello-form" role="group" aria-label="Identificar arnés">
              <label className="pf-sello-campo">
                <span>id</span>
                <input
                  className="mono"
                  value={idInput}
                  placeholder="(nombre de la carpeta)"
                  onChange={(e) => setIdInput(e.target.value)}
                />
              </label>
              <label className="pf-sello-campo">
                <span>nombre</span>
                <input
                  value={nombreInput}
                  placeholder="(= id)"
                  onChange={(e) => setNombreInput(e.target.value)}
                />
              </label>
              {identificarError && (
                <p role="alert" className="pf-error">
                  {identificarError}
                </p>
              )}
              <div className="pf-acciones">
                <button
                  type="button"
                  className="pf-btn-secundario"
                  onClick={() => setSellando(false)}
                >
                  Cancelar
                </button>
                <button
                  type="button"
                  className="pf-btn-primary"
                  disabled={identificando || !puedeIdentificar}
                  onClick={identificar}
                >
                  {identificando ? "Sellando…" : "Sellar in-situ"}
                </button>
              </div>
            </div>
          )}
        </section>
      )}

      <div className="pf-drawer-facetas">
        <div className="pf-faceta">
          <span className="pf-faceta-label">empresas</span>
          {empresas.length > 0 ? (
            <span className="pf-faceta-chips">
              {empresas.map((emp) => (
                <span key={emp} className="pf-chip">
                  {emp}
                </span>
              ))}
            </span>
          ) : (
            <span className="pf-mut">desconocida</span>
          )}
        </div>
        <div className="pf-faceta">
          <span className="pf-faceta-label">marketplaces</span>
          {marketplaces.length > 0 ? (
            <span className="pf-faceta-chips">
              {marketplaces.map((r) => (
                <span key={r} className="pf-chip mono">
                  {r}
                </span>
              ))}
            </span>
          ) : (
            <span className="pf-mut">desconocido</span>
          )}
        </div>
      </div>

      <div className="pf-callout-antidrift">
        ⚖ Ley anti-drift: se autorea SOLO el canónico — las instalaciones se observan, nunca se
        editan sueltas. Ningún botón edita una instalación, ni siquiera deshabilitado.
      </div>

      <p className="pf-update-linea" title={TOOLTIP_UPDATE}>
        update: no-verificado
      </p>

      {/* ── Zona Canónico ── */}
      <section className="pf-zona">
        <h3 className="pf-zona-titulo">
          Canónico <span className="pf-zona-badge">única copia editable</span>
        </h3>
        {canonico ? (
          <div className="pf-canonico">
            <p className="mono">{canonico.path}</p>
            <p className="pf-mut">
              v{canonico.version ?? "?"} · estado del checkout: no evaluado en este slice
            </p>
            <div className="pf-acciones">
              <button
                type="button"
                className="pf-btn-secundario"
                disabled={!observarHabilitado}
                title={observarHabilitado ? undefined : observarDisabledMotivo}
                onClick={() => observar(canonico.path)}
              >
                ◉ Abrir en Mapa
              </button>
              <button type="button" className="pf-btn-secundario" disabled title={TOOLTIP_S2}>
                ✎ Mejorar
              </button>
              <button type="button" className="pf-btn-secundario" disabled title={TOOLTIP_S3}>
                ▲ Publicar
              </button>
            </div>
          </div>
        ) : (
          <div className="pf-canonico-ausente">
            <p className="pf-mut">No tenés el canónico local — necesario para autorear.</p>
            {/* S8 (AG-D17) — DOS PUERTAS, UN ACTO: este botón y la fila del catálogo llaman al
                mismo callback (AG-D8 decisión 5, cero vocabulario nuevo). Sin `onTraerCanonico`
                queda exactamente como en el Slice 1: `disabled` + `TOOLTIP_S2` (superset
                estricto — las stories firmadas que no pasan el callback no cambian). */}
            <button
              type="button"
              className="pf-btn-secundario"
              disabled={!onTraerCanonico || trayendo}
              title={onTraerCanonico ? undefined : TOOLTIP_S2}
              aria-busy={trayendo}
              onClick={onTraerCanonico}
            >
              {trayendo ? "Trayendo…" : "↧ Traer canónico"}
            </button>
            {traerError && (
              <p role="alert" className="pf-error">
                {traerError}
              </p>
            )}
          </div>
        )}
      </section>

      {/* ── Zona Instalaciones ── */}
      <section className="pf-zona">
        <h3 className="pf-zona-titulo">
          Instalaciones <span className="pf-zona-badge">observación · read-only</span>
        </h3>
        {instalaciones.length === 0 ? (
          <p className="pf-mut">sin instalaciones registradas</p>
        ) : (
          <ul className="pf-instalaciones">
            {instalaciones.map((inst, idx) => {
              const discrepancias = inst.origen.discrepancias ?? []
              const eslabones = inst.origen.eslabones ?? []
              return (
                // biome-ignore lint/suspicious/noArrayIndexKey: sin id propio en el wire (install_path+idx alcanza, la lista no reordena).
                <li key={`${inst.install_path}-${idx}`} className="pf-instalacion">
                  <div className="pf-instalacion-head">
                    <TipoInstalacionChip tipo={inst.tipo} />
                    <span className="pf-mut">v{inst.origen.version ?? "?"}</span>
                    <DerivaChip estado={inst.deriva} detalle={inst.deriva_detalle} />
                    <AvisoChip aviso={inst.aviso} />
                  </div>
                  <p className="mono pf-instalacion-path">{inst.proyecto_path}</p>
                  <p className="mono pf-instalacion-path">{inst.install_path}</p>

                  <div className="pf-origen-copia">
                    <span className="pf-faceta-label">origen de la copia</span>{" "}
                    {inst.origen.registry ? (
                      <span className="mono">{inst.origen.registry}</span>
                    ) : (
                      <span className="pf-mut">desconocido</span>
                    )}
                    {discrepancias.length > 0 && (
                      <ul className="pf-discrepancias">
                        {discrepancias.map((d) => (
                          <li key={d}>{d}</li>
                        ))}
                      </ul>
                    )}
                    {eslabones.length > 0 && (
                      <details className="pf-trazabilidad">
                        <summary>trazabilidad</summary>
                        <ul>
                          {eslabones.map((e, i) => (
                            // biome-ignore lint/suspicious/noArrayIndexKey: eslabones no traen id (fuente+campo pueden repetirse, i desambigua).
                            <li key={`${e.fuente}-${e.campo}-${i}`} className="mono">
                              {e.fuente} · {e.campo} · {e.valor}
                            </li>
                          ))}
                        </ul>
                      </details>
                    )}
                  </div>

                  <div className="pf-acciones">
                    <button
                      type="button"
                      className="pf-btn-secundario"
                      disabled={!observarHabilitado}
                      title={observarHabilitado ? undefined : observarDisabledMotivo}
                      onClick={() => observar(inst.install_path)}
                    >
                      ◉ Observar en Mapa
                    </button>
                    <button type="button" className="pf-btn-secundario" disabled title={TOOLTIP_S5}>
                      ⚒ Reparar
                    </button>
                    <button type="button" className="pf-btn-secundario" disabled title={TOOLTIP_S5}>
                      ↩ Backport
                    </button>
                  </div>
                </li>
              )
            })}
          </ul>
        )}
      </section>

      <p className="pf-mut pf-frescura">
        datos del último escaneo — agregado {fechaCorta(entrada.agregado)}
      </p>

      <footer className="pf-drawer-footer">
        {!confirmando && (
          <button type="button" className="pf-btn-secundario" onClick={() => setConfirmando(true)}>
            ⊘ Desvincular
          </button>
        )}
        {confirmando && (
          <div className="pf-confirm-desvincular" role="group" aria-label="Confirmar desvincular">
            <p>
              Desvincular {idMostrado}: solo lo saca del portafolio — no desinstala ni borra nada
              del disco.
            </p>
            <label className="pf-confirm-checkbox">
              <input type="checkbox" disabled title={TOOLTIP_S2} /> …y borrar clon local
            </label>
            {desvincularError && (
              <p role="alert" className="pf-error">
                {desvincularError}
              </p>
            )}
            <div className="pf-acciones">
              <button
                type="button"
                ref={cancelarBtnRef}
                className="pf-btn-secundario"
                onClick={() => setConfirmando(false)}
              >
                Cancelar
              </button>
              <button
                type="button"
                className="pf-btn-primary"
                disabled={desvinculando}
                onClick={onDesvincular}
              >
                {desvinculando ? "Desvinculando…" : "Confirmar"}
              </button>
            </div>
          </div>
        )}
      </footer>
    </div>
  )
}
