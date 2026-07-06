import { useRef, useState } from "react"
import { type Graph, selectBase, selectEdges, selectGuardia, selectLanes } from "@/entities/arnes"
import type { Capa } from "../model/layers"
import { useEdgePaths } from "../model/use-edge-paths"
import { Band } from "./band"
import { EdgeLayer } from "./edge-layer"
import { Lane } from "./lane"
import { MapBar } from "./map-bar"

// MapCanvas is the map surface (Hito 1, read-only): a fixed geography — Guardia band (top) · one
// lane per declared fase (middle, scroll-x) · Base band (bottom) — with an SVG edge overlay. The
// substrate is HTML+SVG (fiel al mockup v3 firmado), NOT React Flow: the geography is fixed
// swim-lanes, so the layout is CSS and the connectors are measured over it (see use-edge-paths).
// React Flow is reserved for the free-canvas Organigrama.
//
// Selection is controlled (the page owns it → feeds the inspector, Hito 2). The layer is internal
// state (only Estructura is enabled in the MVP).

interface MapCanvasProps {
  graph: Graph
  selectedId?: string | undefined
  onSelect?: ((id: string) => void) | undefined
}

export function MapCanvas({ graph, selectedId, onSelect }: MapCanvasProps) {
  const contentRef = useRef<HTMLDivElement>(null)
  const [capa, setCapa] = useState<Capa>("estructura")

  const guardia = selectGuardia(graph)
  const lanes = selectLanes(graph)
  const base = selectBase(graph)
  const paths = useEdgePaths(contentRef, selectEdges(graph))

  return (
    <div className="flex h-full flex-col bg-background">
      <MapBar arnes={graph.arnes} capa={capa} onCapa={setCapa} />
      <div className="flex-1 overflow-auto">
        <div ref={contentRef} className="relative flex min-w-[900px] flex-col gap-3 p-4">
          <EdgeLayer paths={paths} />
          <Band
            label="Guardia"
            sublabel="hooks transversales — actúan en todas las fases"
            nodes={guardia}
            selectedId={selectedId}
            onSelect={onSelect}
          />
          <div className="flex items-start gap-3">
            {lanes.map((l) => (
              <Lane
                key={l.fase}
                fase={l.fase}
                nodes={l.nodos}
                selectedId={selectedId}
                onSelect={onSelect}
              />
            ))}
          </div>
          <Band
            label="Base"
            sublabel="siempre en contexto — reglas · knowledge · mcp"
            nodes={base}
            selectedId={selectedId}
            onSelect={onSelect}
          />
        </div>
      </div>
    </div>
  )
}
