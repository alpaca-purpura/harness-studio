//! Shell de escritorio ArnesIA (Tauri 2 — HS-04 · endurecido HS-06).
//!
//! El shell es CLIENTE del daemon Go `arnesia` (API HTTP/SSE en http://localhost:4200):
//! aporta ventana + single-instance + lanza el daemon como sidecar. NO importa el core
//! (boundary `core-no-importa-shell`): el WebView consume la misma API que el modo headless.
//!
//! **HS-06 — el shell es la raíz de confianza del token de API** (boundary
//! `superficie-local-confinada`): mint un token por lanzamiento, se lo pasa al daemon por env
//! `ARNESIA_AUTH_TOKEN` al spawnearlo, y el WebView lo pide por `invoke('auth_token')`. Así
//! ninguna web ajena en el navegador puede conducir el agente. El token viaja shell→daemon por
//! env y shell→WebView por comando; el core sigue sin importar al shell.

use std::sync::{Arc, Mutex};
use std::net::TcpStream;
use std::time::Duration;

use tauri_plugin_shell::process::{CommandChild, CommandEvent};
use tauri_plugin_shell::ShellExt;

/// Puerto donde escucha el daemon Go.
const DAEMON_ADDR: &str = "127.0.0.1:4200";

/// Token de capacidad que el WebView usa contra la API del daemon. Se guarda en el estado de
/// Tauri; `auth_token` lo devuelve. Vacío cuando el shell se ATTACHEA a un daemon preexistente
/// (dev) cuyo token no conoce → el WebView opera solo bajo Host+Origin.
struct AuthToken(String);

/// auth_token entrega al WebView el token minteado por el shell (o "" en dev/attach).
#[tauri::command]
fn auth_token(state: tauri::State<'_, AuthToken>) -> String {
    state.0.clone()
}

/// Punto de arranque compartido entre desktop (`main.rs`) y mobile (`mobile_entry_point`).
#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    // Mint del token ANTES de construir: es la raíz de confianza de la superficie local.
    let token = mint_token();
    // Solo spawneamos (y por ende inyectamos token) si el daemon NO está ya arriba. En attach
    // (dev) no conocemos su token → el WebView va sin token (Host+Origin lo protegen igual).
    let spawn = !daemon_running();
    let webview_token = if spawn { token.clone() } else { String::new() };

    // Handle al hijo sidecar para poder matarlo al salir (evita daemon huérfano con un token
    // que un próximo shell no podría replicar).
    let child_slot: Arc<Mutex<Option<CommandChild>>> = Arc::new(Mutex::new(None));
    let child_for_setup = child_slot.clone();

    tauri::Builder::default()
        // single-instance DEBE ir de primero (recomendación oficial del plugin). Mitigación Mint:
        // una sola ventana; al reabrir, la 2a instancia reenfoca la existente en vez de duplicar.
        .plugin(tauri_plugin_single_instance::init(|_app, _argv, _cwd| {
            // TODO(fase futura): reenfocar la ventana "main" y, si aplica, rutear el deep-link
            // `arnesia://` recibido en `_argv` (ojo bug single-instance+deep-link tauri#12726).
        }))
        .plugin(tauri_plugin_shell::init())
        .manage(AuthToken(webview_token))
        .invoke_handler(tauri::generate_handler![auth_token])
        .setup(move |app| {
            // Sidecar: si el daemon NO responde en :4200, spawnear el externalBin
            // `binaries/arnesia-<target-triple>` con `serve` + el token por env. El WebView
            // (dev :5173 / prod frontendDist) habla con esa misma API.
            if !spawn {
                eprintln!("[arnesia] daemon ya activo en {DAEMON_ADDR}; no spawneo (dev: WebView sin token)");
                return Ok(());
            }
            match app.shell().sidecar("arnesia-daemon") {
                // El token viaja por env, no por args → no cambia el allowlist de la capability
                // (`shell:allow-execute` fija args a ["serve"]).
                Ok(cmd) => match cmd.args(["serve"]).env("ARNESIA_AUTH_TOKEN", &token).spawn() {
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
