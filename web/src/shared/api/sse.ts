// SSE Dock client. One EventSource multiplexes every session's conductor stream; the
// daemon tags each `dock` frame with session_id so the store can route them. The
// browser auto-reconnects and replays missed frames via Last-Event-ID.

import { api } from "./client"
import type { DockFrame, MapFrame } from "./types"

export interface DockConnection {
  close(): void
}

// connectDock opens the multiplexed stream and invokes onFrame for each `dock` event.
// onStatus reports connection liveness for the UI. onMap (RF-187) receives the `map`
// events the daemon publishes when a turn reindexed an arnés — the same multiplexed
// stream, no second EventSource. The capability token (when set) rides as
// a query param because EventSource cannot set an Authorization header (boundary
// superficie-local-confinada `sse-token-o-origin`).
export function connectDock(
  onFrame: (frame: DockFrame) => void,
  onStatus?: (connected: boolean) => void,
  onMap?: (frame: MapFrame) => void,
): DockConnection {
  const token = api.token()
  const url = token ? `${api.base}/events?token=${encodeURIComponent(token)}` : `${api.base}/events`
  const es = new EventSource(url)

  es.addEventListener("open", () => onStatus?.(true))
  es.addEventListener("error", () => onStatus?.(false))

  es.addEventListener("dock", (ev) => {
    try {
      onFrame(JSON.parse((ev as MessageEvent).data) as DockFrame)
    } catch {
      // A malformed frame is non-fatal; skip it.
    }
  })

  if (onMap) {
    es.addEventListener("map", (ev) => {
      try {
        onMap(JSON.parse((ev as MessageEvent).data) as MapFrame)
      } catch {
        // A malformed frame is non-fatal; skip it.
      }
    })
  }

  return { close: () => es.close() }
}
