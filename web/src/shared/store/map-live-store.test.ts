// Unit tests del store de revisiones vivas del Mapa (RF-187, proyecto vitest `unit`).

import { describe, expect, it } from "vitest"
import { useMapLive } from "./map-live-store"

describe("map-live-store", () => {
  it("bump incrementa por arnés, sin tocar los demás", () => {
    const { bump } = useMapLive.getState()
    expect(useMapLive.getState().rev["vitalia"] ?? 0).toBe(0)
    bump("vitalia")
    bump("vitalia")
    bump("otro")
    expect(useMapLive.getState().rev["vitalia"]).toBe(2)
    expect(useMapLive.getState().rev["otro"]).toBe(1)
  })
})
