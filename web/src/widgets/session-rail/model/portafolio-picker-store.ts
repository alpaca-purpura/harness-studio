// Store del selector de arnés de «＋ Nueva sesión» (TS-D17). SessionRail es un widget que hoy
// solo consume stores (useSessions/useAppStore) y nunca `shared/api` directo; el picker necesita
// leer el Portafolio real (GET /api/portafolio) SIN romper dos cosas: (1) entities/portafolio/model
// no puede importar shared/api (boundary domain-not-transport, ERROR), y (2) exponer transporte
// desde shell-page.tsx ampliaría el alcance firmado del paquete. Se resuelve con este store propio
// del widget que envuelve `api.listPortafolio` — mismo patrón "el store consume el transporte" que
// ya usa sessions-store, no la página.

import { create } from "zustand"
import type { EntradaCorrupta, EntradaPortafolio, PortafolioListado } from "@/entities/portafolio"
import { api } from "@/shared"

// PickerEstado — los 3 estados honestos que ya usa PortafolioList (cargando/error/datos): se
// calca el vocabulario que la Lista del Portafolio ya usa, no se inventa uno nuevo (TS-D13).
export type PickerEstado = "cargando" | "error" | "datos"

interface PortafolioPickerState {
  estado: PickerEstado
  error?: string | undefined
  entradas: EntradaPortafolio[]
  corruptas: EntradaCorrupta[]
  // cargar — dispara un GET /api/portafolio nuevo (TS-D14/RF-15: refetch en CADA apertura del
  // picker, jamás una lista cacheada de una apertura anterior — el operador pudo agregar un arnés
  // desde la vista Portafolio en la misma corrida de la app).
  cargar: () => Promise<void>
  // reset — vuelve al estado inicial al cerrar/crear (RF-17): la próxima apertura arranca limpia.
  reset: () => void
}

const INICIAL = {
  estado: "cargando" as PickerEstado,
  error: undefined,
  entradas: [] as EntradaPortafolio[],
  corruptas: [] as EntradaCorrupta[],
}

export const usePortafolioPicker = create<PortafolioPickerState>((set) => ({
  ...INICIAL,

  cargar: async () => {
    set({ estado: "cargando", error: undefined })
    try {
      const data = await api.listPortafolio<PortafolioListado>()
      set({ entradas: data.entradas ?? [], corruptas: data.corruptas ?? [], estado: "datos" })
    } catch (e) {
      set({ error: e instanceof Error ? e.message : String(e), estado: "error" })
    }
  },

  reset: () => set({ ...INICIAL }),
}))
