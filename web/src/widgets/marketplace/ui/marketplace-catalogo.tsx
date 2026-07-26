import { useState } from "react"
import {
  CanalChip,
  type Catalogo,
  ClaseChip,
  type EntradaCatalogo,
  type EstadoTraer,
  filtrarEntradasCatalogo,
  filtrarPorSituacion,
  rotuloDeTipoSituacion,
  SituacionChip,
  situacionesDisponibles,
  type TipoSituacion,
  textoDeLectura,
  tonoDeSituacion,
  ViaChip,
} from "@/entities/marketplace"
import { AvisoChip, DotSaludPortafolio, EmblemaInicial } from "@/entities/portafolio"
import { BuscadorFiltro } from "@/shared/ui/buscador-filtro"
import { FiltroDisclosure } from "@/shared/ui/filtro-disclosure"
import { GrupoControl } from "@/shared/ui/grupo-control"
import { ListaLazy } from "@/shared/ui/lista-lazy"

// MarketplaceCatalogo — superficies S3 (propio) y S4 (referencia, read-only) del paquete
// 2026-07-23-portafolio-agregar-marketplace. Props puras: CERO transporte.
//
// La regla central: **el widget PRESENTA la situación y la acción que el dominio calculó**
// (design.md §2 C1 + §6.3). No las re-deriva, no puede habilitar un botón que el dominio dejó
// apagado, y el tooltip de un botón deshabilitado es el literal EXACTO que vino del wire
// (`accion.motivo`). Si mañana Publicar entra en alcance, cambia UNA celda de la tabla Go y su
// caso de test — ningún widget.
//
// La otra regla central: **`entradas: null` ≠ `entradas: []`** (BR-4, normativo en el cable).
// `null` = «no pude leer» ⇒ se muestra el MOTIVO y no se afirma nada sobre cuántos arneses hay.
// `[]` = «leí, y no declara ninguno» ⇒ afirmación evidenciada, con su fecha.

export interface MarketplaceCatalogoProps {
  estado: "cargando" | "error" | "datos"
  /** motivo textual del fallo del GET del catálogo (distinto de `lectura.motivo`, que ES el
   *  veredicto de la lectura del marketplace y viaja DENTRO de un 200). */
  error?: string | undefined
  catalogo?: Catalogo | undefined
  /** inyectado para que «leído hace…» sea determinista en story=test. */
  ahora?: Date | undefined
  busqueda: string
  onBusqueda: (q: string) => void
  filtroSituacion: ReadonlySet<TipoSituacion>
  onFiltroSituacion: (s: ReadonlySet<TipoSituacion>) => void
  onVolver: () => void
  onRefrescar: () => void
  refrescando?: boolean | undefined
  /** link «ver ficha» — la clave viene del wire (`situacion.clave_portafolio`), no se re-deriva. */
  onVerFicha?: ((clavePortafolio: string) => void) | undefined
  /** S8 (AG-D17): la MISMA acción que el botón del drawer — dos puertas, un acto. */
  onTraerCanonico?: ((entrada: string) => void) | undefined
  /** estado de Traer por nombre de entrada: uno en vuelo por identidad, no global (E-78). */
  traer?: Record<string, EstadoTraer> | undefined
  /** abre el canónico que ya existe (rama 409 de BR-14). */
  onAbrirCanonico?: ((destino: string) => void) | undefined
}

const VERBO_LABEL: Record<Exclude<EntradaCatalogo["accion"]["verbo"], "">, string> = {
  "traer-canonico": "↧ Traer canónico",
  publicar: "Publicar",
  "actualizar-mi-copia": "Actualizar mi copia",
  reparar: "Reparar",
}

// ── Cabecera: de dónde venís, qué marketplace es, y CUÁNDO se leyó (BR-3). El refresco es
// EXPLÍCITO (AG-D8 decisión 3): la app es conductor, no proxy. ──
function Cabecera({
  cat,
  ahora,
  onVolver,
  onRefrescar,
  refrescando,
}: {
  cat: Catalogo
  ahora: Date
  onVolver: () => void
  onRefrescar: () => void
  refrescando: boolean
}) {
  const cuenta = cat.lectura.entradas
  return (
    <>
      <div className="cat-head">
        <button type="button" className="pf-btn-mini" onClick={onVolver}>
          ← Marketplaces
        </button>
        <h3 className="mono">{cat.repo ?? cat.marketplace}</h3>
        <ClaseChip clase={cat.clase} />
      </div>
      <p className="cat-sub">
        {cuenta !== undefined && cat.lectura.cuando && <>{cuenta} entradas de catálogo · </>}
        {textoDeLectura(cat.lectura, ahora)} ·{" "}
        <button
          type="button"
          className="pf-btn-mini"
          disabled={refrescando}
          aria-busy={refrescando}
          onClick={onRefrescar}
        >
          {refrescando ? "Refrescando…" : "↻ Refrescar"}
        </button>
        <br />
        La lectura es <b>cacheada</b>: la app es conductor, no proxy — sin red el plano sigue
        legible, y la fila siempre dice <i>cuándo</i> se leyó.
        {cat.clase === "referencia" && (
          <>
            <br />
            <b>Read-only por diseño.</b> Este marketplace existe acá para que un arnés ajeno pueda
            decir de dónde viene. «↧ Traer canónico» <b>no aplica</b> — habilitarlo sería operar
            arneses de terceros, y eso está muerto por visión.
          </>
        )}
        {cat.owner_nombre && (
          <>
            <br />
            owner: {cat.owner_nombre}
            {cat.owner_email && <> · {cat.owner_email}</>}
            {cat.owner_url && <> · {cat.owner_url}</>}
          </>
        )}
      </p>
    </>
  )
}

// ── La acción de la fila. `accion.habilitada` VIENE DEL DOMINIO: el FE nunca decide. Cuando es
// `false`, `disabled` + `title` con el literal exacto del wire (BR-10). Cuando el verbo es "",
// no se pinta ningún botón — la acción que no aplica no existe, no es un botón genérico. ──
function Accion({
  fila,
  estadoTraer,
  onTraerCanonico,
  onAbrirCanonico,
  onVerFicha,
}: {
  fila: EntradaCatalogo
  estadoTraer: EstadoTraer | undefined
  onTraerCanonico: ((entrada: string) => void) | undefined
  onAbrirCanonico: ((destino: string) => void) | undefined
  onVerFicha: ((clave: string) => void) | undefined
}) {
  const { verbo, habilitada, motivo } = fila.accion
  const clave = fila.situacion.clave_portafolio

  // §13.10 — `trayendo` / `traído` / `falló`. Un Traer resuelto reemplaza al botón: el veredicto
  // de deriva se muestra TAL CUAL salga (BR-17), incluso si es incómodo.
  if (estadoTraer?.fase === "trayendo") {
    return (
      <span className="cat-accion">
        <button type="button" className="pf-btn-mini acento" disabled aria-busy="true">
          Trayendo…
        </button>
      </span>
    )
  }
  if (estadoTraer?.fase === "traido") {
    const claveTraida = estadoTraer.clavePortafolio
    return (
      <span className="cat-accion">
        <span className="pf-mut cat-traido">
          traído a <span className="mono">{estadoTraer.destino}</span> ({estadoTraer.camino}) ·
          deriva: {estadoTraer.deriva}
        </span>
        {estadoTraer.derivaDetalle && <AvisoChip aviso={estadoTraer.derivaDetalle} />}
        {(estadoTraer.avisos ?? []).map((a) => (
          <AvisoChip key={a} aviso={a} />
        ))}
        {claveTraida && onVerFicha && (
          <button type="button" className="pf-btn-mini" onClick={() => onVerFicha(claveTraida)}>
            ver ficha
          </button>
        )}
      </span>
    )
  }
  if (estadoTraer?.fase === "fallo") {
    const yaPoblado = estadoTraer.destinoExistente
    return (
      <span className="cat-accion">
        <span className="pf-error cat-traer-error" role="alert">
          {estadoTraer.motivo}
        </span>
        <button
          type="button"
          className="pf-btn-mini acento"
          onClick={() => onTraerCanonico?.(fila.nombre)}
        >
          Reintentar
        </button>
        {yaPoblado && onAbrirCanonico && (
          <button type="button" className="pf-btn-mini" onClick={() => onAbrirCanonico(yaPoblado)}>
            abrir el canónico que ya tenés
          </button>
        )}
      </span>
    )
  }

  if (verbo === "") {
    return (
      <span className="cat-accion">
        <span className="pf-mut cat-sin-accion">
          nada que hacer
          {clave && onVerFicha && (
            <>
              {" · "}
              <button type="button" className="pf-btn-enlace" onClick={() => onVerFicha(clave)}>
                ver ficha
              </button>
            </>
          )}
        </span>
      </span>
    )
  }

  const esTraer = verbo === "traer-canonico"
  return (
    <span className="cat-accion">
      <button
        type="button"
        className="pf-btn-mini acento"
        disabled={!habilitada}
        title={habilitada ? undefined : motivo}
        onClick={esTraer && habilitada ? () => onTraerCanonico?.(fila.nombre) : undefined}
      >
        {VERBO_LABEL[verbo]}
      </button>
    </span>
  )
}

// ── Fila del catálogo: la ENTRADA DE ÍNDICE (AG-D11 FIRMADA), o sea el canal instalable. Dos
// entradas pueden compartir `source` y NO son dos arneses: cada una lleva el `CanalChip` que lo
// aclara. ──
function Fila({
  fila,
  estadoTraer,
  onTraerCanonico,
  onAbrirCanonico,
  onVerFicha,
}: {
  fila: EntradaCatalogo
  estadoTraer: EstadoTraer | undefined
  onTraerCanonico: ((entrada: string) => void) | undefined
  onAbrirCanonico: ((destino: string) => void) | undefined
  onVerFicha: ((clave: string) => void) | undefined
}) {
  const trayendo = estadoTraer?.fase === "trayendo"
  return (
    <li className="cat-fila" aria-busy={trayendo || undefined}>
      <EmblemaInicial texto={fila.nombre} />
      <span className="pf-fila-id">
        <span className="pf-fila-idline">
          <span className="mono">{fila.nombre}</span>
          {fila.estado_canal && <span className="pf-chip">{fila.estado_canal}</span>}
        </span>
        <span className="pf-fila-descr" title={fila.descripcion}>
          {fila.version ? `v${fila.version}` : "sin versión declarada"}
          {fila.descripcion ? ` · ${fila.descripcion}` : ""}
        </span>
        <span className="pf-fila-chips">
          <CanalChip otros={fila.comparte_source_con} />
          <ViaChip
            via={fila.situacion.via}
            nombre={fila.nombre}
            nombreAnterior={fila.nombre_anterior}
          />
          {(fila.aviso ?? []).map((a) => (
            <AvisoChip key={a} aviso={a} />
          ))}
        </span>
      </span>
      <span className="cat-situacion">
        <DotSaludPortafolio salud={tonoDeSituacion(fila.situacion)} />
        <SituacionChip situacion={fila.situacion} />
        {fila.situacion.motivo && <AvisoChip aviso={fila.situacion.motivo} />}
      </span>
      <Accion
        fila={fila}
        estadoTraer={estadoTraer}
        onTraerCanonico={onTraerCanonico}
        onAbrirCanonico={onAbrirCanonico}
        onVerFicha={onVerFicha}
      />
    </li>
  )
}

// ── El cuerpo y sus ramas honestas. La distinción `null` vs `[]` es la razón de ser de este
// bloque (BR-4): con `null` NO se afirma nada sobre la cantidad de arneses. ──
function Cuerpo({
  cat,
  ahora,
  busqueda,
  onBusqueda,
  filtroSituacion,
  onFiltroSituacion,
  onRefrescar,
  refrescando,
  traer,
  onTraerCanonico,
  onAbrirCanonico,
  onVerFicha,
}: {
  cat: Catalogo
  ahora: Date
  busqueda: string
  onBusqueda: (q: string) => void
  filtroSituacion: ReadonlySet<TipoSituacion>
  onFiltroSituacion: (s: ReadonlySet<TipoSituacion>) => void
  onRefrescar: () => void
  refrescando: boolean
  traer: Record<string, EstadoTraer>
  onTraerCanonico: ((entrada: string) => void) | undefined
  onAbrirCanonico: ((destino: string) => void) | undefined
  onVerFicha: ((clave: string) => void) | undefined
}) {
  const [filtroAbierto, setFiltroAbierto] = useState(false)

  // `null` = «no sé» (BR-4). NO se dice «no tiene arneses»: se dice el motivo, y se ofrece
  // reintentar. Es el mismo pass fabricado que G3 mató en el Slice 1.
  if (cat.entradas === null) {
    return (
      <div className="pf-estado-vacio">
        <p role="alert">No pude leer el catálogo de este marketplace.</p>
        <p className="pf-mut">
          {cat.lectura.motivo ?? "sin motivo declarado por el backend"}
          <br />
          Esto <b>no</b> significa que no tenga arneses: significa que no pude mirar.
        </p>
        <button
          type="button"
          className="pf-btn-primary"
          disabled={refrescando}
          onClick={onRefrescar}
        >
          {refrescando ? "Reintentando…" : "Reintentar"}
        </button>
      </div>
    )
  }

  // `[]` = «leí, y no declara ninguno». Afirmación EVIDENCIADA, con su fecha.
  if (cat.entradas.length === 0) {
    return (
      <div className="pf-estado-vacio">
        <p>Este marketplace no declara ningún arnés.</p>
        <p className="pf-mut">
          {textoDeLectura(cat.lectura, ahora)} — su <span className="mono">plugins[]</span> está
          vacío. Es una afirmación leída del archivo, no un «no sé».
        </p>
      </div>
    )
  }

  const disponibles = situacionesDisponibles(cat.entradas)
  const filtradas = filtrarPorSituacion(
    filtrarEntradasCatalogo(cat.entradas, busqueda),
    filtroSituacion,
  )

  return (
    <>
      <BuscadorFiltro
        busqueda={busqueda}
        onBusqueda={onBusqueda}
        placeholder="Buscar en el catálogo…"
        ariaLabel="Buscar en el catálogo"
        filtros={
          <GrupoControl
            rotulo="Filtros"
            ariaLabel="Filtros del catálogo"
            claseControles="pf-filtros"
          >
            <FiltroDisclosure<TipoSituacion>
              etiqueta="Situación"
              abierto={filtroAbierto}
              onToggleAbierto={() => setFiltroAbierto((a) => !a)}
              panelId="cat-filtro-panel-situacion"
              valores={disponibles}
              labelDe={rotuloDeTipoSituacion}
              seleccion={filtroSituacion}
              onCambiar={onFiltroSituacion}
              vacio="sin situaciones en el catálogo actual"
            />
          </GrupoControl>
        }
      />

      {cat.truncado !== undefined && cat.truncado > 0 && (
        <p className="pf-corruptas-banner" role="alert">
          {cat.truncado} entradas no cargadas (techo de seguridad) — el catálogo declara más filas
          de las que el lector acepta; lo que ves está completo hasta ese corte.
        </p>
      )}

      {filtradas.length === 0 ? (
        <div className="pf-estado-vacio">
          <p>Ninguna entrada del catálogo coincide con la búsqueda/filtros aplicados.</p>
          <button
            type="button"
            className="pf-btn-secundario"
            onClick={() => {
              onBusqueda("")
              onFiltroSituacion(new Set())
            }}
          >
            Limpiar búsqueda y filtros
          </button>
        </div>
      ) : (
        <ListaLazy<EntradaCatalogo>
          items={filtradas}
          claveDe={(e) => e.nombre}
          claseLista="cat-lista"
          claseScroll="cat-lista-scroll"
          etiquetaRestantes={(n) => `… ${n} más (se cargan al bajar)`}
          render={(fila) => (
            <Fila
              fila={fila}
              estadoTraer={traer[fila.nombre]}
              onTraerCanonico={onTraerCanonico}
              onAbrirCanonico={onAbrirCanonico}
              onVerFicha={onVerFicha}
            />
          )}
        />
      )}
    </>
  )
}

export function MarketplaceCatalogo({
  estado,
  error,
  catalogo,
  ahora,
  busqueda,
  onBusqueda,
  filtroSituacion,
  onFiltroSituacion,
  onVolver,
  onRefrescar,
  refrescando,
  onVerFicha,
  onTraerCanonico,
  traer,
  onAbrirCanonico,
}: MarketplaceCatalogoProps) {
  const momento = ahora ?? new Date()

  if (estado === "cargando") {
    return (
      <div className="pf-list">
        <div className="cat-head">
          <button type="button" className="pf-btn-mini" onClick={onVolver}>
            ← Marketplaces
          </button>
        </div>
        <div
          className="pf-skeleton"
          role="status"
          aria-live="polite"
          aria-label="Cargando catálogo"
        >
          <span className="pf-skeleton-fila" aria-hidden="true" />
          <span className="pf-skeleton-fila" aria-hidden="true" />
        </div>
      </div>
    )
  }

  if (estado === "error" || !catalogo) {
    return (
      <div className="pf-list">
        <div className="cat-head">
          <button type="button" className="pf-btn-mini" onClick={onVolver}>
            ← Marketplaces
          </button>
        </div>
        <div className="pf-estado-vacio">
          <p className="pf-mut" role="alert">
            No se pudo pedir el catálogo — {error ?? "motivo desconocido"}
          </p>
          <button type="button" className="pf-btn-primary" onClick={onRefrescar}>
            Reintentar
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="pf-list">
      <Cabecera
        cat={catalogo}
        ahora={momento}
        onVolver={onVolver}
        onRefrescar={onRefrescar}
        refrescando={refrescando ?? false}
      />
      <Cuerpo
        cat={catalogo}
        ahora={momento}
        busqueda={busqueda}
        onBusqueda={onBusqueda}
        filtroSituacion={filtroSituacion}
        onFiltroSituacion={onFiltroSituacion}
        onRefrescar={onRefrescar}
        refrescando={refrescando ?? false}
        traer={traer ?? {}}
        onTraerCanonico={onTraerCanonico}
        onAbrirCanonico={onAbrirCanonico}
        onVerFicha={onVerFicha}
      />
    </div>
  )
}
