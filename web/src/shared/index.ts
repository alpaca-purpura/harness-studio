// Public API de la capa `shared` (FSD). Los consumidores importan desde aquí o desde
// segmentos públicos (`@/shared/ui/button`), nunca de rutas internas profundas.
export { cn } from "./lib/cn";
export { Button, buttonVariants, type ButtonProps } from "./ui/button";
export { useAppStore, bindHashState, type Theme } from "./store/app-store";
export { tokens, type TokenName } from "./config/tokens";
