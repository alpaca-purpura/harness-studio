import gi
import sys

gi.require_version("Gtk", "3.0")
gi.require_version("WebKit2", "4.1")
from gi.repository import Gtk, WebKit2, GLib

HTML = """
<!doctype html><html><body>
<script>
function report(obj) {
  document.title = "RESULT:" + JSON.stringify(obj);
}
(async () => {
  const out = {
    SpeechRecognition: 'SpeechRecognition' in window,
    webkitSpeechRecognition: 'webkitSpeechRecognition' in window,
    mediaDevices: !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia),
    userAgent: navigator.userAgent,
  };
  if (out.webkitSpeechRecognition || out.SpeechRecognition) {
    try {
      const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
      const r = new SR();
      out.instantiated = true;
      out.lang_default = r.lang || null;
    } catch (e) {
      out.instantiated = false;
      out.instantiate_error = String(e);
    }
  }
  report(out);
})();
</script>
</body></html>
"""

win = Gtk.OffscreenWindow()
webview = WebKit2.WebView()
win.add(webview)
win.set_default_size(800, 600)

result_holder = {}

def on_title(view, pspec):
    title = view.get_title()
    if title and title.startswith("RESULT:"):
        result_holder["json"] = title[len("RESULT:"):]
        GLib.timeout_add(200, lambda: (Gtk.main_quit(), False)[1])

webview.connect("notify::title", on_title)
webview.load_html(HTML, "file:///test/")
win.show_all()

def timeout_bail():
    if "json" not in result_holder:
        result_holder["json"] = "TIMEOUT — no title event received"
    Gtk.main_quit()
    return False

GLib.timeout_add(6000, timeout_bail)
Gtk.main()

print(result_holder.get("json", "NO RESULT"))
