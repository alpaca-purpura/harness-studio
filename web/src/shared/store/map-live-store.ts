import { create } from "zustand"

// Revisiones vivas del Mapa (RF-187): el daemon publica `event: map` cuando un turno del
// chat reindexó un arnés; acá se bumpea la revisión de ese arnés y la vista Mapa que lo
// esté mirando refetchea su grafo — sin recargar la página, sin resetear selección.
interface MapLiveState {
  rev: Record<string, number>
  bump: (harnessId: string) => void
}

export const useMapLive = create<MapLiveState>((set) => ({
  rev: {},
  bump: (harnessId) =>
    set((st) => ({ rev: { ...st.rev, [harnessId]: (st.rev[harnessId] ?? 0) + 1 } })),
}))
