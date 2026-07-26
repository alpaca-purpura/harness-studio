import { useCallback, useEffect, useMemo, useState } from "react"
import type { ArtefactosMode, ConformanceResult, Graph } from "@/entities/arnes"
import { isCaja } from "@/entities/arnes"
import type {
  BucketToken,
  CifraCaja,
  EstadoDetector,
  ParidadCosto,
  PuntoMejora,
  ResumenTelemetria,
  SaludTelemetria,
  Ventana,
} from "@/entities/telemetria"
import { CAMPOS_PERSISTIDOS_HOOK, motivoSinDato } from "@/entities/telemetria"
import {
  api,
  ComingSoon,
  type HarnessSummary,
  selectActive,
  useAppStore,
  useMapLive,
  useSessions,
  VIEWS,
} from "@/shared"
import {
  type Capa,
  CapaMejoraStage,
  Inspector,
  InspectorMejora,
  MapBar,
  PoliticaDatosDialog,
} from "@/widgets/map-canvas"

const viewGlyph = (v: string) => VIEWS.find((x) => x[0] === v)?.[1] ?? "◵"

const ETIQUETA_VENTANA: Readonly<Record<Ventana, string>> = {
  "7d": "7 días",
  "30d": "30 días",
  todo: "todo",
}

/** `domain.DetalleCaja` recortado a lo que la 4ª tab necesita. Los buckets llegan como punteros:
 *  un campo AUSENTE en el JSON es «no aplica», y eso se conserva como `null` (RF-260). */
interface DetalleCajaWire {
  turnos_totales: number
  tokens?: Partial<Record<string, number>> | undefined
  paridad?: ParidadCosto | undefined
  detectores?: EstadoDetector[] | undefined
}

/** Los seis buckets, en orden fijo. **Un bucket ausente en el wire viaja `null`, no 0**: el
 *  `omitempty` de Go significa «este runtime no tiene el concepto», y un 0 sería mentira. */
const BUCKETS: readonly { id: BucketToken["id"]; etiqueta: string; campo: string }[] = [
  { id: "entrada", etiqueta: "entrada", campo: "entrada" },
  { id: "salida", etiqueta: "salida", campo: "salida" },
  { id: "cache_lectura", etiqueta: "cache · lectura", campo: "cache_lectura" },
  { id: "cache_escritura_5m", etiqueta: "cache · escritura 5 m", campo: "cache_escritura_5m" },
  { id: "cache_escritura_1h", etiqueta: "cache · escritura 1 h", campo: "cache_escritura_1h" },
  { id: "razonamiento", etiqueta: "razonamiento", campo: "razonamiento" },
]

function bucketsDe(d: DetalleCajaWire | null): BucketToken[] {
  return BUCKETS.map((b) => ({
    id: b.id,
    etiqueta: b.etiqueta,
    tokens: d?.tokens?.[b.campo] ?? null,
    // ⚠️ El wire NO manda el costo por bucket: `domain.DetalleCaja` trae `Tokens` y `Paridad`,
    // no un desglose de dinero por bucket. La columna USD queda `null` —«no aplica»— hasta que
    // el backend lo mande. Inventarla acá sería costear en el FE, que es justo lo que
    // `design.md` §1.3 prohíbe («entities presenta; no calcula»).
    costo_micros: null,
  }))
}

// WorkspaceStage is the near-fullscreen canvas of the active session (a page = composition-root).
// El header de sesión vive SOLO en `Topbar` (TS-D19) — este componente ya no pinta uno propio.
// La «Mapa» view renderiza la REAL Map surface — fetches the
// arnés graph from the daemon (transport lives here, not in the entity/canvas — fe-transporte-
// independiente) and hands it to <MapCanvas>. Other views stay «próximamente». Picker = RF-72.
export function WorkspaceStage() {
  const s = useSessions(selectActive)
  const arnesId = s?.arnes
  const isMapa = s?.view === "Mapa"

  const [viewedId, setViewedId] = useState<string | undefined>(arnesId)
  const [graph, setGraph] = useState<Graph | null>(null)
  const [loadErr, setLoadErr] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string>()
  // Reporte de conformance del arnés visto (RF-91). undefined = no disponible — el
  // inspector lo DICE; jamás se finge un «sin hallazgos» sin dato.
  const [conformance, setConformance] = useState<ConformanceResult[]>()
  // ¿El arnés visto tiene directorio registrado (S2)? Los fixtures embebidos no viven
  // en disco: sin registro, loadFuente rechaza LOCAL con el mismo mensaje honesto del
  // daemon — mismo estado, sin un 404 de red que ensucie la consola en el camino normal.
  const [registrado, setRegistrado] = useState(false)
  const [harnesses, setHarnesses] = useState<HarnessSummary[]>([])
  // Whether the portfolio has been fetched at least once — gates the graph load so a session
  // pointing at a non-indexed arnés never fires a 404 before we know the portfolio.
  const [harnessesLoaded, setHarnessesLoaded] = useState(false)
  // La capa activa. `estructura` por default; el conmutador vive en MapBar (chrome).
  const [capa, setCapa] = useState<Capa>("estructura")

  // ── Capa «Mejora» (T37) ─────────────────────────────────────────────────────────────────
  //
  // **Esta página es la ÚNICA que hace transporte** (`fe-transporte-independiente`): los
  // widgets reciben props puras. Acá viven el estado de la ventana y las tres consultas.
  const [ventana, setVentana] = useState<Ventana>("7d")
  const [mejEstado, setMejEstado] = useState<"datos" | "cargando" | "error">("cargando")
  const [mejError, setMejError] = useState<string>()
  const [resumen, setResumen] = useState<ResumenTelemetria | null>(null)
  const [cajas, setCajas] = useState<CifraCaja[]>([])
  const [puntos, setPuntos] = useState<PuntoMejora[]>([])
  // Los 7 de fuera del MVP. Alimentan la 4ª tab del inspector (RF-263), **no** el vacío de la
  // lista: ahí dirían «sin hallazgos», que es exactamente lo que RF-263 prohíbe.
  const [noMedidos, setNoMedidos] = useState<EstadoDetector[]>([])
  // Los del MVP que NO pudieron correr, con su motivo. El vacío de la lista los necesita para
  // no afirmar que corrieron los seis cuando alguno no pudo (s2-degradado).
  const [noAplican, setNoAplican] = useState<EstadoDetector[]>([])
  const [politicaAbierta, setPoliticaAbierta] = useState(false)
  // A-2 · `GET /api/telemetria/salud` estaba construido y **nadie lo consumía**: la retención se
  // hardcodeaba en 90 «(propuesto)» aunque el daemon dijera 400 firmados, y **el chip de reenvío
  // externo no se dibujaba nunca**. D13 · H-7 dicen que un estado peligroso no se esconde a la
  // derecha; no dibujarlo es peor que esconderlo — la pantalla repite «Nada de tu cuenta. Nada
  // de la conversación.» sin poder saber si los datos se están reenviando afuera.
  const [salud, setSalud] = useState<SaludTelemetria | null>(null)
  const [detalle, setDetalle] = useState<DetalleCajaWire | null>(null)
  const [detalleError, setDetalleError] = useState<string>()
  const [nonce, setNonce] = useState(0)
  // Franja Artefactos (RF-143): default auto — reposo idéntico al mapa actual, chips al
  // seleccionar. Mismo patrón de estado que `capa` (sin persistencia dura, como el resto).
  const [artefactos, setArtefactos] = useState<ArtefactosMode>("auto")

  // The picker previews any arnés; it defaults to (and resets with) the session's own arnés.
  // La selección TAMBIÉN se limpia: si la vista nueva no recarga grafo (p.ej. Diag), el
  // selectedBox viejo seguiría vivo y re-dispararía el chip de alcance sobre la sesión
  // nueva (leak RF-118 cazado por el E2E de casuística).
  useEffect(() => {
    setViewedId(arnesId)
    setSelectedId(undefined)
  }, [arnesId])

  // Load the graph of the viewed arnés when the Map view is active (RF-70). shared/api is the
  // only transport seam; the entity/canvas never fetch.
  useEffect(() => {
    if (!viewedId || !isMapa) return
    // Wait until the portfolio is known before deciding: a session may point at an arnés the
    // index doesn't hold (e.g. a persisted session to a since-removed arnés). Fetching it blindly
    // spams the console with a 404; instead we degrade to a clean "no indexado" state and let the
    // picker be the escape hatch.
    if (!harnessesLoaded) return
    let alive = true
    setGraph(null)
    setLoadErr(null)
    setSelectedId(undefined)
    setConformance(undefined)
    if (!harnesses.some((h) => h.id === viewedId)) {
      setLoadErr(`«${viewedId}» no está en el índice del daemon.`)
      return () => {
        alive = false
      }
    }
    api
      .getGraph<Graph>(viewedId)
      .then((g) => {
        if (alive) setGraph(g)
      })
      .catch((e: unknown) => {
        if (alive) setLoadErr(e instanceof Error ? e.message : String(e))
      })
    // Conformance del arnés (RF-91): un fetch por arnés; si falla queda undefined y el
    // inspector muestra el estado honesto «no disponible».
    api
      .getConformance<{ results?: ConformanceResult[] }>(viewedId)
      .then((r) => {
        if (alive) setConformance(r.results ?? [])
      })
      .catch(() => {
        if (alive) setConformance(undefined)
      })
    // Registro arnés→dir (S2): decide si la lectura de fuente puede confinarse.
    setRegistrado(false)
    api
      .listArneses()
      .then((entries) => {
        if (alive) setRegistrado(entries.some((e) => e.arnes === viewedId))
      })
      .catch(() => {
        if (alive) setRegistrado(false)
      })
    return () => {
      alive = false
    }
  }, [viewedId, isMapa, harnesses, harnessesLoaded])

  // Reindex-en-vivo (RF-187): el daemon avisó por `event: map` que el grafo del arnés visto
  // cambió tras un turno del chat → refetch del grafo SIN resetear selección ni conformance
  // (a diferencia del efecto de carga de arriba — esto es un refresh, no una navegación).
  const mapRev = useMapLive((st) => (viewedId ? (st.rev[viewedId] ?? 0) : 0))
  useEffect(() => {
    if (!viewedId || !isMapa || mapRev === 0) return
    let alive = true
    api
      .getGraph<Graph>(viewedId)
      .then((g) => {
        if (alive) setGraph(g)
      })
      .catch(() => {
        // El grafo viejo queda en pantalla; el efecto de navegación reporta errores — acá
        // un refetch fallido no borra lo que el usuario está viendo.
      })
    return () => {
      alive = false
    }
  }, [mapRev, viewedId, isMapa])

  // Índice del portafolio — necesario para detectar un arnés no indexado (loadErr, RF-70). No
  // alimenta más ningún picker (TS-D21 sacó el de MapBar).
  useEffect(() => {
    if (!isMapa) return
    let alive = true
    api
      .listHarnesses<HarnessSummary[]>()
      .then((hs) => {
        if (alive) {
          setHarnesses(hs ?? [])
          setHarnessesLoaded(true)
        }
      })
      .catch(() => {
        if (alive) {
          setHarnesses([])
          setHarnessesLoaded(true)
        }
      })
    return () => {
      alive = false
    }
  }, [isMapa])

  // Consumo del peek Portafolio→Mapa (Slice 1, plan §2.7/S1-D13): «Abrir en Mapa» del
  // Portafolio deja un id en el buzón `mapaPeek` (app-store) y aparca la sesión en Mapa
  // (parkView, hecho por PortafolioView) — acá, cuando el peek está seteado Y la vista activa
  // ES Mapa, se re-fetchea el índice (el guard de arriba solo acepta ids que ya estaban en el
  // snapshot `harnesses`; un arnés recién observado no está ahí todavía) ANTES de apuntar
  // `viewedId`, y el peek se limpia siempre — es un buzón de un solo consumo, nunca vuelve a
  // disparar solo. NO toca nada más del stage (el chip de alcance y el picker existentes
  // siguen intactos).
  const mapaPeek = useAppStore((st) => st.mapaPeek)
  const setMapaPeek = useAppStore((st) => st.setMapaPeek)
  useEffect(() => {
    if (!mapaPeek || !isMapa) return
    let alive = true
    api
      .listHarnesses<HarnessSummary[]>()
      .then((hs) => {
        if (!alive) return
        setHarnesses(hs ?? [])
        setHarnessesLoaded(true)
        setViewedId(mapaPeek)
      })
      .catch(() => {
        if (!alive) return
        setViewedId(mapaPeek)
      })
      .finally(() => {
        setMapaPeek(null)
      })
    return () => {
      alive = false
    }
  }, [mapaPeek, isMapa, setMapaPeek])

  // La ventana en instantes: el wire recibe `desde`/`hasta` en RFC3339, no un enum
  // (`ventanaDeQuery`, telemetria.go:25). Un valor ilegible da 400 en vez de devolver en
  // silencio una ventana distinta de la pedida.
  const rangoDeVentana = useCallback((v: Ventana) => {
    if (v === "todo") return {}
    const dias = v === "30d" ? 30 : 7
    return { desde: new Date(Date.now() - dias * 86_400_000).toISOString() }
  }, [])

  // Las tres consultas de la capa. Solo corren con la capa ENCENDIDA: con `estructura` activa
  // no se le pide nada al daemon, que es lo que hace que la capa apagada no cueste nada.
  // biome-ignore lint/correctness/useExhaustiveDependencies: `nonce` es el disparador de recarga (Reintentar · descartar · borrar), no un valor que el efecto lea — mismo patrón que `arnesId` en el efecto de carga del grafo.
  useEffect(() => {
    if (!viewedId || !isMapa || capa !== "mejora") return
    let alive = true
    setMejEstado("cargando")
    setMejError(undefined)
    const q = { ...rangoDeVentana(ventana), arnes: viewedId }
    Promise.all([
      api.telemetriaResumen<ResumenTelemetria>(q),
      api.telemetriaCajas<{ cajas?: CifraCaja[] }>(viewedId, q),
      api.telemetriaMejoras<{
        puntos?: PuntoMejora[]
        no_aplican?: EstadoDetector[]
        no_medidos?: EstadoDetector[]
      }>(viewedId, q),
    ])
      .then(([r, c, m]) => {
        if (!alive) return
        // 🔴 C-2 · el resumen viaja CRUDO. Colapsarlo acá mirando `r.corridas > 0` era el
        // defecto: fuera de S1 ese campo es 0 por construcción y la franja decía «nunca corrió»
        // sobre un arnés con dinero medido. Quién muestra qué lo decide `vistaCapaMejora()`,
        // una sola vez, para los cinco bloques.
        setResumen(r ?? null)
        setCajas(c.cajas ?? [])
        setPuntos(m.puntos ?? [])
        setNoAplican(m.no_aplican ?? [])
        setNoMedidos(m.no_medidos ?? [])
        setMejEstado("datos")
      })
      .catch((e: unknown) => {
        if (!alive) return
        setMejError(e instanceof Error ? e.message : String(e))
        setMejEstado("error")
      })
    return () => {
      alive = false
    }
  }, [viewedId, isMapa, capa, ventana, rangoDeVentana, nonce])

  // biome-ignore lint/correctness/useExhaustiveDependencies: `nonce` es el disparador de recarga (Reintentar · borrar), no un valor que el efecto lea.
  useEffect(() => {
    if (!isMapa || capa !== "mejora") return
    let alive = true
    api
      .telemetriaSalud<SaludTelemetria>()
      .then((d) => {
        if (alive) setSalud(d)
      })
      .catch(() => {
        // Sin `/salud` no se inventa una retención ni se afirma que el reenvío está apagado:
        // los dos bloques dicen que el dato no llegó.
        if (alive) setSalud(null)
      })
    return () => {
      alive = false
    }
  }, [isMapa, capa, nonce])

  // El detalle de la caja seleccionada — el cuerpo de la 4ª tab. Se pide SOLO cuando hay caja
  // seleccionada y la capa está encendida: es un drill-down, no parte de la carga del Mapa.
  useEffect(() => {
    if (!viewedId || !isMapa || capa !== "mejora" || !selectedId) {
      setDetalle(null)
      setDetalleError(undefined)
      return
    }
    let alive = true
    api
      .telemetriaDetalleCaja<DetalleCajaWire>(viewedId, selectedId, {
        ...rangoDeVentana(ventana),
      })
      .then((d) => {
        if (alive) {
          setDetalle(d)
          setDetalleError(undefined)
        }
      })
      .catch((e: unknown) => {
        // 🔴 C-3 · un GET que falla es estado de TRANSPORTE, no dato. Tragarlo hacía que la 4ª
        // tab afirmara ocho cosas falsas —«0 corridas», seis «no aplica en este runtime» y
        // «catálogo sin construir»— con la nota «"No aplica" no es 0» desplegada EN DEFENSA de
        // la mentira, mientras el nodo de al lado mostraba USD 1,08.
        if (alive) {
          setDetalle(null)
          setDetalleError(e instanceof Error ? e.message : String(e))
        }
      })
    return () => {
      alive = false
    }
  }, [viewedId, isMapa, capa, selectedId, ventana, rangoDeVentana])

  // The node the inspector shows (S3), read straight from the loaded graph — real, complete data.
  const selectedBox = useMemo(
    () => (selectedId ? graph?.nodos.find((n) => n.id === selectedId) : undefined),
    [graph, selectedId],
  )

  // RF-111 (decisión #1): la selección del Mapa se OFRECE como chip de alcance del chat —
  // solo cuando el arnés visto ES el de la sesión activa (el alcance pertenece a UNA
  // sesión, RF-118). El chip es removible desde el Dock; deseleccionar no lo borra.
  const setScope = useSessions((st) => st.setScope)
  useEffect(() => {
    if (!selectedBox || viewedId !== arnesId) return
    setScope({
      nodeId: selectedBox.id,
      clase: selectedBox.clase,
      fuentePath: selectedBox.fuente_path,
    })
  }, [selectedBox, viewedId, arnesId, setScope])

  // Lectura de la fuente real del nodo (RF-93) — el transporte vive en la página; el
  // inspector recibe el callback (fe-transporte-independiente). Un arnés sin directorio
  // registrado rechaza LOCAL con el mensaje del daemon (sin 404 de red en consola).
  const loadFuente = useCallback(
    (nodeId: string) => {
      if (!viewedId) return Promise.reject(new Error("sin arnés activo"))
      if (!registrado)
        return Promise.reject(
          new Error(
            "el arnés no tiene directorio registrado — carga la carpeta (PUT /api/arneses/{id}) para leer su fuente",
          ),
        )
      return api.getNodeFuente(viewedId, nodeId)
    },
    [viewedId, registrado],
  )

  if (!s) {
    return (
      <ComingSoon
        glyph="⬡"
        title="Sin sesión activa"
        note="Crea una sesión en el rail para empezar."
      />
    )
  }

  return (
    <div className="flex h-full flex-col">
      <div className="min-h-0 flex-1">
        {!isMapa ? (
          <ComingSoon
            glyph={viewGlyph(s.view)}
            title={`Vista ${s.view}`}
            note={`El interior de «${s.view}» llega después. El shell, la multisesión y la conversación con Claude Code ya están vivos — abre el dock (⌘K) y pídele algo a este arnés.`}
          />
        ) : (
          // MapBar stays ABOVE the canvas at all times, including the load-error state. Texto
          // de solo lectura siempre (TS-D21: "1 sesión = 1 arnés" no tiene excepción de error —
          // ni siquiera para recuperarse se cambia de arnés inline, se abre otra sesión), salvo
          // el link de "volver" al espiar otro arnés vía "Abrir en Mapa" (GAP-1, peek).
          <div className="flex h-full flex-col">
            <MapBar
              arnes={graph?.arnes}
              capa={capa}
              onCapa={setCapa}
              activeId={viewedId}
              onPick={setViewedId}
              ownId={arnesId}
              artefactos={artefactos}
              onArtefactos={setArtefactos}
            />
            {/* 🔴 La COMPOSICIÓN vive en `CapaMejoraStage`, no acá. Los cuatro críticos de la
                auditoría del Tramo B nacieron de componerla en `pages/`, que es el único lugar
                del repo sin stories — y por eso el candado de D24 no los vio. Esta página hace
                transporte y estado; qué muestra cada bloque lo decide un widget con stories. */}
            <CapaMejoraStage
              capa={capa}
              graph={graph}
              selectedId={selectedId}
              onSelect={setSelectedId}
              artefactos={artefactos}
              estado={mejEstado}
              resumen={resumen}
              cajas={cajas}
              puntos={puntos}
              noAplican={noAplican}
              error={mejError}
              ventana={ventana}
              onVentana={setVentana}
              retencionDias={salud?.retencion_dias}
              retencionPropuesta={salud?.retencion_propuesta}
              forwardDestino={
                salud?.forward ? (salud.forward_destino ?? "destino no declarado") : undefined
              }
              onPolitica={() => setPoliticaAbierta(true)}
              onReintentar={() => setNonce((n) => n + 1)}
              onDescartar={() => setNonce((n) => n + 1)}
              onProponer={() => setNonce((n) => n + 1)}
              cuerpoAlternativo={
                loadErr ? (
                  <ComingSoon
                    glyph="⚠"
                    title="No se pudo cargar el arnés"
                    note={`El daemon no devolvió el grafo de «${viewedId}». Abrí otra sesión para ver un arnés distinto. ${loadErr}`}
                  />
                ) : !graph ? (
                  <ComingSoon
                    glyph="⬡"
                    title={`Cargando ${viewedId}…`}
                    note="Leyendo el grafo del arnés."
                  />
                ) : undefined
              }
              inspector={
                selectedBox ? (
                  <Inspector
                    box={selectedBox}
                    onClose={() => setSelectedId(undefined)}
                    graph={graph ?? undefined}
                    onSelect={setSelectedId}
                    conformance={conformance}
                    loadFuente={loadFuente}
                    mejora={
                      capa === "mejora" ? (
                        <InspectorMejora
                          estado={detalleError !== undefined ? "error" : "datos"}
                          error={detalleError}
                          esCaja={isCaja(selectedBox)}
                          motivoNoCaja={motivoSinDato(selectedBox.clase)}
                          ventanaLabel={ETIQUETA_VENTANA[ventana]}
                          corridas={detalle?.turnos_totales ?? 0}
                          buckets={bucketsDe(detalle)}
                          totalMicros={detalle?.paridad?.reportado_micros ?? null}
                          paridad={
                            detalle?.paridad ?? {
                              reportado_micros: null,
                              calculado_micros: null,
                              divergencia_pct: null,
                              completo: false,
                              catalogo_sin_construir: true,
                            }
                          }
                          join={{
                            corridas: detalle?.turnos_totales ?? 0,
                            // `null`, no 0: sin señal de gate, «ninguna se rechazó» sería una
                            // afirmación sobre el proceso que nadie midió (T22 sigue abierto).
                            rechazadas: null,
                            costo_rechazadas_micros: null,
                            rotaciones: null,
                          }}
                          detectores={detalle?.detectores ?? []}
                          noMedidos={noMedidos}
                          onReintentar={() => setNonce((n) => n + 1)}
                        />
                      ) : undefined
                    }
                  />
                ) : undefined
              }
            />
            {/* H-5 — el diálogo vive detrás del enlace «qué guardamos» de la franja. */}
            {politicaAbierta && viewedId && (
              <PoliticaDatosDialog
                arnes={viewedId}
                camposPersistidos={CAMPOS_PERSISTIDOS_HOOK}
                retencionDias={salud?.retencion_dias ?? 0}
                retencionPropuesta={salud?.retencion_propuesta ?? true}
                corridasPorBorrar={resumen?.corridas ?? 0}
                onBorrar={() => {
                  api
                    .telemetriaBorrarArnes(viewedId)
                    .finally(() => {
                      setPoliticaAbierta(false)
                      setNonce((n) => n + 1)
                    })
                    .catch(() => undefined)
                }}
                onCerrar={() => setPoliticaAbierta(false)}
              />
            )}
          </div>
        )}
      </div>
    </div>
  )
}
