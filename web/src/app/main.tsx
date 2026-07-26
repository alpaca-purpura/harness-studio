import { StrictMode } from "react"
import { createRoot } from "react-dom/client"
import { bindHashState, instalarCazadorDeErrores, useAppStore } from "@/shared"
import { App } from "./App"
import "./styles/index.css"

// Arranque de la SPA. Aplica el tema inicial al <html data-theme> y cablea el hash-state.
document.documentElement.setAttribute("data-theme", useAppStore.getState().theme)
bindHashState()
// Primero de todo (RF-230): un error de arranque es el que menos chance tiene de contarse
// solo, y el WebView de la app instalada no tiene devtools donde mirarlo.
instalarCazadorDeErrores()

const root = document.getElementById("root")
if (!root) throw new Error("no #root")

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
