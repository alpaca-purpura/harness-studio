import { create } from "zustand"
import { ApiError, api, type Conversacion, useSessions } from "@/shared"

// El transporte del panel de conversaciones (RF-334 · CV-D2 · design.md §3.6).
//
// Vive en `widgets/chat-dock/model` y no en `shared/store` por el mismo motivo que
// `portafolio-picker-store`: es transporte de UN widget. Los componentes de `ui/` son
// props-puras; el único que lee stores es `ChatDock`.
//
// Reemplaza a `widgets/session-rail/model/conversaciones-store.ts`, que se ELIMINÓ: sus dos
// endpoints ya no existen. La mudanza es física porque `no-sibling-widget-imports` no deja
// que un widget importe del otro.
//
// Tres disciplinas HEREDADAS, no inventadas:
//  1. refetch en CADA apertura (`portafolio-picker-store.ts:22-25`) — entre dos aperturas el
//     operador pudo mandar un turno o rotar, y una lista cacheada mentiría.
//  2. `cerrar()` descarta búsqueda y estado (`session-rail.tsx:43-47`), cero efectos.
//  3. el error del backend se guarda TAL CUAL, sin reescribirlo
//     (`portafolio-picker-store.ts:45-47`).

export type PanelEstado = "cargando" | "error" | "datos"

// El seam con `useSessions` es una suscripción DESCENDENTE a `convRev` (arquitectura.md §6.2):
// `shared` no puede importar `widgets` (`shared-no-upward`), así que es el panel el que se
// suscribe a la revisión que el store compartido bumpea con cada frame `conversacion`. Un
// contador, no la lista: quién es dueño de la lista no cambia.
let desuscribir: (() => void) | null = null
let enVuelo: AbortController | null = null

interface ConversacionesState {
  sesionId: string | null
  estado: PanelEstado
  error?: string | undefined
  conversaciones: Conversacion[]
  total: number
  busqueda: string
  abierta: boolean
  focoInicial: "filas" | "buscador"
  /** true mientras `--resume` rehidrata: pinta la franja efímera (RF-310). */
  retomando: string | null
  /**
   * el motivo del fallo de una MUTACIÓN (crear · retomar · renombrar), para la franja en
   * variante `bad` (C-11, RF-348).
   *
   * `design.md` §3.6 lo llama `retomaFallo`, con la retoma como único caso. Se generaliza
   * porque los tres fallan por lo mismo (el daemon no está, o hay un turno en vuelo) y
   * pintan la misma franja: dos campos para una superficie serían dos campos que se separan.
   * El nombre dice lo que es.
   */
  fallo?: string | undefined

  abrir: (sesionId: string, foco: "filas" | "buscador") => void
  cerrar: () => void
  buscar: (q: string) => Promise<void>
  crear: () => Promise<void>
  activar: (cid: string) => Promise<void>
  renombrar: (cid: string, titulo: string) => Promise<void>
  reintentar: () => Promise<void>
}

// motivoDe traduce un fallo a algo que el operador pueda leer SIN perder el original. El 409
// es el único que se nombra aparte, porque no es una caída: es un turno en vuelo, y la
// respuesta del operador («esperá y probá de nuevo») es distinta (E-39).
function motivoDe(e: unknown): string {
  if (e instanceof ApiError && e.status === 409) {
    return `hay un turno en vuelo — esperá a que termine (${e.message})`
  }
  return e instanceof Error ? e.message : String(e)
}

const INICIAL = {
  sesionId: null,
  estado: "cargando" as PanelEstado,
  error: undefined,
  conversaciones: [] as Conversacion[],
  total: 0,
  busqueda: "",
  abierta: false,
  focoInicial: "filas" as const,
  retomando: null,
  fallo: undefined,
}

export const useConversaciones = create<ConversacionesState>((set, get) => {
  // cargar es el único que toca la red para leer. Aborta la petición anterior: teclear rápido
  // genera búsquedas que se pisan, y la última en llegar no es necesariamente la última que
  // se pidió.
  const cargar = async (sesionId: string, q: string) => {
    enVuelo?.abort()
    const ctl = new AbortController()
    enVuelo = ctl
    set({ estado: "cargando", error: undefined })
    try {
      const r = await api.conversaciones(sesionId, q || undefined, ctl.signal)
      if (ctl.signal.aborted) return
      set({ estado: "datos", conversaciones: r.conversaciones ?? [], total: r.total ?? 0 })
    } catch (e) {
      if (ctl.signal.aborted) return
      // Jamás «0 conversaciones» cuando lo que pasó es que no se pudo preguntar (BR-CV-10):
      // el estado va a `error` con el motivo, y la lista queda vacía porque no hay lista.
      set({ estado: "error", error: motivoDe(e), conversaciones: [], total: 0 })
    }
  }

  const suscribir = (sesionId: string) => {
    desuscribir?.()
    desuscribir = useSessions.subscribe((st, prev) => {
      if (st.convRev[sesionId] === prev.convRev[sesionId]) return
      if (!get().abierta || get().sesionId !== sesionId) return
      void cargar(sesionId, get().busqueda)
    })
  }

  // mutar corre una de las tres operaciones y deja el estado ANTERIOR intacto si falla
  // (E-36/E-37): nada optimista, porque la transición la decide el dominio y un rollback de
  // lista sería adivinar qué había.
  const mutar = async (accion: (sesionId: string) => Promise<void>) => {
    const sesionId = get().sesionId ?? useSessions.getState().activeId
    if (!sesionId) return
    set({ fallo: undefined })
    try {
      await accion(sesionId)
      if (get().abierta) await cargar(sesionId, get().busqueda)
    } catch (e) {
      set({ fallo: motivoDe(e) })
    }
  }

  return {
    ...INICIAL,

    abrir: (sesionId, foco) => {
      // Refetch SIEMPRE, jamás la lista de la apertura anterior (disciplina 1).
      set({ ...INICIAL, sesionId, abierta: true, focoInicial: foco })
      suscribir(sesionId)
      void cargar(sesionId, "")
    },

    cerrar: () => {
      desuscribir?.()
      desuscribir = null
      enVuelo?.abort()
      enVuelo = null
      set({ ...INICIAL })
    },

    buscar: async (q) => {
      const sesionId = get().sesionId
      set({ busqueda: q })
      if (!sesionId) return
      await cargar(sesionId, q)
    },

    reintentar: async () => {
      const sesionId = get().sesionId
      if (!sesionId) return
      await cargar(sesionId, get().busqueda)
    },

    crear: () =>
      mutar(async (sesionId) => {
        await api.crearConversacion(sesionId)
      }),

    activar: (cid) =>
      mutar(async (sesionId) => {
        set({ retomando: cid })
        try {
          await api.activarConversacion(sesionId, cid)
          set({ abierta: false, retomando: null })
        } catch (e) {
          set({ retomando: null })
          throw e
        }
      }),

    // Vacío NO viaja: el título anterior queda intacto y el daemon ni se entera (E-30). Es la
    // misma regla que ya aplica el rail al renombrar el frente.
    renombrar: (cid, titulo) => {
      const t = titulo.trim()
      if (!t) return Promise.resolve()
      return mutar(async (sesionId) => {
        await api.renombrarConversacion(sesionId, cid, t)
      })
    },
  }
})
