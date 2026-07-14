/** true dentro del shell Tauri (global inyectado por el core, no por un plugin
 *  particular); false en un browser plano — el shell no llegó a levantar el WebView.
 *  Bugfix fix-repo-self-update (RF-110): gate del selector nativo de carpeta — un
 *  browser de dev sigue con el mensaje textual apuntando a --repo/ARNESIA_REPO. */
export function isTauri(): boolean {
  return typeof window !== "undefined" && "__TAURI_INTERNALS__" in window
}
