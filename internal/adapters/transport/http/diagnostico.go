package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// maxDiagnostico caps the diagnostic body. 64 KB entra un stack largo con holgura y le pone
// techo a lo que un POST puede hacerle crecer al log del disco del operador.
const maxDiagnostico = 64 << 10

// maxClave acota `origen`/`evento`: son claves para grepear el log, no prosa. Sin tope, un
// evento de 60 KB haría ilegible el archivo que este endpoint existe para hacer legible.
const maxClave = 120

// eventoDiagnostico is one FE failure, told with enough detail to fix it without reproducing.
//
// `Detalle` es libre (`map[string]any`) a propósito: cada fallo tiene sus propias variables
// —un dictado necesita muestras/pico/hz, un crash de render necesita el stack— y encorsetarlas
// en un schema fijo haría que la próxima falla desconocida no tuviera dónde contarse. Se
// loguea tal cual; **nadie parsea su forma como contrato** (mismo criterio que
// `conductor-no-parsea-jsonl.md`).
type eventoDiagnostico struct {
	Origen  string         `json:"origen"`
	Evento  string         `json:"evento"`
	Mensaje string         `json:"mensaje,omitempty"`
	Detalle map[string]any `json:"detalle,omitempty"`
}

// postDiagnostico (RF-230) recibe un fallo del FE y lo escribe al log del daemon.
//
// **Por qué el FE necesita esto.** La app instalada corre en un WebView sin devtools: su
// `console.error` no va a ningún lado. El daemon es el único proceso de la app que escribe a
// disco, así que es el único que puede dejar rastro de un fallo del navegador. El caso que lo
// motivó: `MediaRecorder` de WebKitGTK entregó 0 bytes durante toda la v0.2.20 sin emitir
// `error`, y del incidente solo quedaba la frase que el operador leyó en pantalla.
//
// Responde 204: el FE ya le mostró el fallo real al operador y no tiene nada que hacer con
// esta respuesta — que este POST falle no puede convertirse en un segundo error.
func postDiagnostico() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxDiagnostico)
		var ev eventoDiagnostico
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "diagnóstico ilegible: " + err.Error()})
			return
		}
		ev.Origen = recortarClave(ev.Origen)
		ev.Evento = recortarClave(ev.Evento)
		if ev.Evento == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "el diagnóstico no dice qué evento es"})
			return
		}
		if ev.Origen == "" {
			ev.Origen = "desconocido"
		}

		// Error y no Warn: lo que llega acá es un fallo que el operador ya vio en la
		// superficie. Bajarlo de nivel lo escondería del grep que se hace justo después.
		slog.Error("diagnostico",
			"origen", ev.Origen,
			"evento", ev.Evento,
			"mensaje", ev.Mensaje,
			"detalle", ev.Detalle,
		)
		w.WriteHeader(http.StatusNoContent)
	}
}

// recortarClave trims and caps a key field.
func recortarClave(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > maxClave {
		return s[:maxClave]
	}
	return s
}
