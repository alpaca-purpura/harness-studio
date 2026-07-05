//! Shell de escritorio ArnesIA (Tauri 2 — HS-04).
//!
//! El shell es CLIENTE del daemon Go `arnesia` (API HTTP/SSE en http://localhost:4200):
//! aporta ventana + single-instance + lanza el daemon como sidecar. NO importa el core
//! (boundary `core-no-importa-shell`): el WebView consume la misma API que el modo headless.

use std::net::TcpStream;
use std::time::Duration;

use tauri_plugin_shell::process::CommandEvent;
use tauri_plugin_shell::ShellExt;

/// Puerto donde escucha el daemon Go.
const DAEMON_ADDR: &str = "127.0.0.1:4200";

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
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            // Sidecar: si el daemon NO responde en :4200, spawnear el externalBin
            // `binaries/arnesia-<target-triple>` con `serve`. El WebView (dev :5173 /
            // prod frontendDist) habla con esa misma API.
            if daemon_running() {
                eprintln!("[arnesia] daemon ya activo en {DAEMON_ADDR}; no spawneo sidecar");
                return Ok(());
            }
            match app.shell().sidecar("arnesia-daemon") {
                Ok(cmd) => match cmd.args(["serve"]).spawn() {
                    Ok((mut rx, _child)) => {
                        // Drena stdout/stderr del daemon para que su pipe no se llene, y
                        // registra su salida. Tauri NO auto-reapea en crash duro (supervisor
                        // = fase futura).
                        tauri::async_runtime::spawn(async move {
                            while let Some(event) = rx.recv().await {
                                match event {
                                    CommandEvent::Stdout(line) | CommandEvent::Stderr(line) => {
                                        eprintln!("[arnesia] {}", String::from_utf8_lossy(&line));
                                    }
                                    CommandEvent::Terminated(payload) => {
                                        eprintln!("[arnesia] daemon terminó: {payload:?}");
                                    }
                                    _ => {}
                                }
                            }
                        });
                    }
                    Err(e) => eprintln!("[arnesia] no pude spawnear el daemon: {e}"),
                },
                Err(e) => eprintln!("[arnesia] sidecar 'arnesia-daemon' no disponible: {e}"),
            }
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error al arrancar el shell de ArnesIA");
}

/// daemon_running sondea el puerto del daemon con un timeout corto (bind-or-bail: si algo
/// ya escucha, no spawneamos una segunda instancia).
fn daemon_running() -> bool {
    DAEMON_ADDR
        .parse()
        .ok()
        .and_then(|addr| TcpStream::connect_timeout(&addr, Duration::from_millis(300)).ok())
        .is_some()
}
