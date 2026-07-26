#!/usr/bin/env python3
"""¿El motor GRABA, o solo dice que puede? — sonda de la ronda 3 (2026-07-26).

Complementa a `webkit_captura_test.py`, que resultó insuficiente: aquel prueba que
`MediaRecorder` **existe** y qué mimes **dice** soportar. Eso pasó en verde mientras el dictado
estaba roto en la app instalada durante toda la v0.2.20. **Presencia de API ≠ bytes grabados** —
la misma lección que V-D8, una vuelta más abajo.

Lo que esta sonda mide de verdad, en tres caminos:

  A) MediaRecorder audio/mp4 sin timeslice   — lo que hacía v0.2.20
  B) MediaRecorder audio/mp4 con timeslice   — ¿lo arregla forzar entregas parciales?
  C) WebAudio (ScriptProcessor) → PCM crudo   — la salida que se adoptó (RF-229)

Resultado del 2026-07-26 (WebKitGTK 2.52.3, `http://127.0.0.1:4200`):

  A → onstart ✅ · state "recording" ✅ · 1 evento de 0 bytes · blob 0 bytes · onerror NUNCA
  B → idéntico: 0 bytes
  C → 37 bloques · 151552 muestras · 3.44 s · pico 0.0787 ✅

Causa raíz, visible con `GST_DEBUG="*:2,webkit*:6"`:

  MediaRecorderPrivateGStreamer.cpp:412: Setting audio restriction caps to audio/x-raw, rate=(int)0
  gstencodebasebin.c:718: <encodebin2-0> Couldn't find a compatible stream profile
  gstbasesrc.c:3177: <capture-audiosrc0> error: streaming stopped, reason not-linked (-1)
  MediaRecorderPrivateGStreamer.cpp:272: Transfering 0 encoded bytes

WebKit arma el perfil con `rate=0`, `encodebin` no casa ningún profile, el pad de audio nunca se
linkea y el micrófono empuja a la nada. El pipeline SÍ postea `error` en el bus de GStreamer, pero
WebKit no lo propaga a JS — por eso `onerror` no dispara y el fallo es mudo.

Uso:

    DISPLAY=:0 python3 webkit_grabacion_test.py
    DISPLAY=:0 GST_DEBUG="*:2,webkit*:6" python3 webkit_grabacion_test.py 2> gst.log

**Re-correr tras un upgrade de WebKitGTK.** Si A vuelve con bytes > 0, el bug del motor se arregló
y RF-229 puede revisarse (no antes: la conducta del motor es el único dato que manda).

GRABA el micrófono unos segundos por camino. El audio vive en memoria y no se escribe a disco.
"""

import sys

import gi

gi.require_version("Gtk", "3.0")
gi.require_version("WebKit2", "4.1")
from gi.repository import GLib, Gtk, WebKit2  # noqa: E402

SEGUNDOS = 4

HTML = """
<!doctype html><html><body style="font:13px monospace;padding:12px">
<h3>ArnesIA — ¿el motor graba? (A/B: MediaRecorder · C: WebAudio)</h3>
<p>Hablá mientras corre. Se cierra sola.</p><pre id=s>...</pre>
<script>
const SEG = __SEG__ * 1000;
const out = {caminos: []};
function report(){ document.title = "R:" + JSON.stringify(out);
  document.getElementById('s').textContent = JSON.stringify(out, null, 1); }

// A/B — MediaRecorder. `timeslice` null = lo que hacía v0.2.20.
function conRecorder(nombre, timeslice, opciones) {
  return new Promise(async (resolve) => {
    const r = {camino: nombre, timeslice: timeslice ?? "ninguno"};
    try {
      r.isTypeSupported = MediaRecorder.isTypeSupported('audio/mp4');
      const stream = await navigator.mediaDevices.getUserMedia({audio:true});
      r.tracks = stream.getAudioTracks().map(t=>({label:t.label, state:t.readyState, muted:t.muted}));
      const rec = new MediaRecorder(stream, {mimeType:'audio/mp4', ...(opciones||{})});
      r.mimeType_efectivo = rec.mimeType;
      const trozos = [];
      rec.ondataavailable = (e) => { r.eventos_data = (r.eventos_data||0)+1;
        r.tam_por_evento = (r.tam_por_evento||[]).concat(e.data.size);
        if (e.data.size > 0) trozos.push(e.data); };
      // onerror cableado A PROPÓSITO: el hallazgo es que NUNCA dispara pese al error del bus.
      rec.onerror = (e) => { r.onerror = String(e.error || e.name || e); };
      rec.onstart = () => { r.onstart = true; };
      rec.onstop = () => {
        r.blob_bytes = new Blob(trozos, {type:'audio/mp4'}).size;
        stream.getTracks().forEach(t=>t.stop());
        out.caminos.push(r); report(); resolve();
      };
      if (timeslice) rec.start(timeslice); else rec.start();
      r.state_tras_start = rec.state;
      setTimeout(() => { try { rec.stop(); } catch(e){ r.stop_error = String(e);
        out.caminos.push(r); report(); resolve(); } }, SEG);
    } catch(e) { r.error = String(e); out.caminos.push(r); report(); resolve(); }
  });
}

// C — WebAudio: PCM crudo sin pasar por el encodebin de GStreamer.
async function conWebAudio() {
  const r = {camino: "WebAudio (ScriptProcessor) — RF-229"};
  try {
    const stream = await navigator.mediaDevices.getUserMedia({audio:true});
    const ctx = new AudioContext();
    await ctx.resume().catch(()=>{});
    r.sampleRate = ctx.sampleRate; r.estado_ctx = ctx.state;
    const src = ctx.createMediaStreamSource(stream);
    const proc = ctx.createScriptProcessor(4096, 1, 1);
    let muestras = 0, pico = 0, bloques = 0;
    proc.onaudioprocess = (e) => {
      const d = e.inputBuffer.getChannelData(0);
      bloques++; muestras += d.length;
      for (let i=0;i<d.length;i++) { const a = Math.abs(d[i]); if (a>pico) pico=a; }
    };
    src.connect(proc); proc.connect(ctx.destination);
    await new Promise(res => setTimeout(res, SEG));
    proc.disconnect(); src.disconnect(); await ctx.close();
    stream.getTracks().forEach(t=>t.stop());
    r.bloques = bloques; r.muestras = muestras; r.pico = Number(pico.toFixed(4));
    r.segundos = Number((muestras / (r.sampleRate||1)).toFixed(2));
    // Lo que pesaría el WAV 16 kHz que sube la app (16000 muestras/s × 2 bytes + header).
    r.wav16k_bytes = Math.floor(muestras / (r.sampleRate/16000)) * 2 + 44;
  } catch(e) { r.error = String(e); }
  out.caminos.push(r); report();
}

(async () => {
  out.etapa='A'; report(); await conRecorder("MediaRecorder sin timeslice (v0.2.20)", null, null);
  out.etapa='B'; report(); await conRecorder("MediaRecorder timeslice 250ms", 250, null);
  out.etapa='C'; report(); await conWebAudio();
  out.etapa='listo';
  out.veredicto = out.caminos.map(c => `${c.camino}: ${c.blob_bytes ?? c.wav16k_bytes ?? "?"} bytes`);
  report();
})();
</script></body></html>
""".replace("__SEG__", str(SEGUNDOS))

win = Gtk.Window()
webview = WebKit2.WebView()
webview.get_settings().set_property("enable-media-stream", True)


def on_perm(_view, request):
    # El shell real hace lo mismo desde Rust (RF-215): sin esto getUserMedia se cuelga mudo.
    print(f"permission-request: {type(request).__name__} → allow", file=sys.stderr)
    request.allow()
    return True


webview.connect("permission-request", on_perm)
win.add(webview)
win.set_default_size(760, 520)
win.set_title("ArnesIA — sonda de grabación")

ultimo = {}


def on_title(view, _pspec):
    t = view.get_title()
    if t and t.startswith("R:"):
        ultimo["json"] = t[2:]


webview.connect("notify::title", on_title)
# El origin importa: es el que sirve el daemon. Probado literal, no por analogía con localhost.
webview.load_html(HTML, "http://127.0.0.1:4200/")
win.show_all()
win.present()


def fin():
    print("FINAL:", ultimo.get("json", "(la página no reportó)"))
    Gtk.main_quit()
    return False


# 3 caminos × SEGUNDOS + margen para getUserMedia y el arranque del pipeline.
GLib.timeout_add((SEGUNDOS * 3 + 10) * 1000, fin)
Gtk.main()
