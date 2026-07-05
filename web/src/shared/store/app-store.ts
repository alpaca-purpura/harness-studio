import { create } from "zustand"

// Estado global mínimo (Zustand + hash-state). Decisión HS-05: app de escritorio SIN router;
// la "ruta" = fragmento de hash (`#/mapa`, `#/portafolio`), y las `pages` son composition-roots
// que reaccionan a ella. Aquí solo el esqueleto: tema + vista actual.

export type Theme = "light" | "dark"

function readView(): string {
  if (typeof window === "undefined") return "mapa"
  return window.location.hash.replace(/^#\/?/, "") || "mapa"
}

interface AppState {
  theme: Theme
  view: string
  toggleTheme: () => void
  setView: (view: string) => void
  syncFromHash: () => void
}

export const useAppStore = create<AppState>((set, get) => ({
  theme: "light",
  view: readView(),
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
}))

/** Cablea el store al hashchange del navegador. Llamar una vez en el arranque (app layer). */
export function bindHashState(): () => void {
  if (typeof window === "undefined") return () => {}
  const handler = () => useAppStore.getState().syncFromHash()
  window.addEventListener("hashchange", handler)
  return () => window.removeEventListener("hashchange", handler)
}
