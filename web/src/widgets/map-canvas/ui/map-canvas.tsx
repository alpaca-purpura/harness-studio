import { Fragment, useCallback, useLayoutEffect, useMemo, useRef, useState } from "react"
import {
  ArnesNode,
  type ArtefactosMode,
  artEdges,
  type Graph,
  planGutter,
  selectArtefactos,
  selectEdges,
  selectGuardia,
  selectLanes,
  selectRefsEntrada,
  selectSoporte,
} from "@/entities/arnes"
import type { CifraCaja } from "@/entities/telemetria"
import { usd } from "@/entities/telemetria"
import { cn } from "@/shared/lib/cn"
import { ErrorBoundary } from "@/shared/ui/error-boundary"
import { SUPPORT_BANDS } from "../model/bands"
import type { Capa } from "../model/layers"
import { propsDeMejora } from "../model/props-de-mejora"
import { type DrawableEdge, useEdgePaths } from "../model/use-edge-paths"
import { useViewport } from "../model/use-viewport"
import { Band } from "./band"
import { BaseBand } from "./base-band"
import { EdgeLayer } from "./edge-layer"
import { HandoffGutter } from "./handoff-gutter"
import { HelpPanel } from "./help-panel"
import { Lane } from "./lane"
import { Region } from "./region"

// MapCanvas is the map surface (Hito 1, read-only) — a VERBATIM composition of the signed v3
// mockup: a fixed geography (Guardia band · one lane per fase · Base band) inside a pan/zoom
// stage, with an SVG edge overlay (spine-always + hover-reveal). Substrate is HTML+SVG, NOT
// React Flow (RF → Organigrama). It is CANVAS: it knows nothing of the rail, topbar or dock —
// the chrome feeds it a graph and (Hito 2) a harness picker; it never reaches upward
// (canvas ⊥ chrome). Only the Estructura layer exists in the MVP.

interface MapCanvasProps {
  graph: Graph
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
  // The arnés picker (RF-72 · Hito 2). In stories these are the demo fixtures (the mockup's
  // Luana/dev-full-cycle example toggle, RF-55); in production, listHarnesses.
  harnesses?: readonly { id: string; label: string }[] | undefined
  activeId?: string | undefined
  onPick?: ((id: string) => void) | undefined
  // Franja Artefactos (D2/D11, RF-143): off = mapa actual idéntico (cero DOM extra) ·
  // auto = chips solo al seleccionar · todos = siempre. Default off; el toggle vive en MapBar.
  artefactos?: ArtefactosMode | undefined
  /**
   * Capa activa (T37). **Cero rama de layout**: la geografía es la misma en las cuatro. La capa
   * solo decide si los nodos reciben sus props de mejora — con `estructura` no reciben ninguna
   * y el DOM es idéntico al de hoy (RF-245, guardián `CapaMejoraSupersetGeografia`).
   */
  capa?: Capa | undefined
  /**
   * El gasto por caja, keyeado por `nodeId`. **Este widget es el único que puede importar las
   * DOS entities** (`arnes` y `telemetria`), porque widgets→entities es la dirección legal —
   * por eso es acá donde `CifraCaja` se compone en las props PRIMITIVAS del nodo (D18).
   */
  mejora?: ReadonlyMap<string, CifraCaja> | undefined
  /**
   * Total por fase, keyeado por nombre de fase. `null` en una fase ⇒ el carril dice «sin dato»
   * (RF-244). **El mapa entero `undefined` ⇒ el carril no muestra NADA**: sin medición en la
   * ventana manda el estado 1 de la franja, y repetir «sin dato» en cada carril sería una
   * segunda afirmación sobre el mismo hecho (C-2).
   */
  totalesPorFase?: ReadonlyMap<string, number | null> | undefined
  /** El motivo de «sin dato atribuible» por nodo, resuelto por la página con la clase real. */
  motivosSinDato?: ReadonlyMap<string, string> | undefined
}

// The render-error net (nomenclatura-arnes §4.5): a malformed node (e.g. a clase outside the
// enum) must fail VISIBLY and take down ONLY the canvas — never the whole shell tree. The
// boundary wraps the inner component so anything thrown while rendering the map (selectors
// included) lands in the honest fallback panel.
export function MapCanvas(props: MapCanvasProps) {
  return (
    <ErrorBoundary label={props.graph.arnes?.id}>
      <MapCanvasInner {...props} />
    </ErrorBoundary>
  )
}

function MapCanvasInner({
  graph,
  selectedId,
  onSelect,
  harnesses,
  activeId,
  onPick,
  artefactos = "off",
  capa = "estructura",
  mejora,
  totalesPorFase,
  motivosSinDato,
}: MapCanvasProps) {
  const viewportRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const [focus, setFocus] = useState<string | null>(null)
  const [helpOpen, setHelpOpen] = useState(false)
  // Gutters expandidos por «+N más» (D11c) — índice del gutter; se pliega al cambiar arnés.
  const [expanded, setExpanded] = useState<ReadonlySet<number>>(new Set())

  const guardia = selectGuardia(graph)
  // Memoizado: lanes alimenta la cadena gutterChips→planes→drawableEdges→useEdgePaths;
  // una identidad nueva por render dispararía el efecto de medición en bucle.
  const lanes = useMemo(() => selectLanes(graph), [graph])
  const { z, transform, grabbing, fitView, zoomIn, zoomOut } = useViewport(viewportRef, contentRef)

  // Franja Artefactos (RF-140/142): chips derivados + plan por gutter (tope D11c) +
  // edges escribe/lee de los chips VISIBLES, suprimiendo el invoca del hand-off cubierto.
  const chips = useMemo(
    () => (artefactos === "off" ? [] : selectArtefactos(graph)),
    [graph, artefactos],
  )
  // Gutter i vive ANTES del carril i (nacidos tras el carril i-1 + externos que entran a i);
  // el gutter lanes.length es el de cierre (tras el último carril).
  const gutterChips = useMemo(
    () =>
      Array.from({ length: lanes.length + 1 }, (_, i) => {
        const prev = i > 0 ? chips.filter((c) => c.after === lanes[i - 1]?.fase) : []
        const ext = i < lanes.length ? chips.filter((c) => c.before === lanes[i]?.fase) : []
        return [...prev, ...ext]
      }),
    [chips, lanes],
  )
  const planes = useMemo(
    () =>
      gutterChips.map((cs, i) =>
        planGutter(cs, { mode: artefactos, selectedId, expanded: expanded.has(i) }),
      ),
    [gutterChips, artefactos, selectedId, expanded],
  )
  const drawableEdges = useMemo<DrawableEdge[]>(() => {
    const visibles = new Set(planes.flatMap((p) => p.visibles.map((c) => c.id)))
    const { edges: artes, suprimidos } = artEdges(chips, visibles)
    const base = selectEdges(graph).filter(
      (e) => !(e.tipo === "invoca" && suprimidos.has(`${e.de}>${e.a}`)),
    )
    return [...base, ...artes]
  }, [graph, chips, planes])
  const paths = useEdgePaths(contentRef, drawableEdges, { z, focusId: focus })

  // La composición `CifraCaja → props primitivas`, una sola vez por render (D18). Con la capa
  // apagada el mapa queda VACÍO y el nodo no recibe ninguna prop nueva: el DOM es el de hoy.
  const mejoraDeCarril = useMemo(() => {
    if (capa !== "mejora") return undefined
    const m = new Map<string, ReturnType<typeof propsDeMejora>>()
    for (const nodo of graph.nodos) {
      const cifra = mejora?.get(nodo.id)
      const motivo = motivosSinDato?.get(nodo.id)
      if (cifra) m.set(nodo.id, propsDeMejora(cifra, motivo))
      else if (motivo !== undefined) m.set(nodo.id, { motivoSinDato: motivo })
    }
    return m
  }, [capa, graph, mejora, motivosSinDato])

  // Related set for hover dimming (RF-33): the focused node + its direct neighbors stay
  // lit — including the artefacto chips (the derived edges participate, mockup:715-721).
  const related = useMemo<ReadonlySet<string> | undefined>(() => {
    if (!focus) return undefined
    const s = new Set<string>([focus])
    for (const e of drawableEdges) {
      if (e.de === focus) s.add(e.a)
      if (e.a === focus) s.add(e.de)
    }
    return s
  }, [focus, drawableEdges])
  const dimmed = (id: string) => related !== undefined && !related.has(id)

  // Overview-first: fit the whole arnés on mount and whenever the arnés changes (RF-50). arnesId
  // is the intended trigger (re-fit on arnés swap), not read inside — hence the exhaustive-deps hint.
  const arnesId = graph.arnes?.id
  // biome-ignore lint/correctness/useExhaustiveDependencies: arnesId is the re-fit trigger.
  useLayoutEffect(() => {
    fitView()
    setFocus(null)
    setExpanded(new Set())
  }, [fitView, arnesId])

  // Focus reveal (RF-32): entering/focusing a node OR an artefacto chip focuses it;
  // leaving to a non-node clears it. Mouse (hover) and keyboard (focus) drive it alike.
  const enter = useCallback(
    (target: EventTarget | null) => {
      const n = (target as HTMLElement | null)?.closest?.<HTMLElement>(".node,.artchip")
      const id = n?.dataset["nodeId"]
      if (id && id !== focus) setFocus(id)
    },
    [focus],
  )
  const leave = useCallback(
    (target: EventTarget | null, related: EventTarget | null) => {
      if (!(target as HTMLElement | null)?.closest?.(".node,.artchip")) return
      const to = (related as HTMLElement | null)?.closest?.<HTMLElement>(".node,.artchip")
      if (to && to.dataset["nodeId"] === focus) return
      setFocus(null)
    },
    [focus],
  )
  const toggleExpand = useCallback((i: number) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(i)) next.delete(i)
      else next.add(i)
      return next
    })
  }, [])

  return (
    <div className="arnesia-map">
      <div ref={viewportRef} className={cn("viewport", grabbing && "grabbing")}>
        <div className="stage" style={{ transform }}>
          {/* biome-ignore lint/a11y/noStaticElementInteractions: hover/focus reveal is a
              progressive enhancement over a read-only canvas; the same info is reachable via
              the always-on spine and keyboard focus. */}
          <div
            ref={contentRef}
            className="content"
            onMouseOver={(e) => enter(e.target)}
            onMouseOut={(e) => leave(e.target, e.relatedTarget)}
            onFocus={(e) => enter(e.target)}
            onBlur={(e) => leave(e.target, e.relatedTarget)}
          >
            <Region kind="guardia" title="Guardia · hooks transversales">
              <div className="row guardia-row">
                {guardia.length === 0 ? (
                  <span className="empty">— sin hooks —</span>
                ) : (
                  guardia.map((b) => (
                    <ArnesNode
                      key={b.id}
                      box={b}
                      compact
                      dim={dimmed(b.id)}
                      selected={b.id === selectedId}
                      onSelect={onSelect}
                      {...mejoraDeCarril?.get(b.id)}
                    />
                  ))
                )}
              </div>
            </Region>

            <Region kind="proceso" title="Proceso · carriles por fase">
              <div className="lanes">
                {lanes.map((l, i) => {
                  // Panel de entrada (D11b): refs ↖ SOLO de la caja seleccionada de este
                  // carril, en su gutter de entrada.
                  const refs =
                    artefactos !== "off" && selectedId && l.nodos.some((n) => n.id === selectedId)
                      ? selectRefsEntrada(
                          graph,
                          selectedId,
                          i > 0 ? (lanes[i - 1]?.fase ?? null) : null,
                        )
                      : []
                  const plan = planes[i]
                  return (
                    <Fragment key={l.fase}>
                      {artefactos !== "off" && plan && (
                        <HandoffGutter
                          plan={plan}
                          refs={refs}
                          expanded={expanded.has(i)}
                          onToggleExpand={() => toggleExpand(i)}
                          onSelect={onSelect}
                          dimmed={dimmed}
                        />
                      )}
                      <Lane
                        fase={l.fase}
                        nodes={l.nodos}
                        related={related}
                        selectedId={selectedId}
                        onSelect={onSelect}
                        // El total del carril y las marcas del nodo son DOS decisiones, y
                        // gatearlas juntas fue un error propio: sin `totalesPorFase` los nodos
                        // se quedaban sin cifra. El carril calla cuando no hay medición en la
                        // ventana (manda el estado 1 de la franja, C-2); los nodos dependen de
                        // su propio mapa.
                        mejora={mejoraDeCarril}
                        {...(capa === "mejora" && totalesPorFase !== undefined
                          ? {
                              totalUsd:
                                totalesPorFase.get(l.fase) == null
                                  ? null
                                  : usd(totalesPorFase.get(l.fase) as number),
                            }
                          : {})}
                      />
                    </Fragment>
                  )
                })}
                {artefactos !== "off" && lanes.length > 0 && planes[lanes.length] && (
                  <HandoffGutter
                    plan={planes[lanes.length] as NonNullable<(typeof planes)[number]>}
                    refs={[]}
                    expanded={expanded.has(lanes.length)}
                    onToggleExpand={() => toggleExpand(lanes.length)}
                    onSelect={onSelect}
                    dimmed={dimmed}
                  />
                )}
              </div>
            </Region>

            <Region kind="soporte" title="Base · conocimiento, reglas y soporte del arnés">
              {SUPPORT_BANDS.map((bd) => {
                // selectSoporte (not a raw banda filter): the `base` band also receives the
                // banda-desconocida nodes — visible, never silently dropped (§4.5).
                const ns = selectSoporte(graph, bd.id)
                if (ns.length === 0) return null
                return bd.id === "base" ? (
                  <BaseBand
                    key={bd.id}
                    nodes={ns}
                    related={related}
                    selectedId={selectedId}
                    onSelect={onSelect}
                    mejora={mejoraDeCarril}
                  />
                ) : (
                  <Band
                    key={bd.id}
                    band={bd}
                    nodes={ns}
                    related={related}
                    selectedId={selectedId}
                    onSelect={onSelect}
                    mejora={mejoraDeCarril}
                  />
                )
              })}
            </Region>

            <EdgeLayer paths={paths} />
          </div>
        </div>
      </div>

      {/* floating controls (mockup:163-173) */}
      <div className="ctl zoom">
        <button type="button" title="Acercar" onClick={zoomIn}>
          +
        </button>
        <button type="button" title="Alejar" onClick={zoomOut}>
          −
        </button>
        <button type="button" title="Ajustar" style={{ fontSize: 13 }} onClick={fitView}>
          ⤢
        </button>
      </div>

      {harnesses && harnesses.length > 0 && (
        <div className="ctl example">
          {harnesses.map((h) => (
            <button
              key={h.id}
              type="button"
              aria-pressed={h.id === activeId}
              onClick={() => onPick?.(h.id)}
            >
              {h.label}
            </button>
          ))}
        </div>
      )}

      <button
        type="button"
        className="help-fab"
        title="Leyenda y notas"
        aria-label="Leyenda y notas"
        onClick={() => setHelpOpen((o) => !o)}
      >
        ?
      </button>
      <HelpPanel hidden={!helpOpen} />
    </div>
  )
}
