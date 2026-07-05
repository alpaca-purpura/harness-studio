import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { bindHashState, useAppStore } from "@/shared";
import { App } from "./App";
import "./styles/index.css";

// Arranque de la SPA. Aplica el tema inicial al <html data-theme> y cablea el hash-state.
document.documentElement.dataset.theme = useAppStore.getState().theme;
bindHashState();

const root = document.getElementById("root");
if (!root) throw new Error("no #root");

createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
