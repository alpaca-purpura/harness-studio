import { useCallback, useEffect, useMemo, useState } from "react"
import type { ArtefactosMode, ConformanceResult, Graph } from "@/entities/arnes"
import {
  api,
  ComingSoon,
  type HarnessSummary,
  SALUD_LABEL,
  STATUS_LABEL,
  selectActive,
  useAppStore,
  useSessions,
  VIEWS,
} from "@/shared"
import { HealthDot, Pip } from "@/shared/ui/indicators"
import { type Capa, Inspector, MapBar, MapCanvas } from "@/widgets/map-canvas"

const viewGlyph = (v: string) => VIEWS.find((x) => x[0] === v)?.[1] ?? "◵"

// WorkspaceStage is the near-fullscreen canvas of the active session (a page = composition-root).
// The session header is real; the «Mapa» view now renders the REAL Map surface — it fetches the
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
  // Only Estructura is enabled in the MVP; the switcher lives in MapBar (chrome), staged layers
  // render disabled with the "Necesita telemetría" tooltip (RF-60 / spec §8).
  const [capa, setCapa] = useState<Capa>("estructura")
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

  // Portfolio for the picker (RF-72). Empty until the daemon lists harnesses (Hito 2 backend).
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

  const pickerItems = useMemo(
    () => harnesses.map((h) => ({ id: h.id, label: h.rol ?? h.id })),
    [harnesses],
  )

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
      <header className="flex flex-wrap items-center gap-2.5 border-b border-border bg-card px-4 py-3">
        <span className="font-mono text-base font-bold">{s.arnes}</span>
        <span
          className="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-semibold"
          style={{
            background: s.status === "streaming" ? "var(--c-skill)" : "var(--secondary)",
            color:
              s.status === "streaming" ? "var(--primary-foreground)" : "var(--muted-foreground)",
          }}
        >
          <Pip status={s.status} />
          {STATUS_LABEL[s.status]}
        </span>
        {/* TS-D18: empresa/puesto salen de acá — mismo motivo de TS-D1 (fixture hardcodeada,
            N:M sin "la" empresa) y ya duplicado por los chips reales de MapBar (empresa/reporta a) */}
        <span className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
          <HealthDot salud={s.salud} /> {s.salud ? SALUD_LABEL[s.salud] : ""}
        </span>
      </header>

      <div className="min-h-0 flex-1">
        {!isMapa ? (
          <ComingSoon
            glyph={viewGlyph(s.view)}
            title={`Vista ${s.view}`}
            note={`El interior de «${s.view}» llega después. El shell, la multisesión y la conversación con Claude Code ya están vivos — abre el dock (⌘K) y pídele algo a este arnés.`}
          />
        ) : (
          // MapBar (with the arnés picker, RF-72) stays ABOVE the canvas at all times — including
          // the load-error state — so the picker is always the escape hatch to a different arnés
          // when the current one is missing from the index.
          <div className="flex h-full flex-col">
            <MapBar
              arnes={graph?.arnes}
              capa={capa}
              onCapa={setCapa}
              harnesses={pickerItems}
              activeId={viewedId}
              onPick={setViewedId}
              artefactos={artefactos}
              onArtefactos={setArtefactos}
            />
            <div className="relative min-h-0 flex-1">
              {loadErr ? (
                <ComingSoon
                  glyph="⚠"
                  title="No se pudo cargar el arnés"
                  note={`El daemon no devolvió el grafo de «${viewedId}». Elige otro arnés en la barra de arriba. ${loadErr}`}
                />
              ) : !graph ? (
                <ComingSoon
                  glyph="⬡"
                  title={`Cargando ${viewedId}…`}
                  note="Leyendo el grafo del arnés."
                />
              ) : (
                <>
                  <MapCanvas
                    graph={graph}
                    selectedId={selectedId}
                    onSelect={setSelectedId}
                    artefactos={artefactos}
                  />
                  {/* El drawer pinta algo o NO existe (decisión del operador 2026-07-07,
                      supersede el estado vacío RF-84): sin selección no se monta; ✕ la
                      limpia y el drawer desaparece entero. */}
                  {selectedBox && (
                    <Inspector
                      box={selectedBox}
                      onClose={() => setSelectedId(undefined)}
                      graph={graph}
                      onSelect={setSelectedId}
                      conformance={conformance}
                      loadFuente={loadFuente}
                    />
                  )}
                </>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
