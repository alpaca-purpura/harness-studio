#!/usr/bin/env python3
"""Prueba REAL de captura de audio en WebKitGTK — el motor que corre el shell Tauri de ArnesIA.

Complementa a `webkit_speech_test.py` (que solo pregunta si existe `SpeechRecognition`). Acá la
pregunta es la que importa para el build: **¿la captura FUNCIONA, no solo existe la API?**

Uso:
    DISPLAY=:0 python3 webkit_captura_test.py            # sin handler de permiso → se cuelga (a propósito)
    DISPLAY=:0 python3 webkit_captura_test.py --allow    # concede el permiso → abre el mic real

Lo que respondió esta prueba el 2026-07-25 (WebKitGTK 2.52.3, ver `spike-spec.md` §1.6):

  - `MediaRecorder` existe, pero SOLO soporta `audio/mp4` (webm/opus/ogg/wav = false), y su
    `mimeType` por defecto viene vacío → el FE debe pasar el mime explícito.
  - SIN handler de `permission-request`, `getUserMedia` **no rechaza: se cuelga para siempre**
    (queda en stage `calling-gum`). No hay excepción que catchear.
  - CON handler + ventana VISIBLE: la señal llega como `UserMediaPermissionRequest`, se concede, y
    el track abre contra el device real (`state: live`). En `Gtk.OffscreenWindow` la señal ni se
    emite — WebKit difiere la captura mientras el documento está oculto.
  - `http://127.0.0.1:4200` (el origin real del daemon) es `isSecureContext: true` → sin HTTPS.

Consecuencia para el build: `wry 0.55.1` NO conecta `permission-request` en el backend webkitgtk
(sí en macOS/Android) → hace falta el puente Rust (ticket T0) o el botón de mic queda mudo en la app
instalada aunque funcione en `pnpm dev`.

La prueba ABRE el micrófono y lo cierra en el acto (`track.stop()` en el mismo tick): NO graba audio
ni escribe nada a disco.
"""

import sys

import gi

gi.require_version("Gtk", "3.0")
gi.require_version("WebKit2", "4.1")
from gi.repository import GLib, Gtk, WebKit2  # noqa: E402

HTML = """
<!doctype html><html><body style="font:14px sans-serif;padding:12px">
<h3>ArnesIA spike voz — prueba de captura (se cierra sola)</h3>
<div id=s>...</div>
<script>
const out = {stage:'init'};
function report(){ document.title = "R:" + JSON.stringify(out);
  document.getElementById('s').textContent = out.stage; }

out.visibility = document.visibilityState;
out.secureContext = window.isSecureContext;
out.mediaDevices = !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia);
out.MediaRecorder = typeof window.MediaRecorder !== 'undefined';
out.mimes = {};
if (out.MediaRecorder) {
  for (const m of ['audio/webm','audio/webm;codecs=opus','audio/ogg',
                   'audio/ogg;codecs=opus','audio/mp4','audio/wav']) {
    try { out.mimes[m] = MediaRecorder.isTypeSupported(m); } catch(e) { out.mimes[m] = 'ERR'; }
  }
}
out.stage='sync'; report();

// watchdog: si getUserMedia se cuelga (el bug real), el stage lo delata en vez de mentir con silencio
setTimeout(() => { out.stage = out.stage + '|watchdog-3s'; report(); }, 3000);

(async () => {
  if (!out.mediaDevices) { out.stage='no-gum'; report(); return; }
  out.stage='calling-gum'; report();
  try {
    const s = await navigator.mediaDevices.getUserMedia({audio:true});
    out.getUserMedia='OK';
    out.tracks = s.getAudioTracks().map(t=>({label:t.label, state:t.readyState}));
    if (out.MediaRecorder) {
      try { const r = new MediaRecorder(s); out.recorder_default_mime = r.mimeType; }
      catch(e){ out.recorder_error = String(e); }
    }
    s.getTracks().forEach(t=>t.stop());   // corta el mic YA — no se graba nada
    out.stage='gum-ok-stream-stopped'; report();
  } catch(e){ out.getUserMedia='FAIL: '+String(e); out.stage='gum-fail'; report(); }
})();
</script></body></html>
"""

allow = "--allow" in sys.argv

win = Gtk.Window()  # ventana VISIBLE a propósito: offscreen no emite la permission-request
webview = WebKit2.WebView()
webview.get_settings().set_property("enable-media-stream", True)


def on_perm(view, request):
    kind = type(request).__name__
    print(f"permission-request FIRED: {kind}", file=sys.stderr)
    if allow:
        request.allow()
        return True
    print("  (no se concede — así se reproduce el cuelgue)", file=sys.stderr)
    return False


webview.connect("permission-request", on_perm)
win.add(webview)
win.set_default_size(520, 200)
win.set_title("spike voz — captura")

last = {}


def on_title(view, pspec):
    t = view.get_title()
    if t and t.startswith("R:"):
        last["json"] = t[2:]
        print("TITLE:", last["json"], file=sys.stderr)


webview.connect("notify::title", on_title)
# El origin importa: es el que sirve el daemon Go. Probado literal, no por analogía con localhost.
webview.load_html(HTML, "http://127.0.0.1:4200/")
win.show_all()
win.present()

GLib.timeout_add(8000, lambda: (Gtk.main_quit(), False)[1])
Gtk.main()
print("FINAL:", last.get("json", "NO RESULT"))
