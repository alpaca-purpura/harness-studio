import { create } from "zustand"
import { api, type Session, type Turn } from "@/shared"

// Historial de conversaciones de un arnés (RF-203, B2/MC-D7): vivas + cerradas (metadata)
// y los turnos de una cerrada reconstruidos desde la JSONL nativa. Mismo patrón que
// portafolio-picker-store: el widget consume; el transporte vive acá.
interface ConversacionesState {
  arnes: string | null
  vivas: Session[]
  cerradas: Session[]
  error?: string | undefined
  historialDe: string | null
  turnos: Turn[]
  faltantes: string[]
  cargar: (arnes: string) => Promise<void>
  abrirHistorial: (id: string) => Promise<void>
  reset: () => void
}

export const useConversaciones = create<ConversacionesState>((set) => ({
  arnes: null,
  vivas: [],
  cerradas: [],
  error: undefined,
  historialDe: null,
  turnos: [],
  faltantes: [],

  cargar: async (arnes) => {
    set({ arnes, vivas: [], cerradas: [], error: undefined, historialDe: null, turnos: [] })
    try {
      const r = await api.conversacionesDeArnes(arnes)
      set({
        vivas: r.sesiones ?? [],
        cerradas: r.cerradas ?? [],
        // El backend reporta el hueco de cerradas aparte (degradación honesta) — se muestra.
        error: r.cerradas_error,
      })
    } catch (e) {
      set({ error: e instanceof Error ? e.message : String(e) })
    }
  },

  abrirHistorial: async (id) => {
    set({ historialDe: id, turnos: [], faltantes: [] })
    try {
      const r = await api.historialCerrada(id)
      set({ turnos: r.turnos ?? [], faltantes: r.faltantes ?? [] })
    } catch (e) {
      set({ error: e instanceof Error ? e.message : String(e) })
    }
  },

  reset: () =>
    set({
      arnes: null,
      vivas: [],
      cerradas: [],
      error: undefined,
      historialDe: null,
      turnos: [],
      faltantes: [],
    }),
}))
