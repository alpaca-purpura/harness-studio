//! Shell de escritorio ArnesIA (Tauri 2 — HS-04 · endurecido HS-06 · UI-del-daemon HS-11 #8).
//!
//! El shell es CLIENTE del daemon Go `arnesia` (API HTTP/SSE en http://localhost:4200):
//! aporta ventana + single-instance + lanza el daemon como sidecar. NO importa el core
//! (boundary `core-no-importa-shell`).
//!
//! **La UI vive en el daemon (candidata #8 firmada):** el WebView nace en la página
//! embebida `conectando.html`, que sondea `/healthz` (HS-14 fix ②: el único endpoint
//! exento de los 3 gates de auth.go) y salta a
//! `http://127.0.0.1:4200/` — la SPA que se ve SIEMPRE es la servida por el binario Go
//! (un solo cuerpo desplegable UI+API; el self-update refresca ambas). La SPA embebida
//! del bundle deja de mostrarse: queda solo como transporte de `conectando.html`.
//!
//! **HS-06 — el shell es la raíz de confianza del token de API** (boundary
//! `superficie-local-confinada`): mint un token por lanzamiento, se lo pasa al daemon por
//! env `ARNESIA_AUTH_TOKEN` al spawnearlo, y al WebView por `initialization_script`
//! (`window.__ARNESIA_TOKEN__` — corre en cada documento, incluido el origin del daemon;
//! reemplaza al comando `invoke('auth_token')`, que exigiría abrir IPC a un origin
//! remoto). Así ninguna web ajena en el navegador puede conducir el agente.

use std::net::TcpStream;
use std::sync::{Arc, Mutex};
use std::time::Duration;

use tauri::{Manager, WebviewUrl, WebviewWindowBuilder};
use tauri_plugin_shell::process::{CommandChild, CommandEvent};
use tauri_plugin_shell::ShellExt;

/// Puerto donde escucha el daemon Go.
const DAEMON_ADDR: &str = "127.0.0.1:4200";

/// Punto de arranque compartido entre desktop (`main.rs`) y mobile (`mobile_entry_point`).
#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    // Mint del token ANTES de construir: es la raíz de confianza de la superficie local.
    let token = mint_token();
    // Solo spawneamos (y por ende inyectamos token) si el daemon NO está ya arriba. En attach
    // (dev) no conocemos su token → el WebView va sin token (Host+Origin lo protegen igual).
    let spawn = !daemon_running();

    // Handle al hijo sidecar para poder matarlo al salir (evita daemon huérfano con un token
    // que un próximo shell no podría replicar).
    let child_slot: Arc<Mutex<Option<CommandChild>>> = Arc::new(Mutex::new(None));
    let child_for_setup = child_slot.clone();
    let token_for_setup = token.clone();

    tauri::Builder::default()
        // single-instance DEBE ir de primero (recomendación oficial del plugin). Mitigación Mint:
        // una sola ventana; al reabrir, la 2a instancia reenfoca la existente en vez de duplicar
        // (HS-14 fix ③ — el callback estaba vacío: una 2a instancia se tragaba en silencio sin
        // reenfocar nada).
        .plugin(tauri_plugin_single_instance::init(|app, _argv, _cwd| {
            if let Some(win) = app.get_webview_window("main") {
                let _ = win.unminimize();
                let _ = win.show();
                let _ = win.set_focus();
            }
            // TODO(fase futura): rutear el deep-link `arnesia://` recibido en `_argv`
            // (ojo bug single-instance+deep-link tauri#12726).
        }))
        .plugin(tauri_plugin_shell::init())
        // Selector nativo de carpeta (bugfix fix-repo-self-update, RF-110): gesto real
        // de OS para configurar el repo del self-update — no scriptable por contenido
        // web/XSS de la SPA (decisión #3 del paquete).
        .plugin(tauri_plugin_dialog::init())
        .setup(move |app| {
            // Ventana programática (no en tauri.conf.json): el initialization_script se fija
            // al construir y el token se mintea en runtime. El script corre en CADA documento
            // del WebView — conectando.html y la SPA del daemon lo ven; un browser normal no.
            let mut win = WebviewWindowBuilder::new(
                app,
                "main",
                WebviewUrl::App("conectando.html".into()),
            )
            .title("ArnesIA")
            .inner_size(1440.0, 900.0);
            if spawn {
                win = win.initialization_script(format!(
                    "window.__ARNESIA_TOKEN__ = \"{token_for_setup}\";"
                ));
            }
            win.build()?;

            // Sidecar: si el daemon NO responde en :4200, spawnear el externalBin
            // `binaries/arnesia-<target-triple>` con `serve` + el token por env.
            if !spawn {
                eprintln!("[arnesia] daemon ya activo en {DAEMON_ADDR}; no spawneo (attach: WebView sin token)");
                return Ok(());
            }
            match app.shell().sidecar("arnesia-daemon") {
                // El token viaja por env, no por args → no cambia el allowlist de la capability
                // (`shell:allow-execute` fija args a ["serve"]).
                Ok(cmd) => match cmd
                    .args(["serve"])
                    .env("ARNESIA_AUTH_TOKEN", &token_for_setup)
                    .spawn()
                {
                    Ok((mut rx, child)) => {
                        *child_for_setup.lock().expect("child slot") = Some(child);
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
        .build(tauri::generate_context!())
        .expect("error al arrancar el shell de ArnesIA")
        .run(move |_app, event| {
            // Matar el sidecar al salir: un daemon huérfano quedaría con un token que el próximo
            // lanzamiento no podría replicar (el WebView atacharía sin token → 401).
            if let tauri::RunEvent::Exit = event {
                if let Some(child) = child_slot.lock().expect("child slot").take() {
                    let _ = child.kill();
                }
            }
        });
}

/// mint_token genera un token aleatorio de 256 bits en hex (raíz de confianza de la API local).
fn mint_token() -> String {
    let mut b = [0u8; 32];
    getrandom::getrandom(&mut b).expect("getrandom: sin fuente de entropía");
    b.iter().map(|x| format!("{x:02x}")).collect()
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
