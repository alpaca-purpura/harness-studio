import { create } from "zustand"

// Estado global mínimo (Zustand + hash-state). Decisión HS-05: app de escritorio SIN router;
// la "ruta" = fragmento de hash (`#/mapa`, `#/portafolio`), y las `pages` son composition-roots
// que reaccionan a ella. Aquí solo el esqueleto: tema + vista actual.

export type Theme = "light" | "dark"

// readView — la "ruta" es el fragmento de hash. El query-string del hash NO forma parte de la
// ruta: es estado DE la vista (p.ej. `#/portafolio?plano=marketplaces`, AG-D8 decisión 8 del
// paquete 2026-07-23), y la página que lo escribe es la que lo lee. Sin este recorte,
// `#/portafolio?plano=marketplaces` no matchearía la ruta `portafolio` y GlobalView caería a
// ComingSoon.
function readView(): string {
  if (typeof window === "undefined") return "mapa"
  const sinPrefijo = window.location.hash.replace(/^#\/?/, "")
  return (sinPrefijo.split("?")[0] ?? "") || "mapa"
}

interface AppState {
  theme: Theme
  view: string
  // mapaPeek (Slice 1, plan §2.7/S1-D13): puente Portafolio→Mapa, un solo consumo —
  // WorkspaceStage lo lee una vez (apunta viewedId al id efectivo devuelto por
  // observarEnMapa) y lo limpia. NO más lógica que esto, es solo un buzón.
  mapaPeek: string | null
  // propuestaChat (D26.4 · RF-255 · BR-M12): el buzón de «Proponerlo en el chat». Mismo patrón
  // de un solo consumo que `mapaPeek` — la tarjeta lo escribe, el Dock lo lee UNA vez y lo
  // limpia.
  //
  // 🔴 **Puebla el composer; NO envía.** Auto-enviar convertiría un clic en un turno real
  // contra el código del operador, y el fix se aplica por el camino de siempre, con sus
  // permisos y su gate. Es exactamente la razón por la que la tarjeta no tiene ninguna prop
  // de escritura.
  propuestaChat: string | null
  toggleTheme: () => void
  setView: (view: string) => void
  syncFromHash: () => void
  setMapaPeek: (id: string | null) => void
  setPropuestaChat: (texto: string | null) => void
}

export const useAppStore = create<AppState>((set, get) => ({
  // dark-first (rebrand PRENTER, decisiones.md D2): identidad real de marca = negro+teal.
  // El pipeline de tokens NO cambia de convención ($value=light sigue siendo la raíz DTCG,
  // $extensions.mode.dark el override) — lo único que cambia es este default runtime.
  theme: "dark",
  view: readView(),
  mapaPeek: null,
  propuestaChat: null,
  toggleTheme: () => {
    const next: Theme = get().theme === "light" ? "dark" : "light"
    if (typeof document !== "undefined") document.documentElement.setAttribute("data-theme", next)
    set({ theme: next })
  },
  setView: (view) => {
    if (typeof window !== "undefined") window.location.hash = `#/${view}`
    set({ view })
  },
  syncFromHash: () => set({ view: readView() }),
  setMapaPeek: (id) => set({ mapaPeek: id }),
  setPropuestaChat: (texto) => set({ propuestaChat: texto }),
}))

/** Cablea el store al hashchange del navegador. Llamar una vez en el arranque (app layer). */
export function bindHashState(): () => void {
  if (typeof window === "undefined") return () => {}
  const handler = () => useAppStore.getState().syncFromHash()
  window.addEventListener("hashchange", handler)
  return () => window.removeEventListener("hashchange", handler)
}
