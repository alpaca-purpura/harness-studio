// Evita abrir una consola extra en Windows en release. NO REMOVER.
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

fn main() {
    // Mitigación Mint/Linux load-bearing (HS-04): NVIDIA + DMABUF => ventana negra/basura
    // en WebKitGTK 4.1. DEBE setearse ANTES de construir el WebView, por eso va aquí,
    // antes de spawnear cualquier hilo o crear el runtime de Tauri.
    // Edición 2021: `set_var` es segura (en edición 2024 pasa a ser `unsafe`). // verificar al instalar
    #[cfg(target_os = "linux")]
    std::env::set_var("WEBKIT_DISABLE_DMABUF_RENDERER", "1");

    arnesia_lib::run()
}
