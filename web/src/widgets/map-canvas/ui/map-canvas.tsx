import { useCallback, useLayoutEffect, useMemo, useRef, useState } from "react"
import { ArnesNode, type Graph, selectEdges, selectGuardia, selectLanes } from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"
import { SUPPORT_BANDS } from "../model/bands"
import { useEdgePaths } from "../model/use-edge-paths"
import { useViewport } from "../model/use-viewport"
import { Band } from "./band"
import { BaseBand } from "./base-band"
import { EdgeLayer } from "./edge-layer"
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
}

export function MapCanvas({
  graph,
  selectedId,
  onSelect,
  harnesses,
  activeId,
  onPick,
}: MapCanvasProps) {
  const viewportRef = useRef<HTMLDivElement>(null)
  const contentRef = useRef<HTMLDivElement>(null)
  const [focus, setFocus] = useState<string | null>(null)
  const [helpOpen, setHelpOpen] = useState(false)

  const guardia = selectGuardia(graph)
  const lanes = selectLanes(graph)
  const { z, transform, grabbing, fitView, zoomIn, zoomOut } = useViewport(viewportRef, contentRef)
  const paths = useEdgePaths(contentRef, selectEdges(graph), { z, focusId: focus })

  // Related set for hover dimming (RF-33): the focused node + its direct neighbors stay lit.
  const related = useMemo<ReadonlySet<string> | undefined>(() => {
    if (!focus) return undefined
    const s = new Set<string>([focus])
    for (const e of graph.edges ?? []) {
      if (e.de === focus) s.add(e.a)
      if (e.a === focus) s.add(e.de)
    }
    return s
  }, [focus, graph.edges])
  const dimmed = (id: string) => related !== undefined && !related.has(id)

  // Overview-first: fit the whole arnés on mount and whenever the arnés changes (RF-50). arnesId
  // is the intended trigger (re-fit on arnés swap), not read inside — hence the exhaustive-deps hint.
  const arnesId = graph.arnes?.id
  // biome-ignore lint/correctness/useExhaustiveDependencies: arnesId is the re-fit trigger.
  useLayoutEffect(() => {
    fitView()
    setFocus(null)
  }, [fitView, arnesId])

  // Focus reveal (RF-32): entering/focusing a node focuses it; leaving to a non-node clears it.
  // Both mouse (hover) and keyboard (focus) drive it — the same progressive enhancement.
  const enter = useCallback(
    (target: EventTarget | null) => {
      const n = (target as HTMLElement | null)?.closest?.<HTMLElement>(".node")
      const id = n?.dataset["nodeId"]
      if (id && id !== focus) setFocus(id)
    },
    [focus],
  )
  const leave = useCallback(
    (target: EventTarget | null, related: EventTarget | null) => {
      if (!(target as HTMLElement | null)?.closest?.(".node")) return
      const to = (related as HTMLElement | null)?.closest?.<HTMLElement>(".node")
      if (to && to.dataset["nodeId"] === focus) return
      setFocus(null)
    },
    [focus],
  )

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
                    />
                  ))
                )}
              </div>
            </Region>

            <Region kind="proceso" title="Proceso · carriles por fase">
              <div className="lanes">
                {lanes.map((l) => (
                  <Lane
                    key={l.fase}
                    fase={l.fase}
                    nodes={l.nodos}
                    related={related}
                    selectedId={selectedId}
                    onSelect={onSelect}
                  />
                ))}
              </div>
            </Region>

            <Region kind="soporte" title="Base · conocimiento, reglas y soporte del arnés">
              {SUPPORT_BANDS.map((bd) => {
                const ns = graph.nodos.filter((n) => n.banda === bd.id)
                if (ns.length === 0) return null
                return bd.id === "base" ? (
                  <BaseBand
                    key={bd.id}
                    nodes={ns}
                    related={related}
                    selectedId={selectedId}
                    onSelect={onSelect}
                  />
                ) : (
                  <Band
                    key={bd.id}
                    band={bd}
                    nodes={ns}
                    related={related}
                    selectedId={selectedId}
                    onSelect={onSelect}
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
