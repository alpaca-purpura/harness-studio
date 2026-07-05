//! Shell de escritorio ArnesIA (Tauri 2 — HS-04).
//!
//! El shell es CLIENTE del daemon Go `arnesia` (API HTTP/SSE en http://localhost:4200):
//! aporta ventana + single-instance + (fase futura) lanzar el sidecar. NO importa el core
//! (boundary `core-no-importa-shell`): el WebView consume la misma API que el modo headless.

/// Punto de arranque compartido entre desktop (`main.rs`) y mobile (`mobile_entry_point`).
#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        // single-instance DEBE ir de primero (recomendación oficial del plugin). Mitigación Mint:
        // una sola ventana; al reabrir, la 2a instancia reenfoca la existente en vez de duplicar.
        .plugin(tauri_plugin_single_instance::init(|_app, _argv, _cwd| {
            // TODO(fase futura): reenfocar la ventana "main" y, si aplica, rutear el deep-link
            // `arnesia://` recibido en `_argv` (ojo bug single-instance+deep-link tauri#12726).
        }))
        .setup(|_app| {
            // FASE FUTURA (sidecar): aquí se cablea el daemon Go `arnesia serve`.
            //   1. Sondar http://localhost:4200 (bind-or-bail / flock del daemon).
            //   2. Si NO responde -> spawnear el externalBin `binaries/arnesia-<target-triple>`
            //      vía tauri-plugin-shell (`app.shell().sidecar("arnesia")`), sub-comando `serve`.
            //      Requiere sumar `tauri-plugin-shell` a Cargo.toml + capability `shell:allow-execute`.
            //   3. Tauri no auto-reapea el sidecar en crash duro -> heartbeat/supervisor propio.
            // El WebView ya apunta a la SPA (dev: :5173 / prod: frontendDist), que habla con :4200.
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error al arrancar el shell de ArnesIA");
}
