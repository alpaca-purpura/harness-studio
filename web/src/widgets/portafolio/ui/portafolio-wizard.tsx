import type { ChangeEvent, KeyboardEvent } from "react"
import { useEffect, useId, useRef, useState } from "react"
import {
  AvisoChip,
  type Candidato,
  DerivaChip,
  gruposCandidatosDe,
  identificadorDe,
  TipoInstalacionChip,
} from "@/entities/portafolio"
import { trapTabKeyDown } from "@/shared/lib/focus-trap"

// PortafolioWizard — superficie 2 del Portafolio (plan §2.6/§3 T6, G3/G5/G8). Props puras:
// CERO transporte (fe-transporte-independiente) — el POST /escaneos y POST /proyectos reales
// los hace la página (T7); acá solo se llaman los callbacks. Reusa el primitivo de a11y de T5
// (shared/lib/focus-trap.ts, S1-D18: sin @base-ui-components/react/dialog por el conflicto de
// portal con `within(canvasElement)` — misma razón, mismo patrón que portafolio-drawer.tsx).

export interface PortafolioWizardProps {
  abierto: boolean
  onClose: () => void
  estado: "fuente" | "escaneando" | "candidatos" | "agregando"
  /** motivo 400 del backend, textual — jamás un candidato fantasma (G5). */
  error?: string | undefined
  candidatos?: Candidato[] | undefined
  /** badge «ya en el portafolio» (S1-D10) — checkbox sigue HABILITADO. */
  clavesExistentes: ReadonlySet<string>
  onEscanear: (path: string) => void
  /** aborta el fetch en curso (S1-D9: AbortSignal en la página). */
  onCancelarEscaneo: () => void
  onAgregar: (elegidos: string[]) => void
  /** undefined fuera de Tauri — el radio «Elegir carpeta» queda disabled+tooltip (S1-D23). */
  onElegirCarpeta?: (() => Promise<string | undefined>) | undefined
}

/** Cómo el usuario carga la ruta del proyecto en el paso Fuente (S1-D23, supersede S1-D9). */
type ModoFuente = "escribir" | "elegir"

const TOOLTIP_S2 = "próximo · S2"
const TOOLTIP_WEB = "solo disponible en la app de escritorio"

// S1-D19 (T6, criterio delegado por el ticket): Esc/✕ NO cierran mientras estado==="agregando"
// — el POST /proyectos está en vuelo y cancelar a mitad de un POST es más riesgoso que un
// botón bloqueado un instante (mismo espíritu que UpdateCard, que tampoco deja cerrar a mitad
// de un self-update). En cualquier otro paso, cancelar = CERO efectos (S1-D9): el widget no
// persiste nada por su cuenta.
function bloqueadoParaCierre(estado: PortafolioWizardProps["estado"]): boolean {
  return estado === "agregando"
}

// ── PasoFuente — 2 niveles de radiogroup, ambos con look ButtonGroup (CSS, S1-D23): Nivel 1
// fuente (Carpeta local activo / Repositorio GitHub disabled+S2) y Nivel 2 modo (Escribir ruta /
// Elegir carpeta — Elegir carpeta disabled+TOOLTIP_WEB fuera de Tauri, jamás oculto en
// silencio). El input es UN SOLO elemento siempre presente: editable en modo "escribir",
// `readOnly` en modo "elegir" (la ruta la pone el picker, mostrada arriba solo-lectura). El
// botón trigger «Elegir carpeta…» solo aparece en modo "elegir" — deliberadamente separado del
// radio (evita disparar un diálogo nativo del SO como efecto secundario de seleccionar un radio,
// mal patrón de a11y; también permite reintentar tras cancelar el picker, algo que un click
// repetido sobre un radio ya marcado no puede disparar). «Escanear» exige ruta no-vacía además
// del flag `disabled` de todo el paso. Reusado en el paso 1, dentro de "escaneando" (disabled) y
// como retry embebido en 0-hallazgos/error (candidatos) — no hay callback propio de "volver a
// fuente" en el contrato (§2.6 cerrado): re-mostrar este mismo formulario ES la forma de "vuelve
// a paso 1" sin inventar una prop nueva. ──
function PasoFuente({
  path,
  onPathChange,
  onEscanear,
  onElegirCarpeta,
  modo,
  onModoChange,
  disabled,
}: {
  path: string
  onPathChange: (v: string) => void
  onEscanear: () => void
  onElegirCarpeta?: (() => void) | undefined
  modo: ModoFuente
  onModoChange: (m: ModoFuente) => void
  disabled: boolean
}) {
  const elegirNoDisponible = !onElegirCarpeta
  const escanearDeshabilitado = disabled || path.trim() === ""

  return (
    <div className="pf-wizard-fuente">
      <span className="pf-faceta-label">Origen</span>
      <div className="pf-wizard-radios" role="radiogroup" aria-label="Origen del proyecto">
        <label className="pf-wizard-radio">
          <input type="radio" name="pf-wizard-fuente-radio" checked readOnly disabled={disabled} />
          Carpeta local
        </label>
        <label className="pf-wizard-radio" title={TOOLTIP_S2}>
          <input type="radio" name="pf-wizard-fuente-radio" disabled title={TOOLTIP_S2} />
          Repositorio GitHub
        </label>
      </div>
      <span className="pf-faceta-label">Cómo cargar la ruta</span>
      <div className="pf-wizard-radios" role="radiogroup" aria-label="Cómo indicar la carpeta">
        <label className="pf-wizard-radio">
          <input
            type="radio"
            name="pf-wizard-modo-radio"
            checked={modo === "escribir"}
            disabled={disabled}
            onChange={() => onModoChange("escribir")}
          />
          Escribir ruta
        </label>
        <label className="pf-wizard-radio" title={elegirNoDisponible ? TOOLTIP_WEB : undefined}>
          <input
            type="radio"
            name="pf-wizard-modo-radio"
            checked={modo === "elegir"}
            disabled={disabled || elegirNoDisponible}
            title={elegirNoDisponible ? TOOLTIP_WEB : undefined}
            onChange={() => onModoChange("elegir")}
          />
          Elegir carpeta
        </label>
      </div>
      <div className="pf-wizard-path-row">
        <input
          type="text"
          className="pf-wizard-input"
          aria-label="Ruta del proyecto"
          placeholder={
            modo === "elegir" ? "ninguna carpeta elegida todavía" : "~/Proyectos/mi-arnes"
          }
          value={path}
          disabled={disabled}
          readOnly={modo === "elegir"}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onPathChange(e.target.value)}
        />
        {modo === "elegir" && onElegirCarpeta && (
          <button
            type="button"
            className="pf-btn-secundario"
            disabled={disabled}
            onClick={onElegirCarpeta}
          >
            Elegir carpeta…
          </button>
        )}
      </div>
      <button
        type="button"
        className="pf-btn-primary pf-wizard-escanear"
        disabled={escanearDeshabilitado}
        onClick={onEscanear}
      >
        Escanear
      </button>
    </div>
  )
}

// ── CandidatoFila — identificador mono+nombre · v<version>|v? · registry resuelto|«origen
// desconocido» · tipo/canónico · DerivaChip · AvisoChip; badge «ya en el portafolio» con
// checkbox HABILITADO (S1-D10); es_canonico con copy propio (NO se pinta como espejo). El
// identificador sigue la cadena id → scope → «(sin id)» (S1-D26: sin manifiesto, el scope —
// ruta relativa o remote — es lo único que distingue N tarjetas de un monorepo); un candidato
// sin id lleva el chip «sin manifiesto» que explica el porqué. ──
function CandidatoFila({
  candidato,
  yaPresente,
  elegido,
  onToggle,
}: {
  candidato: Candidato
  yaPresente: boolean
  elegido: boolean
  onToggle: () => void
}) {
  const idMostrado = identificadorDe(candidato.identidad)
  const version = candidato.instalacion.origen.version

  return (
    <li className="pf-wizard-candidato">
      <label className="pf-wizard-candidato-check">
        <input type="checkbox" checked={elegido} onChange={onToggle} />
        <span className="mono">{idMostrado}</span>
        {candidato.nombre && <span className="pf-mut">{candidato.nombre}</span>}
        {!candidato.identidad.id && <span className="pf-chip">sin manifiesto</span>}
        {yaPresente && <span className="pf-chip pf-chip-ya-presente">ya en el portafolio</span>}
      </label>
      <div className="pf-wizard-candidato-meta">
        <span className="pf-mut">v{version ?? "?"}</span>
        {candidato.instalacion.origen.registry ? (
          <span className="mono">{candidato.instalacion.origen.registry}</span>
        ) : (
          <span className="pf-mut">origen desconocido</span>
        )}
        {candidato.es_canonico ? (
          <span className="pf-chip pf-chip-canonico">
            checkout editable — se registra como canónico, no como instalación
          </span>
        ) : (
          <TipoInstalacionChip tipo={candidato.instalacion.tipo} />
        )}
        <DerivaChip
          estado={candidato.instalacion.deriva}
          detalle={candidato.instalacion.deriva_detalle}
        />
        <AvisoChip aviso={candidato.instalacion.aviso} />
      </div>
    </li>
  )
}

// ── PasoCandidatos — 3 ramas honestas (G5): error 400 textual · 0 hallazgos + retry ·
// checklist real. El contador del botón sigue a `elegidos` (local, ver PortafolioWizard). ──
function PasoCandidatos({
  error,
  candidatos,
  clavesExistentes,
  elegidos,
  onToggle,
  onAgregar,
  path,
  onPathChange,
  onEscanear,
  onElegirCarpeta,
  modo,
  onModoChange,
}: {
  error: string | undefined
  candidatos: Candidato[]
  clavesExistentes: ReadonlySet<string>
  elegidos: ReadonlySet<string>
  onToggle: (clave: string) => void
  onAgregar: () => void
  path: string
  onPathChange: (v: string) => void
  onEscanear: () => void
  onElegirCarpeta?: (() => void) | undefined
  modo: ModoFuente
  onModoChange: (m: ModoFuente) => void
}) {
  if (error) {
    return (
      <div className="pf-wizard-error">
        <p role="alert" className="pf-error">
          {error}
        </p>
        <PasoFuente
          path={path}
          onPathChange={onPathChange}
          onEscanear={onEscanear}
          onElegirCarpeta={onElegirCarpeta}
          modo={modo}
          onModoChange={onModoChange}
          disabled={false}
        />
      </div>
    )
  }

  if (candidatos.length === 0) {
    return (
      <div className="pf-wizard-vacio">
        <p>
          No encontré arneses instalados aquí. Busco proyectos con <code>.claude/</code> poblado,
          plugins materializados en <code>.claude/plugins/</code> o habilitados por Claude Code — si
          esta carpeta es la fuente de un arnés/plugin en sí misma (no un proyecto que lo instaló),
          este asistente no la reconoce.
        </p>
        <PasoFuente
          path={path}
          onPathChange={onPathChange}
          onEscanear={onEscanear}
          onElegirCarpeta={onElegirCarpeta}
          modo={modo}
          onModoChange={onModoChange}
          disabled={false}
        />
      </div>
    )
  }

  // Agrupación por subcarpeta (monorepo, S1-D26): con UN solo grupo la lista queda plana
  // (proyecto simple, cero ruido); con varios, cada grupo lleva su heading muted — el walker
  // desciende hasta 4 niveles (C-P-11) y sin esto N hallazgos anidados eran indistinguibles.
  const grupos = gruposCandidatosDe(candidatos)
  const listaDe = (cs: Candidato[]) => (
    <ul className="pf-wizard-lista-candidatos">
      {cs.map((c) => (
        <CandidatoFila
          key={c.clave}
          candidato={c}
          yaPresente={clavesExistentes.has(c.clave)}
          elegido={elegidos.has(c.clave)}
          onToggle={() => onToggle(c.clave)}
        />
      ))}
    </ul>
  )

  return (
    <div className="pf-wizard-candidatos">
      {grupos.length === 1 && grupos[0] ? (
        listaDe(grupos[0].candidatos)
      ) : (
        <div className="pf-wizard-grupos">
          {grupos.map((g) => (
            <section
              key={g.grupo}
              className="pf-wizard-grupo"
              aria-label={`hallazgos en ${g.grupo}`}
            >
              <span className="pf-faceta-label mono">{g.grupo}</span>
              {listaDe(g.candidatos)}
            </section>
          ))}
        </div>
      )}
      <p className="pf-mut pf-wizard-nota">
        espejos read-only — la única copia editable es el canónico.
      </p>
      <button
        type="button"
        className="pf-btn-primary"
        disabled={elegidos.size === 0}
        onClick={onAgregar}
      >
        Agregar {elegidos.size} al portafolio
      </button>
    </div>
  )
}

export function PortafolioWizard({
  abierto,
  onClose,
  estado,
  error,
  candidatos,
  clavesExistentes,
  onEscanear,
  onCancelarEscaneo,
  onAgregar,
  onElegirCarpeta,
}: PortafolioWizardProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const closeBtnRef = useRef<HTMLButtonElement>(null)
  const titleId = useId()
  // path/elegidos: estado LOCAL de UI (no transporte) — el contrato §2.6 no trae un valor
  // controlado de path ni de selección; el widget lo posee, igual que `confirmando` en el
  // drawer (T5). Default de `elegidos` = vacío: nada se agrega sin elección explícita del
  // usuario (BR-8, "honestidad > limpieza" — ni siquiera un "seleccionar todo" implícito).
  const [path, setPath] = useState("")
  const [elegidos, setElegidos] = useState<ReadonlySet<string>>(new Set())
  // modo — S1-D23: lifted igual que `path` (misma razón: el contrato §2.6 no trae un valor
  // controlado). Default SIEMPRE "escribir" (funciona con y sin Tauri); se resetea solo porque
  // el widget entero desmonta al cerrar el wizard (la página lo renderiza condicional).
  const [modo, setModo] = useState<ModoFuente>("escribir")

  useEffect(() => {
    if (abierto) closeBtnRef.current?.focus()
  }, [abierto])

  const cierreBloqueado = bloqueadoParaCierre(estado)

  function handleKeyDown(e: KeyboardEvent<HTMLDivElement>) {
    if (e.key === "Escape") {
      e.stopPropagation()
      if (!cierreBloqueado) onClose()
      return
    }
    if (dialogRef.current) trapTabKeyDown(dialogRef.current, e)
  }

  function toggleElegido(clave: string) {
    setElegidos((prev) => {
      const next = new Set(prev)
      if (next.has(clave)) next.delete(clave)
      else next.add(clave)
      return next
    })
  }

  // elegirCarpetaClick — llama el picker async (patrón RF-110 de AjustesView): si resuelve con
  // un path lo pone en el input. El contrato no trae un slot de error dedicado para esta rama
  // (distinto del 400 del backend, que sí viaja por `error`) — un rechazo del picker nativo
  // (cancelación del SO, permiso denegado) se ignora silenciosamente; el trigger sigue visible
  // en modo "elegir" (S1-D23) para reintentar con otro click.
  async function elegirCarpetaClick() {
    if (!onElegirCarpeta) return
    try {
      const elegido = await onElegirCarpeta()
      if (elegido) setPath(elegido)
    } catch {
      // ver comentario arriba — sin slot de error dedicado en el contrato de este widget.
    }
  }

  if (!abierto) return null

  return (
    <div
      ref={dialogRef}
      className="pf-wizard"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
      onKeyDown={handleKeyDown}
    >
      <header className="pf-wizard-head">
        <h2 id={titleId}>Agregar al portafolio</h2>
        <button
          type="button"
          ref={closeBtnRef}
          className="pf-drawer-cerrar"
          aria-label="Cerrar"
          disabled={cierreBloqueado}
          title={cierreBloqueado ? "agregando en curso — esperá a que termine" : undefined}
          onClick={onClose}
        >
          ✕
        </button>
      </header>

      {/* Paso 0 — tabs, SIEMPRE visibles (G3): Marketplace disabled+tooltip, jamás un "✓
          marketplace válido" fingido — la rama disabled no simula ningún resultado. */}
      <div className="pf-wizard-tabs" role="tablist" aria-label="Fuente del arnés">
        <button type="button" role="tab" aria-selected="true" className="pf-wizard-tab activa">
          Proyecto
        </button>
        <button
          type="button"
          role="tab"
          aria-selected="false"
          className="pf-wizard-tab"
          disabled
          title={TOOLTIP_S2}
        >
          Marketplace
        </button>
      </div>

      {estado === "fuente" && (
        <PasoFuente
          path={path}
          onPathChange={setPath}
          onEscanear={() => onEscanear(path)}
          onElegirCarpeta={onElegirCarpeta ? elegirCarpetaClick : undefined}
          modo={modo}
          onModoChange={setModo}
          disabled={false}
        />
      )}

      {estado === "escaneando" && (
        <div className="pf-wizard-escaneando">
          <PasoFuente
            path={path}
            onPathChange={setPath}
            onEscanear={() => onEscanear(path)}
            onElegirCarpeta={onElegirCarpeta ? elegirCarpetaClick : undefined}
            modo={modo}
            onModoChange={setModo}
            disabled
          />
          <div className="pf-wizard-spinner" role="status" aria-live="polite">
            <span className="pf-wizard-spinner-glifo" aria-hidden="true" />
            Escaneando…
          </div>
          <button type="button" className="pf-btn-secundario" onClick={onCancelarEscaneo}>
            Cancelar
          </button>
        </div>
      )}

      {estado === "candidatos" && (
        <PasoCandidatos
          error={error}
          candidatos={candidatos ?? []}
          clavesExistentes={clavesExistentes}
          elegidos={elegidos}
          onToggle={toggleElegido}
          onAgregar={() => onAgregar(Array.from(elegidos))}
          path={path}
          onPathChange={setPath}
          onEscanear={() => onEscanear(path)}
          onElegirCarpeta={onElegirCarpeta ? elegirCarpetaClick : undefined}
          modo={modo}
          onModoChange={setModo}
        />
      )}

      {estado === "agregando" && (
        <div className="pf-wizard-agregando">
          <button type="button" className="pf-btn-primary" disabled>
            Agregando…
          </button>
        </div>
      )}
    </div>
  )
}
