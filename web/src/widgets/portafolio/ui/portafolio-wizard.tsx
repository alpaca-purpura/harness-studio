import type { ChangeEvent, KeyboardEvent } from "react"
import { useEffect, useId, useMemo, useRef, useState } from "react"
import type { ClaseMarketplace, Validacion } from "@/entities/marketplace"
import {
  AvisoChip,
  type Candidato,
  DerivaChip,
  type EstadoDeriva,
  filtrarCandidatos,
  gruposCandidatosDe,
  identificadorDe,
  registriesDeCandidatos,
  TipoInstalacionChip,
} from "@/entities/portafolio"
import { trapTabKeyDown } from "@/shared/lib/focus-trap"
import { BuscadorFiltro } from "@/shared/ui/buscador-filtro"
import { FiltroDisclosure } from "@/shared/ui/filtro-disclosure"
import { GrupoControl } from "@/shared/ui/grupo-control"
import { ListaLazy } from "@/shared/ui/lista-lazy"

// PortafolioWizard — superficie 2 del Portafolio (Slice 1 T6) + las dos ramas del paquete
// 2026-07-23-portafolio-agregar-marketplace (S5 corregida + S6 nueva). Props puras: CERO
// transporte (`fe-transporte-independiente`) — los POST reales los hace la página.
//
// Superset ESTRICTO (BR-12): ninguna prop existente se quita y **todas las nuevas son
// opcionales**, para que las stories firmadas del Slice 1 sigan compilando. La única conducta
// firmada que cambia la manda el diseño: la tab `Marketplace` **deja de estar `disabled`**
// (§9.2) y el `Cancelar` del estado `escaneando` se muda al pie (AG-D6: `Atrás` aborta y vuelve
// a `fuente`, `Cancelar` aborta y cierra) — dos salidas donde antes había una, no una menos.

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
  /** aborta el fetch en curso (S1-D9: AbortSignal en la página). Es lo que hace `← Atrás` desde
   *  `escaneando`: abortar y volver al paso 1 (AG-D6). */
  onCancelarEscaneo: () => void
  onAgregar: (elegidos: string[]) => void
  /** undefined fuera de Tauri — el radio «Elegir carpeta» queda disabled+tooltip (S1-D23). */
  onElegirCarpeta?: (() => Promise<string | undefined>) | undefined

  // ── AG-D6 · Atrás/Cancelar (W3/W4 de la auditoría) ──
  /** vuelve al paso anterior; la página aborta el fetch en vuelo si hay uno. **Opcional** a
   *  propósito (BR-12): sin este callback `← Atrás` queda `disabled` visible, nunca oculto. */
  onAtras?: (() => void) | undefined
  /** ruta ya escaneada, para mostrarla en `candidatos` (W3) y para que volver a `fuente` no
   *  pierda el texto. Semilla del input al montar. */
  pathEscaneado?: string | undefined

  // ── AG-D8 decisión 8 · rama Marketplace (S6) ──
  /** rama activa. Default "proyecto" ⇒ las stories firmadas no cambian. */
  rama?: "proyecto" | "marketplace" | undefined
  onRama?: ((r: "proyecto" | "marketplace") => void) | undefined
  mkEstado?: "url" | "validando" | "validado" | "registrando" | undefined
  mkValidacion?: Validacion | undefined
  mkError?: string | undefined
  /** 409 de BR-7: ya está registrado. Lleva el nombre para poder ofrecer «ir a él» sin parsear
   *  el texto del error (el FE usa `ApiError.status`, no `message`). */
  mkYaRegistrado?: { nombre: string } | undefined
  onValidarMarketplace?: ((url: string) => void) | undefined
  onRegistrarMarketplace?: ((url: string, clase: ClaseMarketplace) => void) | undefined
  // biome-ignore lint/style/useNamingConvention: nombre de prop EXIGIDO literal por design.md §9.1 (contrato de PortafolioWizardProps)
  onIrAMarketplace?: ((nombre: string) => void) | undefined
}

/** Cómo el usuario carga la ruta del proyecto en el paso Fuente (S1-D23, supersede S1-D9). */
type ModoFuente = "escribir" | "elegir"

const TOOLTIP_S2 = "próximo · S2"
const TOOLTIP_WEB = "solo disponible en la app de escritorio"
const TOOLTIP_PRIMER_PASO = "ya estás en el primer paso"
const TOOLTIP_AGREGANDO = "agregando en curso — esperá a que termine"
const TOOLTIP_REGISTRANDO = "registrando en curso — esperá a que termine"

const DERIVAS: readonly EstadoDeriva[] = ["al-hilo", "en-deriva", "deriva-no-evaluable"]

// S1-D19 (T6): Esc/✕ NO cierran mientras hay una escritura en vuelo — `agregando` (POST
// /proyectos) y, con el mismo criterio, `registrando` (POST /marketplaces). En cualquier otro
// paso, cancelar = CERO efectos (S1-D9): el widget no persiste nada por su cuenta.
function bloqueadoParaCierre(
  estado: PortafolioWizardProps["estado"],
  rama: "proyecto" | "marketplace",
  mkEstado: NonNullable<PortafolioWizardProps["mkEstado"]>,
): boolean {
  return rama === "marketplace" ? mkEstado === "registrando" : estado === "agregando"
}

// ── PasoFuente — sin cambios de fondo respecto del Slice 1 (2 niveles de radiogroup con look
// ButtonGroup, input único editable/readOnly según el modo, `Escanear` que exige ruta no-vacía).
// Lo que agrega este paquete es el rótulo de paso («Paso 1 de 2 · origen», W4 del mockup). ──
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
      <span className="pf-faceta-label">Paso 1 de 2 · origen</span>
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

// ── CandidatoFila — sin cambios respecto del Slice 1 (S1-D10/S1-D26). ──
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

// ── PasoCandidatos — 3 ramas honestas (G5): error 400 textual · 0 hallazgos + retry · checklist
// real. Este paquete agrega (AG-D3 + W4): el rótulo de paso con el CONTADOR de hallazgos, la
// ruta escaneada VISIBLE con «cambiar ruta» (que es la MISMA transición que `← Atrás`), y
// buscador + filtro + lazy sobre los hallazgos — obligatorio *porque* AG-D7 no premarca nada. ──
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
  pathEscaneado,
  onCambiarRuta,
  busqueda,
  onBusqueda,
  filtroDeriva,
  onFiltroDeriva,
  filtroOrigen,
  onFiltroOrigen,
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
  pathEscaneado: string
  onCambiarRuta: (() => void) | undefined
  busqueda: string
  onBusqueda: (q: string) => void
  filtroDeriva: ReadonlySet<EstadoDeriva>
  onFiltroDeriva: (s: ReadonlySet<EstadoDeriva>) => void
  filtroOrigen: ReadonlySet<string>
  onFiltroOrigen: (s: ReadonlySet<string>) => void
}) {
  const [abierto, setAbierto] = useState<"deriva" | "origen" | null>(null)
  const toggleAbierto = (panel: "deriva" | "origen") =>
    setAbierto((a) => (a === panel ? null : panel))

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

  const filtradas = filtrarCandidatos(candidatos, busqueda, filtroDeriva, filtroOrigen)
  const origenes = registriesDeCandidatos(candidatos)

  // Agrupación por subcarpeta (monorepo, S1-D26): con UN solo grupo la lista queda plana
  // (proyecto simple, cero ruido); con varios, cada grupo lleva su heading muted.
  const grupos = gruposCandidatosDe(filtradas)
  const listaDe = (cs: Candidato[]) => (
    <ListaLazy<Candidato>
      items={cs}
      claveDe={(c) => c.clave}
      claseLista="pf-wizard-lista-candidatos"
      claseScroll="pf-wizard-lista-scroll"
      etiquetaRestantes={(n) => `… ${n} más (se cargan al bajar)`}
      render={(c) => (
        <CandidatoFila
          candidato={c}
          yaPresente={clavesExistentes.has(c.clave)}
          elegido={elegidos.has(c.clave)}
          onToggle={() => onToggle(c.clave)}
        />
      )}
    />
  )

  return (
    <div className="pf-wizard-candidatos">
      {/* W4 — el contador de HALLAZGOS (cifra distinta de la de elegidos, a propósito). */}
      <span className="pf-faceta-label">
        Paso 2 de 2 · escaneo — encontrados {candidatos.length} arneses
      </span>
      {/* W3 — la ruta escaneada VISIBLE, y «cambiar ruta» = la MISMA transición que ← Atrás. */}
      <p className="pf-mut pf-wizard-nota">
        Escaneado: <span className="mono">{pathEscaneado}</span>
        {onCambiarRuta && (
          <>
            {" — "}
            <button type="button" className="pf-btn-mini" onClick={onCambiarRuta}>
              cambiar ruta
            </button>
          </>
        )}
      </p>

      {/* AG-D3 — buscador + filtro sobre los hallazgos. Los filtros reusan vocabulario YA
          firmado (`EstadoDeriva` literal, mismo que `DerivaChip`; `origen.registry` real): cero
          palabra nueva. */}
      <BuscadorFiltro
        busqueda={busqueda}
        onBusqueda={onBusqueda}
        placeholder="Buscar entre los hallazgos…"
        ariaLabel="Buscar entre los hallazgos"
        filtros={
          <GrupoControl
            rotulo="Filtros"
            ariaLabel="Filtros de los hallazgos"
            claseControles="pf-filtros"
          >
            <FiltroDisclosure<EstadoDeriva>
              etiqueta="Deriva"
              abierto={abierto === "deriva"}
              onToggleAbierto={() => toggleAbierto("deriva")}
              panelId="pf-wizard-filtro-deriva"
              valores={DERIVAS}
              labelDe={(d) => d}
              seleccion={filtroDeriva}
              onCambiar={onFiltroDeriva}
            />
            <FiltroDisclosure<string>
              etiqueta="Origen"
              abierto={abierto === "origen"}
              onToggleAbierto={() => toggleAbierto("origen")}
              panelId="pf-wizard-filtro-origen"
              valores={origenes}
              labelDe={(o) => o}
              seleccion={filtroOrigen}
              onCambiar={onFiltroOrigen}
              vacio="ningún hallazgo declara origen"
            />
          </GrupoControl>
        }
      />

      {filtradas.length === 0 ? (
        <div className="pf-estado-vacio">
          <p>Ningún hallazgo coincide con la búsqueda/filtros aplicados.</p>
          <button
            type="button"
            className="pf-btn-secundario"
            onClick={() => {
              onBusqueda("")
              onFiltroDeriva(new Set())
              onFiltroOrigen(new Set())
            }}
          >
            Limpiar búsqueda y filtros
          </button>
        </div>
      ) : grupos.length === 1 && grupos[0] ? (
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

// ── RamaMarketplace (S6, AG-D8 decisión 8) — SOLO REGISTRAR. Elegir arneses y traerlos pasa en el
// plano Marketplaces, que tiene espacio para un catálogo; un catálogo no cabe en un modal.
//
// BR-5/G3 hecho MÁQUINA DE ESTADOS, no un `disabled` salteable: `Registrar y ver catálogo` no
// existe en el DOM hasta que hubo un 200 real con un `marketplace.json` leído. No hay forma de
// fabricar un ✓. ──
function RamaMarketplace({
  mkEstado,
  mkValidacion,
  mkError,
  mkYaRegistrado,
  url,
  onUrl,
  clase,
  onClase,
  onValidar,
  onRegistrar,
  onIrAMarketplace,
}: {
  mkEstado: NonNullable<PortafolioWizardProps["mkEstado"]>
  mkValidacion: Validacion | undefined
  mkError: string | undefined
  mkYaRegistrado: { nombre: string } | undefined
  url: string
  onUrl: (v: string) => void
  clase: ClaseMarketplace
  onClase: (c: ClaseMarketplace) => void
  onValidar: (() => void) | undefined
  onRegistrar: (() => void) | undefined
  // biome-ignore lint/style/useNamingConvention: espeja el nombre del contrato público (design.md §9.1)
  onIrAMarketplace: ((nombre: string) => void) | undefined
}) {
  const validando = mkEstado === "validando"
  const registrando = mkEstado === "registrando"
  const yaValidado = mkEstado === "validado" || registrando

  return (
    <div className="pf-wizard-fuente pf-wizard-mk">
      <span className="pf-faceta-label">Registrar un marketplace</span>
      <p className="pf-wizard-nota">
        Acá <b>solo se registra</b>. Elegir arneses y traerlos pasa en el plano <b>Marketplaces</b>,
        que tiene espacio para un catálogo — un catálogo no cabe en un modal.
      </p>

      <div className="pf-wizard-path-row">
        <input
          type="text"
          className="pf-wizard-input"
          aria-label="URL del marketplace"
          placeholder="https://github.com/owner/repo"
          value={url}
          disabled={validando || yaValidado}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onUrl(e.target.value)}
        />
        {!yaValidado && (
          <button
            type="button"
            className="pf-btn-secundario"
            disabled={validando || url.trim() === ""}
            onClick={onValidar}
          >
            Validar
          </button>
        )}
      </div>
      <p className="pf-wizard-nota">
        se valida que exista <span className="mono">.claude-plugin/marketplace.json</span> legible;
        sin respuesta real no se pinta ningún ✓ (criterio G3).
      </p>

      {validando && (
        <div className="pf-wizard-spinner" role="status" aria-live="polite">
          <span className="pf-wizard-spinner-glifo" aria-hidden="true" />
          Validando…
        </div>
      )}

      {/* El motivo del backend, LITERAL. 400 («tu url no sirve»), 503 («no puedo mirar») y 409
          («ya está registrado») son tres cosas distintas y el texto que llega ya lo dice: el
          widget no re-interpreta ni suaviza. */}
      {mkError && (
        <p role="alert" className="pf-error">
          {mkError}
        </p>
      )}
      {mkYaRegistrado && (
        <div className="pf-wizard-ya-registrado">
          <p className="pf-mut">
            <b>{mkYaRegistrado.nombre}</b> ya está registrado — no se duplica ni se pisa lo que ya
            declaraste.
          </p>
          {onIrAMarketplace && (
            <button
              type="button"
              className="pf-btn-secundario"
              onClick={() => onIrAMarketplace(mkYaRegistrado.nombre)}
            >
              ir a él
            </button>
          )}
        </div>
      )}

      {/* E-15/C6 — el ✓ muestra lo que se LEYÓ del archivo real: nada inferido, nada inventado. */}
      {yaValidado && mkValidacion && (
        <div className="pf-wizard-validado" role="group" aria-label="Marketplace validado">
          <p className="pf-wizard-validado-titulo">
            ✓ leí <span className="mono">.claude-plugin/marketplace.json</span> de{" "}
            <span className="mono">{mkValidacion.url_canonica}</span>
          </p>
          <ul className="pf-wizard-validado-datos">
            <li>
              nombre: <span className="mono">{mkValidacion.nombre}</span>
            </li>
            <li>
              owner: {mkValidacion.owner_nombre ?? "no declarado"}
              {mkValidacion.owner_email && <> · {mkValidacion.owner_email}</>}
              {mkValidacion.owner_url && <> · {mkValidacion.owner_url}</>}
            </li>
            <li>{mkValidacion.entradas} arneses en el catálogo</li>
            <li>fuente: {mkValidacion.fuente}</li>
          </ul>

          <span className="pf-faceta-label">Clase</span>
          <div className="pf-wizard-radios" role="radiogroup" aria-label="Clase de marketplace">
            <label className="pf-wizard-radio">
              <input
                type="radio"
                name="pf-wizard-clase-radio"
                checked={clase === "propio"}
                disabled={registrando}
                onChange={() => onClase("propio")}
              />
              Propio
            </label>
            <label className="pf-wizard-radio">
              <input
                type="radio"
                name="pf-wizard-clase-radio"
                checked={clase === "referencia"}
                disabled={registrando}
                onChange={() => onClase("referencia")}
              />
              De referencia
            </label>
          </div>
          <p className="pf-wizard-nota">
            <b>Propio</b> = publicamos ahí; habilita traer / publicar / actualizar / reparar.{" "}
            <b>De referencia</b> = solo resuelve procedencia de arneses ajenos; catálogo read-only.
          </p>

          <button
            type="button"
            className="pf-btn-primary"
            disabled={registrando}
            aria-busy={registrando}
            onClick={onRegistrar}
          >
            {registrando ? "Registrando…" : "Registrar y ver catálogo"}
          </button>
        </div>
      )}
    </div>
  )
}

// ── Pie del wizard (AG-D6) — cierra W3: hoy, escaneada una ruta con hallazgos, no hay forma de
// corregirla ni re-escanear sin cerrar el wizard entero. `← Atrás` retrocede un paso; `Cancelar`
// aborta el flujo completo (y reusa `onClose`, que ya hace «cero efectos» + abort: **no se agrega
// un `onCancelar`** — serían dos nombres para un acto). El primer paso muestra `← Atrás`
// **`disabled` y VISIBLE**, nunca oculto (mismo criterio que el resto del Portafolio). ──
function Pie({
  atras,
  atrasMotivo,
  cancelarBloqueado,
  cancelarMotivo,
  onClose,
}: {
  atras: (() => void) | undefined
  atrasMotivo: string | undefined
  cancelarBloqueado: boolean
  cancelarMotivo: string | undefined
  onClose: () => void
}) {
  return (
    <div className="pf-wizard-pie">
      <button
        type="button"
        className="pf-btn-secundario"
        disabled={!atras}
        title={atrasMotivo}
        onClick={atras}
      >
        ← Atrás
      </button>
      <div className="derecha">
        <button
          type="button"
          className="pf-btn-secundario"
          disabled={cancelarBloqueado}
          title={cancelarMotivo}
          onClick={onClose}
        >
          Cancelar
        </button>
      </div>
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
  onAtras,
  pathEscaneado,
  rama,
  onRama,
  mkEstado,
  mkValidacion,
  mkError,
  mkYaRegistrado,
  onValidarMarketplace,
  onRegistrarMarketplace,
  onIrAMarketplace,
}: PortafolioWizardProps) {
  const dialogRef = useRef<HTMLDivElement>(null)
  const closeBtnRef = useRef<HTMLButtonElement>(null)
  const titleId = useId()
  // path/elegidos/modo: estado LOCAL de UI (no transporte) — el contrato no trae un valor
  // controlado. `path` se SIEMBRA con `pathEscaneado` para que volver a `fuente` desde
  // `candidatos` no pierda la ruta (E-11/W3). `elegidos` nace vacío: nada se agrega sin elección
  // explícita (AG-D7) — y sobrevive el viaje a `fuente` y de vuelta porque el widget no desmonta.
  const [path, setPath] = useState(pathEscaneado ?? "")
  const [elegidos, setElegidos] = useState<ReadonlySet<string>>(new Set())
  const [modo, setModo] = useState<ModoFuente>("escribir")
  const [busqueda, setBusqueda] = useState("")
  const [filtroDeriva, setFiltroDeriva] = useState<ReadonlySet<EstadoDeriva>>(new Set())
  const [filtroOrigen, setFiltroOrigen] = useState<ReadonlySet<string>>(new Set())
  // Rama Marketplace: la url y la clase son locales igual que `path` (el contrato no las
  // controla). La clase arranca en `propio` porque es el caso dominante, y la nota explica la
  // diferencia A LA VISTA: no es un default silencioso.
  const [mkUrl, setMkUrl] = useState("")
  const [mkClase, setMkClase] = useState<ClaseMarketplace>("propio")

  const ramaActiva = rama ?? "proyecto"
  const mkFase = mkEstado ?? "url"

  useEffect(() => {
    if (abierto) closeBtnRef.current?.focus()
  }, [abierto])

  const cierreBloqueado = bloqueadoParaCierre(estado, ramaActiva, mkFase)

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

  // elegirCarpetaClick — llama el picker async (patrón RF-110 de AjustesView). Un rechazo del
  // picker nativo se ignora silenciosamente; el trigger sigue visible para reintentar (S1-D23).
  async function elegirCarpetaClick() {
    if (!onElegirCarpeta) return
    try {
      const elegido = await onElegirCarpeta()
      if (elegido) setPath(elegido)
    } catch {
      // ver comentario arriba — sin slot de error dedicado en el contrato de este widget.
    }
  }

  // Las claves que el escaneo actual devolvió: al volver de `fuente` con un escaneo nuevo, las
  // elegidas que ya no existen se descartan del set — sin aviso, porque no hay nada que reportar
  // (la fila ya no está). §9.1: `elegidos ∩ claves(candidatos)`.
  const clavesActuales = useMemo(
    () => new Set((candidatos ?? []).map((c) => c.clave)),
    [candidatos],
  )
  const elegidosVigentes = useMemo(() => {
    if (clavesActuales.size === 0) return elegidos
    const vigentes = new Set<string>()
    for (const clave of elegidos) if (clavesActuales.has(clave)) vigentes.add(clave)
    return vigentes
  }, [elegidos, clavesActuales])

  // ── Habilitación del pie, estado por estado (§9.1 y §9.2). `← Atrás` NUNCA se oculta: cuando
  // no aplica va `disabled` con el motivo literal.
  let atras: (() => void) | undefined
  let atrasMotivo: string | undefined
  if (ramaActiva === "marketplace") {
    // §9.2 — el literal «un solo paso» del mockup se descarta: ahora hay 4 estados, y desde
    // `url` Atrás vuelve a la tab Proyecto.
    if (mkFase === "url") atras = onRama ? () => onRama("proyecto") : undefined
    else if (mkFase === "validando" || mkFase === "validado") atras = onAtras
    else atrasMotivo = TOOLTIP_REGISTRANDO
  } else if (estado === "fuente") {
    atrasMotivo = TOOLTIP_PRIMER_PASO
  } else if (estado === "escaneando") {
    // AG-D6 — abortar el fetch y volver al paso 1 es EXACTAMENTE lo que `onCancelarEscaneo` ya
    // hacía en el Slice 1: se reusa, no se inventa un callback nuevo.
    atras = onCancelarEscaneo
  } else if (estado === "candidatos") {
    atras = onAtras
  } else {
    atrasMotivo = TOOLTIP_AGREGANDO
  }

  const cancelarMotivo = cierreBloqueado
    ? ramaActiva === "marketplace"
      ? TOOLTIP_REGISTRANDO
      : TOOLTIP_AGREGANDO
    : undefined

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
          title={cancelarMotivo}
          onClick={onClose}
        >
          ✕
        </button>
      </header>

      {/* Paso 0 — tabs, SIEMPRE visibles (G3). La tab Marketplace ya NO está `disabled`: la rama
          se construyó (AG-D8 decisión 8). El tooltip «próximo · S2» de «Repositorio GitHub»
          dentro de la rama Proyecto SÍ se queda: eso sigue fuera de alcance. */}
      <div className="pf-wizard-tabs" role="tablist" aria-label="Fuente del arnés">
        <button
          type="button"
          role="tab"
          aria-selected={ramaActiva === "proyecto"}
          className={ramaActiva === "proyecto" ? "pf-wizard-tab activa" : "pf-wizard-tab"}
          onClick={onRama ? () => onRama("proyecto") : undefined}
        >
          Proyecto
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={ramaActiva === "marketplace"}
          className={ramaActiva === "marketplace" ? "pf-wizard-tab activa" : "pf-wizard-tab"}
          onClick={onRama ? () => onRama("marketplace") : undefined}
        >
          Marketplace
        </button>
      </div>

      {ramaActiva === "marketplace" ? (
        <RamaMarketplace
          mkEstado={mkFase}
          mkValidacion={mkValidacion}
          mkError={mkError}
          mkYaRegistrado={mkYaRegistrado}
          url={mkUrl}
          onUrl={setMkUrl}
          clase={mkClase}
          onClase={setMkClase}
          onValidar={onValidarMarketplace ? () => onValidarMarketplace(mkUrl) : undefined}
          onRegistrar={
            onRegistrarMarketplace ? () => onRegistrarMarketplace(mkUrl, mkClase) : undefined
          }
          onIrAMarketplace={onIrAMarketplace}
        />
      ) : (
        <>
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
            </div>
          )}

          {estado === "candidatos" && (
            <PasoCandidatos
              error={error}
              candidatos={candidatos ?? []}
              clavesExistentes={clavesExistentes}
              elegidos={elegidosVigentes}
              onToggle={toggleElegido}
              onAgregar={() => onAgregar(Array.from(elegidosVigentes))}
              path={path}
              onPathChange={setPath}
              onEscanear={() => onEscanear(path)}
              onElegirCarpeta={onElegirCarpeta ? elegirCarpetaClick : undefined}
              modo={modo}
              onModoChange={setModo}
              pathEscaneado={pathEscaneado ?? path}
              onCambiarRuta={onAtras}
              busqueda={busqueda}
              onBusqueda={setBusqueda}
              filtroDeriva={filtroDeriva}
              onFiltroDeriva={setFiltroDeriva}
              filtroOrigen={filtroOrigen}
              onFiltroOrigen={setFiltroOrigen}
            />
          )}

          {estado === "agregando" && (
            <div className="pf-wizard-agregando">
              <button type="button" className="pf-btn-primary" disabled>
                Agregando…
              </button>
            </div>
          )}
        </>
      )}

      <Pie
        atras={atras}
        atrasMotivo={atrasMotivo}
        cancelarBloqueado={cierreBloqueado}
        cancelarMotivo={cancelarMotivo}
        onClose={onClose}
      />
    </div>
  )
}
