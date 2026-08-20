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
    // Solo spawneamos (y por ende inyectamos token) si NADIE ocupa el puerto. En attach (dev)
    // no conocemos su token → el WebView va sin token (Host+Origin lo protegen igual). Y si el
    // ocupante NO sirve la UI, ni attach ni spawn: la ventana lo explica (CW-D8).
    let estado = estado_daemon();
    let spawn = estado == EstadoDaemon::Ausente;

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
            // Alcance decidido HS-14 deuda (2026-07-23): `arnesia://` solo enfoca la app —
            // no hay router en el FE (App.tsx = 1 store Zustand, sin rutas URL) para rutear
            // un destino de página; eso queda a un paquete propio si se pide (BACKLOG).
            if let Some(win) = app.get_webview_window("main") {
                let _ = win.unminimize();
                let _ = win.show();
                let _ = win.set_focus();
            }
        }))
        .plugin(tauri_plugin_deep_link::init())
        .plugin(tauri_plugin_shell::init())
        // Selector nativo de carpeta (bugfix fix-repo-self-update, RF-110): gesto real
        // de OS para configurar el repo del self-update — no scriptable por contenido
        // web/XSS de la SPA (decisión #3 del paquete).
        .plugin(tauri_plugin_dialog::init())
        .setup(move |app| {
            // Registro runtime del scheme (Linux dev únicamente — bundled deb/rpm/AppImage lo
            // registran solos vía el .desktop generado desde `plugins.deep-link.desktop.schemes`
            // en tauri.conf.json; en producción `register_all` es no-op si ya está registrado).
            #[cfg(target_os = "linux")]
            {
                use tauri_plugin_deep_link::DeepLinkExt;
                if let Err(e) = app.deep_link().register_all() {
                    eprintln!("[arnesia] no pude registrar el scheme arnesia://: {e}");
                }
            }

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
            // Conflicto de puerto: la página de arranque no debe saltar a la raíz del ocupante
            // (mostraría SU respuesta, típicamente un 404 en texto plano). Se lo decimos y ella
            // muestra la tarjeta accionable en vez de sondear.
            if estado == EstadoDaemon::SinUI {
                win = win.initialization_script("window.__ARNESIA_CONFLICTO__ = true;");
            }
            let ventana = win.build()?;

            // RF-215 (V-D6 FIRMADA): sin esto, el botón de dictado se cuelga MUDO en la app
            // instalada. Va acá, pegado al build de la ventana, porque el WebView nativo recién
            // existe después.
            conceder_permiso_de_microfono(&ventana);

            // Sidecar: si el daemon NO responde en :4200, spawnear el externalBin
            // `binaries/arnesia-<target-triple>` con `serve` + el token por env.
            match estado {
                EstadoDaemon::ConUI => {
                    eprintln!("[arnesia] daemon ya activo en {DAEMON_ADDR}; no spawneo (attach: WebView sin token)");
                    return Ok(());
                }
                EstadoDaemon::SinUI => {
                    eprintln!("[arnesia] {DAEMON_ADDR} está ocupado por algo que NO sirve la UI: ni attach ni spawn (el bind fallaría). Cerrá ese proceso y reabrí ArnesIA");
                    return Ok(());
                }
                EstadoDaemon::Ausente => {}
            }
            // override local (bugfix self-update-sidecar-ignora-path): tauri_plugin_shell
            // resuelve sidecar() SIEMPRE como dirname(current_exe())/programa — jamás vía
            // $PATH (ver relative_command_path en tauri-plugin-shell/src/process/mod.rs).
            // Un install .deb/.rpm pone arnesia-app Y el sidecar en /usr/bin (root): el hint
            // de la tarjeta Ajustes ("migra a ~/.local/bin/arnesia, el botón funciona sin
            // sudo") era FALSO para el flujo GUI — copiar el binario ahí no cambiaba cuál
            // sidecar se lanzaba. Este check hace ese hint real: si el operador ya migró,
            // usamos ESE binario (espacio de usuario, permite self-update); si no, el
            // sidecar empaquetado de siempre.
            let override_local = app
                .path()
                .home_dir()
                .ok()
                .map(|home| home.join(".local").join("bin").join("arnesia"))
                .filter(|p| p.is_file());

            let cmd = if let Some(bin) = &override_local {
                eprintln!("[arnesia] override local en {}: spawneo ESE binario, no el sidecar empaquetado", bin.display());
                Some(app.shell().command(bin))
            } else {
                eprintln!("[arnesia] sin override en ~/.local/bin/arnesia: spawneo el sidecar empaquetado");
                match app.shell().sidecar("arnesia-daemon") {
                    // El token viaja por env, no por args → no cambia el allowlist de la
                    // capability (`shell:allow-execute` fija args a ["serve"]).
                    Ok(cmd) => Some(cmd),
                    Err(e) => {
                        eprintln!("[arnesia] sidecar 'arnesia-daemon' no disponible: {e}");
                        None
                    }
                }
            };

            if let Some(cmd) = cmd {
                match cmd
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
                }
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

/// conceder_permiso_de_microfono engancha `permission-request` del WebView nativo y concede
/// **solo** captura de audio (RF-215, capability `arnesia.tauri.permiso-de-microfono`).
///
/// **Por qué existe:** `wry` implementa `permission-request` en macOS y Android, **no en el
/// backend `webkitgtk`**. El síntoma no es un rechazo — `getUserMedia` **queda pendiente para
/// siempre** (probado en vivo contra `libwebkit2gtk-4.1 2.52.3`, ver
/// `docs/product/stories/2026-07-25-spike-voz-dictado/spike-spec.md` §1.6 b/d). Sin este puente el
/// dictado anda en `pnpm dev` —donde el navegador concede por su cuenta— y muere mudo instalado.
///
/// Si una versión futura de `wry` maneja la señal en `webkitgtk`, esto se borra: re-chequear con
/// `webkit_captura_test.py` del paquete.
#[cfg(target_os = "linux")]
fn conceder_permiso_de_microfono(win: &tauri::WebviewWindow) {
    use webkit2gtk::WebViewExt;

    // El WebView nativo solo se toca desde el hilo de la UI; `with_webview` garantiza eso.
    if let Err(e) = win.with_webview(|webview| {
        webview
            .inner()
            .connect_permission_request(|_, req| decidir_permiso(req));
    }) {
        // Degradación honesta: sin el hook, el dictado quedará no-disponible. No abortamos el
        // arranque del shell por eso — el resto de la app funciona igual.
        eprintln!("[arnesia] no pude enganchar permission-request del WebView: {e}");
    }
}

/// En macOS/Android lo resuelve `wry`; en el resto no hay superficie que enganchar.
#[cfg(not(target_os = "linux"))]
fn conceder_permiso_de_microfono(_win: &tauri::WebviewWindow) {}

/// decidir_permiso traduce la request nativa a la decisión pura de [`concede_captura`].
///
/// Devuelve `true` = «yo me hice cargo de esta request»; `false` deja correr el manejo por
/// defecto de WebKit (que es negar).
#[cfg(target_os = "linux")]
fn decidir_permiso(req: &webkit2gtk::PermissionRequest) -> bool {
    use webkit2gtk::glib::prelude::Cast;
    use webkit2gtk::{
        PermissionRequestExt, UserMediaPermissionRequest, UserMediaPermissionRequestExt,
    };

    let Some(media) = req.downcast_ref::<UserMediaPermissionRequest>() else {
        // Geolocalización, notificaciones, etc.: ni las miramos.
        return concede_captura(false, false, false);
    };
    if !concede_captura(
        true,
        media.is_for_audio_device(),
        media.is_for_video_device(),
    ) {
        return false;
    }
    req.allow();
    true
}

/// concede_captura es la regla, aislada de GTK para que sea verificable.
///
/// **El shell es la raíz de confianza de la superficie local** (boundary
/// `superficie-local-confinada`): concede EXCLUSIVAMENTE captura de audio. Un allow-all
/// convertiría esa raíz en una llave maestra — de ahí que la cámara se niegue incluso cuando
/// viene en la misma request que el micrófono.
#[cfg(target_os = "linux")]
fn concede_captura(es_user_media: bool, para_audio: bool, para_video: bool) -> bool {
    es_user_media && para_audio && !para_video
}

/// mint_token genera un token aleatorio de 256 bits en hex (raíz de confianza de la API local).
fn mint_token() -> String {
    let mut b = [0u8; 32];
    getrandom::getrandom(&mut b).expect("getrandom: sin fuente de entropía");
    b.iter().map(|x| format!("{x:02x}")).collect()
}

/// Qué encontró el shell en `DAEMON_ADDR` al arrancar (CW-D8).
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum EstadoDaemon {
    /// Nadie escucha: spawneamos nuestro sidecar (camino normal de la app instalada).
    Ausente,
    /// Un daemon que SIRVE la UI: attach (camino normal de dev, con `arnesia serve` a mano).
    ConUI,
    /// Algo ocupa el puerto pero no sirve la UI. Ni attach (mostraría su 404 crudo) ni spawn
    /// (el bind moriría con «address already in use»): se lo decimos al operador.
    SinUI,
}

/// veredicto_daemon es la regla, aislada de la red para que sea verificable — misma doctrina
/// que [`concede_captura`].
///
/// **El bug que cierra (2026-08-14):** el sondeo era `TcpStream::connect` a secas, así que
/// CUALQUIER cosa escuchando en el puerto contaba como «el daemon ya está arriba». Un
/// `arnesia serve` compilado sin la SPA (un `go build` suelto, o el binario Linux corriendo en
/// WSL sobre el mismo repo) responde el puerto Y responde 404 en `/` — y la ventana del shell
/// terminaba mostrando ese 404 en texto plano, sin ninguna pista de qué hacer.
fn veredicto_daemon(conecta: bool, status: Option<u16>) -> EstadoDaemon {
    match (conecta, status) {
        (false, _) => EstadoDaemon::Ausente,
        (true, Some(200)) => EstadoDaemon::ConUI,
        (true, _) => EstadoDaemon::SinUI,
    }
}

/// status_de_respuesta extrae el código de una línea de estado HTTP (`HTTP/1.1 200 OK` → 200).
/// Devuelve `None` si lo que contestó no es HTTP: un ocupante mudo no puede pasar por daemon.
fn status_de_respuesta(linea: &str) -> Option<u16> {
    let mut partes = linea.split_whitespace();
    let version = partes.next()?;
    if !version.starts_with("HTTP/") {
        return None;
    }
    partes.next()?.parse().ok()
}

/// estado_daemon sondea `DAEMON_ADDR` y decide qué hacer (bind-or-bail verificado).
///
/// Pide `GET /` con un HTTP crudo sobre el mismo socket en vez de sumar un cliente HTTP al
/// shell: es una línea de estado, no vale una dependencia. Va acá y NO en el JS de
/// `conectando.html` porque un `fetch` del WebView viaja con el `Origin` de Tauri y chocaría
/// con los gates de `auth.go` — por eso esa página sondea `/healthz`, el único exento
/// (HS-14 fix ②). Todos los timeouts son cortos: el arranque de la app no espera a nadie.
fn estado_daemon() -> EstadoDaemon {
    const ESPERA: Duration = Duration::from_millis(300);

    let Some(addr) = DAEMON_ADDR.parse().ok() else {
        return EstadoDaemon::Ausente;
    };
    let Ok(mut sock) = TcpStream::connect_timeout(&addr, ESPERA) else {
        return veredicto_daemon(false, None);
    };
    let _ = sock.set_read_timeout(Some(ESPERA));
    let _ = sock.set_write_timeout(Some(ESPERA));

    // `Connection: close` para que el servidor no deje el socket abierto esperando otra
    // petición: solo queremos la primera línea.
    let peticion = format!("GET / HTTP/1.1\r\nHost: {DAEMON_ADDR}\r\nConnection: close\r\n\r\n");
    let status = std::io::Write::write_all(&mut sock, peticion.as_bytes())
        .ok()
        .and_then(|()| {
            let mut linea = String::new();
            std::io::BufRead::read_line(&mut std::io::BufReader::new(&sock), &mut linea).ok()?;
            status_de_respuesta(&linea)
        });
    veredicto_daemon(true, status)
}

// Attach verificado (CW-D8): misma doctrina que `concede_captura` — la DECISIÓN se aísla de
// la red para poder testearla. Estos tests SÍ corren en todos los OS (el bug que cierran se
// reportó en Windows) a diferencia de los de permisos, que exigen la superficie GTK.
#[cfg(test)]
mod tests_attach {
    use super::{status_de_respuesta, veredicto_daemon, EstadoDaemon};

    #[test]
    fn puerto_libre_significa_spawnear() {
        assert!(matches!(
            veredicto_daemon(false, None),
            EstadoDaemon::Ausente
        ));
    }

    #[test]
    fn daemon_que_sirve_la_ui_se_attachea() {
        assert!(matches!(
            veredicto_daemon(true, Some(200)),
            EstadoDaemon::ConUI
        ));
    }

    // EL bug (2026-08-14): un `arnesia serve` compilado sin la SPA (go build suelto, WSL)
    // responde 404 en `/`. El attach ciego mostraba ese 404 crudo en la ventana.
    #[test]
    fn ocupante_que_no_sirve_la_ui_no_se_attachea() {
        for status in [404, 401, 403, 500, 302] {
            assert!(
                matches!(veredicto_daemon(true, Some(status)), EstadoDaemon::SinUI),
                "status {status} debe ser SinUI"
            );
        }
    }

    // Algo ocupa el puerto pero no habla HTTP (otro programa cualquiera): tampoco se attachea.
    #[test]
    fn ocupante_mudo_no_se_attachea() {
        assert!(matches!(veredicto_daemon(true, None), EstadoDaemon::SinUI));
    }

    #[test]
    fn lee_el_status_de_la_linea_de_respuesta() {
        assert_eq!(status_de_respuesta("HTTP/1.1 200 OK"), Some(200));
        assert_eq!(status_de_respuesta("HTTP/1.0 404 Not Found"), Some(404));
        assert_eq!(status_de_respuesta("HTTP/1.1 500 "), Some(500));
    }

    #[test]
    fn una_respuesta_que_no_es_http_no_da_status() {
        for basura in [
            "",
            "hola",
            "HTTP/1.1",
            "HTTP/1.1 no-numero OK",
            "220 SMTP listo",
        ] {
            assert_eq!(status_de_respuesta(basura), None, "{basura:?} no es HTTP");
        }
    }
}

// RF-215 · capability arnesia.tauri.permiso-de-microfono. La regla de concesión se testea
// aislada de GTK a propósito: enganchar la señal exige un WebView vivo con display, pero LA
// DECISIÓN —qué se concede y qué no— es la superficie de enforcement, y un enforcement sin
// test es una promesa, no un cerrojo (misma doctrina que los flags de permisos del conductor).
#[cfg(all(test, target_os = "linux"))]
mod tests {
    use super::concede_captura;

    #[test]
    fn concede_solo_captura_de_audio() {
        assert!(concede_captura(true, true, false));
    }

    #[test]
    fn rechaza_permisos_que_no_son_de_medios() {
        // Geolocalización, notificaciones, portapapeles…: no son UserMediaPermissionRequest.
        assert!(!concede_captura(false, false, false));
        // Ni siquiera si algo llegara con las flags de audio puestas.
        assert!(!concede_captura(false, true, false));
    }

    #[test]
    fn rechaza_la_camara_aunque_venga_junto_con_el_microfono() {
        // El caso peligroso: una sola request que pide audio Y video. Conceder «porque pide
        // audio» encendería la cámara del operador sin que nadie la haya pedido.
        assert!(!concede_captura(true, true, true));
        assert!(!concede_captura(true, false, true));
    }

    #[test]
    fn rechaza_media_que_no_pide_audio() {
        assert!(!concede_captura(true, false, false));
    }
}
